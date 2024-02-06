package secgrp

import (
	"fmt"
	"time"

	"github.com/steffakasid/awsclean/internal"
)

type SecGrp struct {
	awsClient  *internal.AWS
	olderthen  *time.Duration
	createdAgo *time.Duration
	dryrun     bool
	onlyUnused bool
	showTags   bool
}

func NewInstance(awsClient *internal.AWS, olderthen, createdAgo *time.Duration, dryrun, onlyUnused, showTags bool) *SecGrp {
	return &SecGrp{
		awsClient:  awsClient,
		olderthen:  olderthen,
		createdAgo: createdAgo,
		dryrun:     dryrun,
		onlyUnused: onlyUnused,
		showTags:   showTags,
	}
}

func (sec SecGrp) GetSecurityGroups(startTime, endTime time.Time) (internal.SecurityGroups, error) {

	secGrpsFromCCTrail := sec.awsClient.GetCloudTrailForSecGroups(startTime, endTime)
	secGrps, err := sec.awsClient.GetSecurityGroups(sec.dryrun, secGrpsFromCCTrail)
	internal.AppendAll(secGrpsFromCCTrail, secGrps)

	if nil != err {
		return nil, fmt.Errorf("could not get SecurityGroups: %w", err)
	}

	if sec.onlyUnused || sec.olderthen != nil {
		notUsed, err := sec.awsClient.GetNotUsedSecGrpsFromENI(secGrps, sec.dryrun)
		internal.AppendAll(notUsed, secGrps)
		if nil != err {
			return nil, fmt.Errorf("could not get not used SecurityGroups from ENIs: %w", err)
		}
		return notUsed, nil
	}
	return secGrps, nil
}

func (sec SecGrp) DeleteSecurityGroups(startTime, endTime time.Time) error {
	secGrps, err := sec.GetSecurityGroups(startTime, endTime)
	if err != nil {
		return err
	}

	for _, secGrp := range secGrps {

		if secGrp.CreationTime == nil ||
			(sec.olderthen != nil && secGrp.CreationTime.Before(time.Now().Add(*sec.olderthen*-1))) {
			if (!secGrp.IsUsed && sec.onlyUnused) || !sec.onlyUnused {
				err := sec.awsClient.DeleteSecurityGroup(secGrp, sec.dryrun)
				if err != nil {
					internal.Logger.Errorf("error deleting security group: %s", err)
				}
			}
		}
	}

	return nil
}
