package internal

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	logger "github.com/sirupsen/logrus"
)

type AmiClean struct {
	awsClient      *AWS
	olderthen      time.Duration
	awsaccount     string
	dryrun         bool
	useLaunchTpls  bool
	usedAMIs       []string
	ignorePatterns []string
}

func NewInstance(
	awsClient *AWS,
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
	a.usedAMIs = append(a.usedAMIs, a.awsClient.getUsedAMIsFromEC2()...)

	if a.useLaunchTpls {
		a.usedAMIs = append(a.usedAMIs, a.awsClient.getUsedAMIsFromLaunchTpls()...)
	}
}

func (a AmiClean) DeleteOlderUnusedAMIs() error {
	describeImageInput := &ec2.DescribeImagesInput{Owners: []string{"self"}}
	if a.awsaccount != "" {
		describeImageInput.Owners = append(describeImageInput.Owners, a.awsaccount)
	}
	images, err := a.awsClient.ec2.DescribeImages(context.TODO(), describeImageInput)
	if err != nil {
		return err
	}
	today := time.Now()
	olderThenDate := today.Add(a.olderthen * -1)
	for _, image := range images.Images {
		if !contains(a.usedAMIs, *image.ImageId) {
			ok, err := matchAny(*image.Name, a.ignorePatterns)
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
					deregisterInput := &ec2.DeregisterImageInput{
						ImageId: image.ImageId,
						DryRun:  aws.Bool(a.dryrun),
					}
					_, err := a.awsClient.ec2.DeregisterImage(context.TODO(), deregisterInput)
					CheckError(err, logger.Errorf)
				} else {
					logger.Infof("Keeping %s:%s as it's creationdate %s is newer then %s", *image.ImageId, *image.Name, *image.CreationDate, olderThenDate.String())
				}
			} else {
				logger.Infof("Ignored %s\n", *image.ImageId)
			}
		} else {
			logger.Infof("Skipping %s\n", *image.ImageId)
		}
	}
	return nil
}
