package secgrp

import (
	"fmt"
	"time"

	logger "github.com/sirupsen/logrus"
	"github.com/steffakasid/awsclean/internal"
)

type SecGrp struct {
	awsClient *internal.AWS
	olderthen *time.Duration
	dryrun    bool
	showTags  bool
}

func NewInstance(awsClient *internal.AWS, olderthen *time.Duration, dryrun bool, showTags bool) *SecGrp {
	return &SecGrp{
		awsClient: awsClient,
		olderthen: olderthen,
		dryrun:    dryrun,
		showTags:  showTags,
	}
}

func (sec SecGrp) GetSecurityGroups(unused bool) (internal.SecurityGroups, error) {
	secGrpIDs, err := sec.awsClient.GetSecurityGroups(sec.dryrun)
	if nil != err {
		return nil, fmt.Errorf("could not get SecurityGroups: %w", err)
	}

	if unused || sec.olderthen != nil {
		notUsed, err := sec.awsClient.GetNotUsedSecGrpsFromENI(secGrpIDs, sec.dryrun)
		if nil != err {
			return nil, fmt.Errorf("could not get not used SecurityGroups from ENIs: %w", err)
		}
		return notUsed, nil
	}
	return secGrpIDs, nil
}

func (sec SecGrp) DeleteUnusedSecurityGroups() error {
	secGrpIDs, err := sec.awsClient.GetSecurityGroups(sec.dryrun)
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
