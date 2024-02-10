package secgrp

import (
	"fmt"
	"time"

	"github.com/steffakasid/awsclean/internal"
)

type SecGrp struct {
	awsClient     *internal.AWS
	olderthen     *time.Duration
	createdAgo    *time.Duration
	dryrun        bool
	onlyUnused    bool
	showTags      bool
	usedSecGrps   internal.SecurityGroups
	unusedSecGrps internal.SecurityGroups
}

func NewInstance(awsClient *internal.AWS, olderthen, createdAgo *time.Duration, dryrun, onlyUnused, showTags bool) *SecGrp {
	return &SecGrp{
		awsClient:     awsClient,
		olderthen:     olderthen,
		createdAgo:    createdAgo,
		dryrun:        dryrun,
		onlyUnused:    onlyUnused,
		showTags:      showTags,
		usedSecGrps:   internal.SecurityGroups{},
		unusedSecGrps: internal.SecurityGroups{},
	}
}

func (sec *SecGrp) GetSecurityGroups(startTime, endTime time.Time) error {

	secGrpsFromCCTrail := sec.awsClient.GetCloudTrailForSecGroups(startTime, endTime)
	// TODO: we might not always get Details from CloudTrail or at least we need multiple describe SecGrp calls
	secGrps, err := sec.awsClient.GetSecurityGroups(secGrpsFromCCTrail)
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
					internal.Logger.Errorf("error deleting security group: %s", err)
				}
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
