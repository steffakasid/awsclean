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

func (sec SecGrp) GetSecurityGroups() (internal.SecurityGroups, error) {

	secGrpsFromCCTrail := internal.SecurityGroups{}
	if nil != sec.createdAgo {

		endtime := time.Now()
		starttime := endtime.Add(*sec.createdAgo * -1)

		secGrpsFromCCTrail = sec.awsClient.GetCloudTrailForSecGroups(&starttime, &endtime)
	}
	secGrps, err := sec.awsClient.GetSecurityGroups(sec.dryrun, secGrpsFromCCTrail)

	if nil != err {
		return nil, fmt.Errorf("could not get SecurityGroups: %w", err)
	}

	if sec.onlyUnused || sec.olderthen != nil {
		notUsed, err := sec.awsClient.GetNotUsedSecGrpsFromENI(secGrps, sec.dryrun)
		if nil != err {
			return nil, fmt.Errorf("could not get not used SecurityGroups from ENIs: %w", err)
		}
		return notUsed, nil
	}
	return secGrps, nil
}

func (sec SecGrp) DeleteUnusedSecurityGroups() error {
	secGrpIDs, err := sec.awsClient.GetSecurityGroups(sec.dryrun, internal.SecurityGroups{})
	if nil != err {
		return fmt.Errorf("could not get SecurityGroups: %w", err)
	}
	notUsed, err := sec.awsClient.GetNotUsedSecGrpsFromENI(secGrpIDs, sec.dryrun)
	if nil != err {
		return fmt.Errorf("could not get not used SecurityGroups from ENIs: %w", err)
	}
	if !sec.dryrun {
		for _, secGrp := range notUsed {

			if secGrp.CreationTime == nil || secGrp.CreationTime.Before(time.Now().Add(*sec.olderthen*-1)) {
				err := sec.awsClient.DeleteSecurityGroup(secGrp, sec.dryrun)
				logger.Errorf("error deleting security group: %s", err)
			}
		}
	}
	return nil
}
