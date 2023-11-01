package secgrp

import (
	"fmt"
	"maps"
	"time"

	"github.com/steffakasid/awsclean/internal"
	eslog "github.com/steffakasid/eslog"
)

type SecGrp struct {
	awsClient     *internal.AWS
	olderthen     *time.Duration
	dryrun        bool
	onlyUnused    bool
	usedSecGrps   *internal.SecurityGroups
	unusedSecGrps *internal.SecurityGroups
}

func NewInstance(awsClient *internal.AWS, olderthen *time.Duration, dryrun, onlyUnused bool) *SecGrp {
	return &SecGrp{
		awsClient:     awsClient,
		olderthen:     olderthen,
		dryrun:        dryrun,
		onlyUnused:    onlyUnused,
		usedSecGrps:   &internal.SecurityGroups{},
		unusedSecGrps: &internal.SecurityGroups{},
	}
}

func (sec *SecGrp) GetSecurityGroups(startTime, endTime time.Time) error {
	ninetyDayOffset := internal.ParseDuration("90d")
	secGrps := internal.SecurityGroups{}
	var err error

	eslog.Logger.Debug("GetCloudTrailForSecGroups")
	secGrpsFromCCTrail := sec.awsClient.GetCloudTrailForSecGroups(startTime, endTime)

	// if startTime is before 90d in past we want to get additional SecurityGroups which are not in CloudTrail
	eslog.Logger.Debug("GetSecurityGroups")
	result, err := sec.awsClient.GetSecurityGroups()
	if nil != err {
		return fmt.Errorf("could not getSecurityGroups: %w", err)
	}
	secGrps.AppendAll(result)

	skippedSecGrps := secGrps.UpdateIfExists(secGrpsFromCCTrail)
	eslog.Logger.Debugf("After additionalDetails len(secGrps) %d", len(secGrps))

	if startTime.After(time.Now().Add(ninetyDayOffset * -1)) {
		eslog.Logger.Debugf("To be deleted len(skippedSecGrps) %d", len(skippedSecGrps))
		secGrps.DeleteSkipped(skippedSecGrps)
		eslog.Logger.Debugf("After delete skipped len(secGrps) %d", len(secGrps))
	}

	if sec.onlyUnused || sec.olderthen != nil {
		eslog.Logger.Debug("GetNotUsedSecGrpsFromENI")
		sec.usedSecGrps, sec.unusedSecGrps, err = sec.awsClient.GetNotUsedSecGrpsFromENI(secGrps)
		eslog.Logger.Debugf("GetNotUsedSecGrpsFromENI() len(secGrps) %d", len(secGrps))
		if err != nil {
			return fmt.Errorf("could not get GetNotUsedSecGrpsFromENI() %w", err)
		}
	}

	eslog.Logger.Debug("secgrp.go GetSecurityGroups returning no error")
	return nil
}

func (sec SecGrp) DeleteSecurityGroups(startTime, endTime time.Time) error {
	err := sec.GetSecurityGroups(startTime, endTime)
	if err != nil {
		return err
	}

	if sec.onlyUnused {
		for _, secGrp := range *sec.unusedSecGrps {
			if secGrp.CreationTime == nil ||
				sec.olderthen == nil ||
				(sec.olderthen != nil && secGrp.CreationTime.Before(time.Now().Add(*sec.olderthen*-1))) {
				if sec.olderthen == nil {
					eslog.Logger.Info("olderthen not set ignoring CreationTime of SecurityGroup")
				}
				err := sec.awsClient.DeleteSecurityGroup(*secGrp, sec.dryrun)
				if err != nil {
					eslog.LogIfErrorf(err, eslog.Errorf, "error deleting security group: %s")
				}
			} else {
				eslog.Logger.Infof("Skipping because of CreationDate %s - %s: %s", *secGrp.SecurityGroup.GroupName, *secGrp.SecurityGroup.GroupId, secGrp.CreationTime.Format(time.RFC3339))
			}
		}
	} else {
		// I guess trying to delete used SwcurityGroups will not work. So Idid not implement it
	}
	return nil
}

func (sec SecGrp) GetAllSecurityGroups() internal.SecurityGroups {
	all := internal.SecurityGroups{}

	eslog.Logger.Debugf("GetAllSecurityGroups() len(all) %d", len(*sec.unusedSecGrps))
	eslog.Logger.Debug("GetAllSecurityGroups append unused")
	maps.Copy(all, *sec.unusedSecGrps)

	if !sec.onlyUnused {
		eslog.Logger.Debugf("GetAllSecurityGroups() len(all) %d", len(*sec.usedSecGrps))
		eslog.Logger.Debug("GetAllSecurityGroups append used")
		maps.Copy(all, *sec.usedSecGrps)
	}
	eslog.Logger.Debugf("GetAllSecurityGroups() len(all) %d", len(all))
	return all
}
