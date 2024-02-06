package secgrp

import (
	"fmt"
	"time"

	logger "github.com/sirupsen/logrus"
	"github.com/steffakasid/awsclean/internal"
)

type SecGrp struct {
	awsClient  *internal.AWS
	olderthen  *time.Duration
	createdAgo *time.Duration
	dryrun     bool
	onlyUnused bool
	showTags   bool
	endTime    *time.Time
}

func NewInstance(awsClient *internal.AWS, olderthen, createdAgo *time.Duration, dryrun, onlyUnused, showTags bool) *SecGrp {
	now := time.Now()
	return &SecGrp{
		awsClient:  awsClient,
		olderthen:  olderthen,
		createdAgo: createdAgo,
		dryrun:     dryrun,
		onlyUnused: onlyUnused,
		showTags:   showTags,
		endTime:    &now,
	}
}

func (sec SecGrp) GetSecurityGroups() (internal.SecurityGroups, error) {

	var starttime time.Time

	if nil != sec.createdAgo {
		starttime = sec.endTime.Add(*sec.createdAgo * -1)
	}
	secGrpsFromCCTrail := sec.awsClient.GetCloudTrailForSecGroups(&starttime, sec.endTime)
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

func (sec SecGrp) DeleteSecurityGroups() error {
	secGrps, err := sec.GetSecurityGroups()
	if err != nil {
		return err
	}

	for _, secGrp := range secGrps {

		if secGrp.CreationTime == nil ||
			(sec.olderthen != nil && secGrp.CreationTime.Before(time.Now().Add(*sec.olderthen*-1))) {
			if (!secGrp.IsUsed && sec.onlyUnused) || !sec.onlyUnused {
				err := sec.awsClient.DeleteSecurityGroup(secGrp, sec.dryrun)
				if err != nil {
					logger.Errorf("error deleting security group: %s", err)
				}
			}
		}
	}

	return nil
}
