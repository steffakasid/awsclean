package amiclean

import (
	"time"

	logger "github.com/sirupsen/logrus"
	"github.com/steffakasid/awsclean/internal"
)

type AmiClean struct {
	awsClient      *internal.AWS
	olderthen      time.Duration
	awsaccount     string
	dryrun         bool
	useLaunchTpls  bool
	usedAMIs       []string
	ignorePatterns []string
}

func NewInstance(
	awsClient *internal.AWS,
	olderthen time.Duration, awsaccount string,
	dryrun bool,
	useLaunchTpls bool,
	ignorePatterns []string) *AmiClean {

	return &AmiClean{
		awsClient:      awsClient,
		olderthen:      olderthen,
		awsaccount:     awsaccount,
		dryrun:         dryrun,
		useLaunchTpls:  useLaunchTpls,
		usedAMIs:       []string{},
		ignorePatterns: ignorePatterns,
	}
}

func (a *AmiClean) GetUsedAMIs() {
	a.usedAMIs = append(a.usedAMIs, a.awsClient.GetUsedAMIsFromEC2()...)

	if a.useLaunchTpls {
		a.usedAMIs = append(a.usedAMIs, a.awsClient.GetUsedAMIsFromLaunchTpls()...)
	}
}

func (a AmiClean) DeleteOlderUnusedAMIs() error {

	images, err := a.awsClient.DescribeImages(a.awsaccount)

	if err != nil {
		return err
	}
	today := time.Now()
	olderThenDate := today.Add(a.olderthen * -1)
	for _, image := range images {
		if !internal.Contains(a.usedAMIs, *image.ImageId) {
			ok, err := internal.MatchAny(*image.Name, a.ignorePatterns)
			if err != nil {
				return err
			}
			if !ok {
				creationDate, err := time.Parse("2006-01-02T15:04:05.000Z", *image.CreationDate)
				if err != nil {
					logger.Error(err)
				}
				if creationDate.Before(olderThenDate) {
					logger.Infof("Delete %s:%s as it's creationdate %s is older then %s", *image.ImageId, *image.Name, *image.CreationDate, olderThenDate.String())
					err := a.awsClient.DeregisterImage(*image.ImageId, a.dryrun)
					internal.CheckError(err, logger.Errorf)
				} else {
					logger.Infof("Keeping %s:%s as it's creationdate %s is newer then %s", *image.ImageId, *image.Name, *image.CreationDate, olderThenDate.String())
				}
			} else {
				logger.Infof("Ignored %s", *image.ImageId)
			}
		} else {
			logger.Infof("Skipping %s", *image.ImageId)
		}
	}
	return nil
}
