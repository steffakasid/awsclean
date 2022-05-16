package internal

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	logger "github.com/sirupsen/logrus"
)

const awsEC2Volume = "AWS::EC2::Volume"

type Ec2client interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	DeregisterImage(ctx context.Context, params *ec2.DeregisterImageInput, optFns ...func(*ec2.Options)) (*ec2.DeregisterImageOutput, error)
	DescribeLaunchTemplateVersions(ctx context.Context, params *ec2.DescribeLaunchTemplateVersionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeLaunchTemplateVersionsOutput, error)
	DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error)
	DeleteVolume(ctx context.Context, params *ec2.DeleteVolumeInput, optFns ...func(*ec2.Options)) (*ec2.DeleteVolumeOutput, error)
}

type AWS struct {
	ec2 Ec2client
}

func NewFromInterface(ec2 Ec2client) *AWS {
	return &AWS{
		ec2: ec2,
	}
}

func NewAWSClient(conf func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (cfg aws.Config, err error),
	ec2InitFunc func(cfg aws.Config, optFns ...func(*ec2.Options)) *ec2.Client,
	clouttrailInitFunc func(cfg aws.Config, optFns ...func(*cloudtrail.Options)) *cloudtrail.Client) *AWS {
	aws := &AWS{}

	cfg, err := conf(context.TODO())
	CheckError(err, logger.Fatalf)

	aws.ec2 = ec2InitFunc(cfg)
	return aws
}

func (a *AWS) GetUsedAMIsFromEC2() []string {
	usedImages := []string{}
	nextToken := ""
	for {
		opts := &ec2.DescribeInstancesInput{}
		if nextToken != "" {
			opts.NextToken = &nextToken
		}
		ec2Instances, err := a.ec2.DescribeInstances(context.TODO(), opts)
		CheckError(err, logger.Errorf)
		if ec2Instances != nil {
			for _, reserveration := range ec2Instances.Reservations {
				for _, instance := range reserveration.Instances {
					usedImages = UniqueAppend(usedImages, *instance.ImageId)
				}
			}
		}

		if ec2Instances == nil || ec2Instances.NextToken == nil {
			break
		}
		nextToken = *ec2Instances.NextToken
	}
	logger.Debug("UsedImages[] from EC2", usedImages)
	return usedImages
}

func (a *AWS) GetUsedAMIsFromLaunchTpls() []string {
	usedImages := []string{}
	nextToken := ""
	for {
		opts := &ec2.DescribeLaunchTemplateVersionsInput{
			Versions: []string{"$Latest"},
		}
		if nextToken != "" {
			opts.NextToken = &nextToken
		}
		launchTpls, err := a.ec2.DescribeLaunchTemplateVersions(context.TODO(), opts)
		CheckError(err, logger.Errorf)
		if launchTpls != nil {
			for _, launchTplVersion := range launchTpls.LaunchTemplateVersions {
				if launchTplVersion.LaunchTemplateData.ImageId != nil {
					usedImages = append(usedImages, *launchTplVersion.LaunchTemplateData.ImageId)
				}
			}
		}
		if launchTpls == nil || launchTpls.NextToken == nil {
			break
		}
		nextToken = *launchTpls.NextToken
	}
	logger.Debug("UsedImages[] from Launch Templates", usedImages)
	return usedImages
}

func (a AWS) DescribeImages(accountId string) ([]ec2Types.Image, error) {
	describeImageInput := &ec2.DescribeImagesInput{Owners: []string{"self"}}
	if accountId != "" {
		describeImageInput.Owners = append(describeImageInput.Owners, accountId)
	}
	imagesOutput, err := a.ec2.DescribeImages(context.TODO(), describeImageInput)
	if err != nil {
		return nil, err
	}
	return imagesOutput.Images, nil
}

func (a AWS) DeregisterImage(imageId string, dryRun bool) error {
	deregisterInput := &ec2.DeregisterImageInput{
		ImageId: &imageId,
		DryRun:  &dryRun,
	}
	_, err := a.ec2.DeregisterImage(context.TODO(), deregisterInput)
	return err
}

func (a AWS) GetAvailableEBSVolumes() []ec2Types.Volume {
	volumes := []ec2Types.Volume{}
	nextToken := ""

	for {
		opts := &ec2.DescribeVolumesInput{}
		if nextToken != "" {
			opts.NextToken = &nextToken
		}
		volumeOutput, err := a.ec2.DescribeVolumes(context.TODO(), opts)
		CheckError(err, logger.Errorf)
		if volumeOutput != nil {
			volumes = append(volumes, volumeOutput.Volumes...)
		}

		if volumeOutput == nil || volumeOutput.NextToken == nil {
			break
		}
	}
	return volumes
}

func (a AWS) DeleteVolume(volumeId string, dryrun bool) error {

	opts := &ec2.DeleteVolumeInput{
		VolumeId: &volumeId,
		DryRun:   &dryrun,
	}

	_, err := a.ec2.DeleteVolume(context.TODO(), opts)
	return err
}

func CheckError(err error, logFunc func(tpl string, args ...interface{})) {
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			logFunc("code: %s, message: %s, fault: %s", ae.ErrorCode(), ae.ErrorMessage(), ae.ErrorFault().String())
		} else {
			logFunc(err.Error())
		}
	}
}
