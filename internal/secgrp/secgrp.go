package secgrp

import (
	"fmt"
	"time"

	"github.com/steffakasid/awsclean/internal"
	extendedslog "github.com/steffakasid/extended-slog"
	"github.com/xhit/go-str2duration/v2"
)

type SecGrp struct {
	awsClient     *internal.AWS
	olderthen     *time.Duration
	dryrun        bool
	onlyUnused    bool
	usedSecGrps   internal.SecurityGroups
	unusedSecGrps internal.SecurityGroups
}

func NewInstance(awsClient *internal.AWS, olderthen *time.Duration, dryrun, onlyUnused bool) *SecGrp {
	return &SecGrp{
		awsClient:     awsClient,
		olderthen:     olderthen,
		dryrun:        dryrun,
		onlyUnused:    onlyUnused,
		usedSecGrps:   internal.SecurityGroups{},
		unusedSecGrps: internal.SecurityGroups{},
	}
}

func (sec *SecGrp) GetSecurityGroups(startTime, endTime time.Time) error {
	ninetyDayOffset, _ := str2duration.ParseDuration("90d")

	secGrpsFromCCTrail := sec.awsClient.GetCloudTrailForSecGroups(startTime, endTime)

	// if startTime is before 90d in past we want to get additional SecurityGroups which are not in CloudTrail
	filterSecGrps := internal.SecurityGroups{}
	if startTime.After(time.Now().Add(ninetyDayOffset * -1)) {
		filterSecGrps = secGrpsFromCCTrail
	}

	secGrps, err := sec.awsClient.GetSecurityGroups(filterSecGrps)
	internal.AppendAll(secGrpsFromCCTrail, secGrps)

	if nil != err {
		return fmt.Errorf("could not get SecurityGroups: %w", err)
	}

	if sec.onlyUnused || sec.olderthen != nil {
		sec.usedSecGrps, sec.unusedSecGrps, err = sec.awsClient.GetNotUsedSecGrpsFromENI(secGrps)
		if nil != err {
			return fmt.Errorf("could not get not used SecurityGroups from ENIs: %w", err)
		}
		return nil
	}
	return nil
}

func (sec SecGrp) DeleteSecurityGroups(startTime, endTime time.Time) error {
	err := sec.GetSecurityGroups(startTime, endTime)
	if err != nil {
		return err
	}

	if sec.onlyUnused {
		for _, secGrp := range sec.unusedSecGrps {
			if secGrp.CreationTime == nil ||
				(sec.olderthen != nil && secGrp.CreationTime.Before(time.Now().Add(*sec.olderthen*-1))) {
				err := sec.awsClient.DeleteSecurityGroup(secGrp, sec.dryrun)
				if err != nil {
					extendedslog.Logger.Errorf("error deleting security group: %s", err)
				}
			} else {
				extendedslog.Logger.Infof("Skipping because of CreationDate %s - %s: %s", *secGrp.SecurityGroup.GroupName, *secGrp.SecurityGroup.GroupId, secGrp.CreationTime.Format(time.RFC3339))
			}
		}
	} else {
		// Iguess trying to delete used SwcurityGroups will not work. So Idid not implement it
	}
	return nil
}

func (sec SecGrp) GetAllSecurityGroups() internal.SecurityGroups {
	all := internal.SecurityGroups{}

	internal.AppendAll(sec.unusedSecGrps, all)

	if !sec.onlyUnused {
		internal.AppendAll(sec.usedSecGrps, all)
	}

	return all
}
