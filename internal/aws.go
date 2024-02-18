/*
Copyright © 2023 steffakasid
*/
package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	cloudtrailTypes "github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	extendedslog "github.com/steffakasid/extended-slog"
)

type Ec2client interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error)
	DeregisterImage(ctx context.Context, params *ec2.DeregisterImageInput, optFns ...func(*ec2.Options)) (*ec2.DeregisterImageOutput, error)
	DescribeLaunchTemplateVersions(ctx context.Context, params *ec2.DescribeLaunchTemplateVersionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeLaunchTemplateVersionsOutput, error)
	DescribeNetworkInterfaces(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput, opftFns ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error)
	DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
	DeleteSecurityGroup(ctx context.Context, params *ec2.DeleteSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error)
	DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error)
	DeleteVolume(ctx context.Context, params *ec2.DeleteVolumeInput, optFns ...func(*ec2.Options)) (*ec2.DeleteVolumeOutput, error)
}

type CloudTrail interface {
	LookupEvents(ctx context.Context, params *cloudtrail.LookupEventsInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.LookupEventsOutput, error)
}

type AWS struct {
	ec2        Ec2client
	cloudtrail CloudTrail
}

func NewFromInterface(ec2 Ec2client, cloudtrail CloudTrail) *AWS {
	return &AWS{
		ec2:        ec2,
		cloudtrail: cloudtrail,
	}
}

func NewAWSClient() *AWS {

	extendedslog.InitLogger()
	aws := &AWS{}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	CheckError(err, extendedslog.Logger.Fatalf)

	aws.ec2 = ec2.NewFromConfig(cfg)
	aws.cloudtrail = cloudtrail.NewFromConfig(cfg)
	return aws
}

func (a *AWS) GetSecurityGroups(secGrps SecurityGroups) (SecurityGroups, error) {
	secGrpsRet := SecurityGroups{}

	secGrpNames := []string{}
	for _, secGrp := range secGrps {
		secGrpNames = append(secGrpNames, *secGrp.SecurityGroup.GroupName)
	}

	in := &ec2.DescribeSecurityGroupsInput{
		MaxResults: aws.Int32(100),
	}

	if len(secGrpNames) > 0 {
		in.GroupNames = secGrpNames
	}

	for {
		out, err := a.ec2.DescribeSecurityGroups(context.TODO(), in)
		CheckError(err, extendedslog.Logger.Debugf)
		if nil != err {
			return secGrpsRet, err
		}

		for _, secGrp := range out.SecurityGroups {

			AddOrUpdate(secGrpsRet, &secGrp, "", nil, true, []string{})
		}

		if out.NextToken != nil {
			in.NextToken = out.NextToken
		} else {
			break
		}
	}

	extendedslog.Logger.Debugf("SecurityGroups[]: %v", secGrpsRet)
	return secGrpsRet, nil
}

func (a *AWS) GetNotUsedSecGrpsFromENI(secGrps SecurityGroups) (used SecurityGroups, unused SecurityGroups, err error) {
	used = SecurityGroups{}
	unused = SecurityGroups{}

	for _, secGrp := range secGrps {

		in := &ec2.DescribeNetworkInterfacesInput{
			Filters: []ec2Types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", *secGrp.SecurityGroup.GroupId)},
				},
			},
		}

		out, err := a.ec2.DescribeNetworkInterfaces(context.TODO(), in)
		CheckError(err, extendedslog.Logger.Debugf)
		if nil != err {
			return used, unused, err
		}
		if len(out.NetworkInterfaces) == 0 {
			extendedslog.Logger.Debug("No ENI attached to group with ID: ", *secGrp.SecurityGroup.GroupId, secGrp.SecurityGroup.GroupName)
			AddOrUpdate(unused, secGrp.SecurityGroup, secGrp.Creator, secGrp.CreationTime, false, []string{})
		}
		if len(out.NetworkInterfaces) > 0 {
			attachedIfaces := []string{}
			for _, iface := range out.NetworkInterfaces {
				attachedIfaces = append(attachedIfaces, *iface.NetworkInterfaceId)
			}
			AddOrUpdate(used, secGrp.SecurityGroup, secGrp.Creator, secGrp.CreationTime, true, attachedIfaces)
		}
	}
	return used, unused, nil
}

