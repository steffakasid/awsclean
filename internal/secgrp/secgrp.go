package secgrp

import (
	"time"

	logger "github.com/sirupsen/logrus"
	"github.com/steffakasid/awsclean/internal"
)

type SecGrp struct {
	awsClient *internal.AWS
	olderthen time.Duration
	dryrun    bool
	showTags  bool
}

func NewInstance(awsClient *internal.AWS, olderthen time.Duration, dryrun bool, showTags bool) *SecGrp {
	return &SecGrp{
		awsClient: awsClient,
		olderthen: olderthen,
		dryrun:    dryrun,
		showTags:  showTags,
	}
}

func (sec *SecGrp) DeleteUnusedSecurityGroups() {
	secGrpIDs := sec.awsClient.GetSecurityGroups()
	notUsed := sec.awsClient.GetNotUsedSecGrpsFromENI(secGrpIDs)
	if !sec.dryrun {
		for _, secGrpID := range notUsed {
			err := sec.awsClient.DeleteSecurityGroup(secGrpID)
			logger.Errorf("error deleting security group: %s", err)
		}
	}
}