func (a AWS) GetCloudTrailForSecGroups(startTime, endTime time.Time) SecurityGroups {
	var nextToken string = "empty"

	secGrps := SecurityGroups{}

	for nextToken != "" {
		lookup := &cloudtrail.LookupEventsInput{
			StartTime: aws.Time(startTime),
			EndTime:   aws.Time(endTime),
			LookupAttributes: []cloudtrailTypes.LookupAttribute{
				{
					AttributeKey:   cloudtrailTypes.LookupAttributeKeyEventName,
					AttributeValue: aws.String("CreateSecurityGroup"),
				},
			},
		}
		if nextToken != "empty" {
			lookup.NextToken = aws.String(nextToken)
		} else {
			nextToken = ""
		}
		// We only get CloudTrailEvents of the last 90d: https://docs.aws.amazon.com/sdk-for-go/api/service/cloudtrail/#CloudTrail.LookupEvents
		// ResouceName: vpc-a51078cd
		// ResouceName: eksctl-eks-dev-nodegroup-apic-gw-1a-green-SG-16ACVO6XMU6HE
		// ResouceName: sg-018ce2cbe787b04ef
		// Time 2024-01-12 14:37:43 +0000 UTC
		// Wer ist schuld? `email@adress.com`
		// ---------------------------------------------
		out, err := a.cloudtrail.LookupEvents(context.TODO(), lookup)
		if out.NextToken != nil {
			nextToken = *out.NextToken
		}
		CheckError(err, extendedslog.Logger.Errorf)

		for _, ev := range out.Events {
			for _, res := range ev.Resources {

				AddOrUpdate(secGrps, &ec2Types.SecurityGroup{GroupName: res.ResourceName}, *ev.Username, ev.EventTime, true, []string{})

				extendedslog.Logger.Debug("Adding ressource", *res.ResourceName, *res.ResourceType)
			}
		}
	}

	return secGrps
}

func (a *AWS) DeleteSecurityGroup(secGrp SecurityGroup, dryrun bool) error {
	if secGrp.SecurityGroup == nil || secGrp.SecurityGroup.GroupId == nil {
		return fmt.Errorf("can not delte SecurityGroup without GroupId") // this should usually never happen
	}
	if secGrp.SecurityGroup.GroupName != nil {
		extendedslog.Logger.Debugf("DeleteSecurityGroup(%v - %v), drydrun: %t", *secGrp.SecurityGroup.GroupName, *secGrp.SecurityGroup.GroupId, dryrun)
	} else {
		extendedslog.Logger.Debugf("DeleteSecurityGroup(%v), drydrun: %t", *secGrp.SecurityGroup.GroupId, dryrun)
	}

	input := &ec2.DeleteSecurityGroupInput{
		DryRun:  &dryrun,
		GroupId: secGrp.SecurityGroup.GroupId,
	}

	_, err := a.ec2.DeleteSecurityGroup(context.TODO(), input)
	return err
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
		CheckError(err, extendedslog.Logger.Errorf)
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
	extendedslog.Logger.Debugf("UsedImages[] from EC2 %v", usedImages)
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
		CheckError(err, extendedslog.Logger.Errorf)
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
	extendedslog.Logger.Debugf("UsedImages[] from Launch Templates %v", usedImages)
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
		CheckError(err, extendedslog.Logger.Errorf)
		if volumeOutput != nil {
			volumes = append(volumes, volumeOutput.Volumes...)
		}

		if volumeOutput == nil || volumeOutput.NextToken == nil {
			break
		}
		nextToken = *volumeOutput.NextToken
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
