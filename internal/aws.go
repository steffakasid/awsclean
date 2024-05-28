/*
Copyright © 2023 steffakasid
*/
package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	cloudtrailTypes "github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/steffakasid/eslog"
)

const CLOUDTRAIL_RESOURCE_TYPE = "AWS::EC2::SecurityGroup"

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

type cloudTrailEventType string

const (
	SECURITYGROUP_CREATED cloudTrailEventType = "CreateSecurityGroup"
)

func NewFromInterface(ec2 Ec2client, cloudtrail CloudTrail) *AWS {
	return &AWS{
		ec2:        ec2,
		cloudtrail: cloudtrail,
	}
}

func NewAWSClient() *AWS {
	aws := &AWS{}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	eslog.LogIfErrorf(err, eslog.Fatalf, "aws.LoadDefaultConfig() failed: %d")

	aws.ec2 = ec2.NewFromConfig(cfg)
	aws.cloudtrail = cloudtrail.NewFromConfig(cfg)
	return aws
}

func (a *AWS) GetSecurityGroups() (SecurityGroups, error) {
	secGrpsRet := SecurityGroups{}

	in := &ec2.DescribeSecurityGroupsInput{}

	in.MaxResults = aws.Int32(1000) // it's not allowed to set in.MaxResults together with in.GroupIds

	// TODO: must use Go routines
	for {
		out, err := a.ec2.DescribeSecurityGroups(context.TODO(), in)
		eslog.LogIfError(err, eslog.Error, err)

		if nil != err {
			return secGrpsRet, err
		}

		for _, secGrp := range out.SecurityGroups {
			err := secGrpsRet.AddOrUpdate(SecurityGroup{
				SecurityGroup: &secGrp,
			})
			eslog.LogIfErrorf(err, eslog.Errorf, "GetSecurityGroups() AddOrUpdate failed %s")
		}

		if out.NextToken != nil {
			in.NextToken = out.NextToken
		} else {
			break
		}
	}

	eslog.Logger.Debugf("SecurityGroups[]: %v", secGrpsRet)
	return secGrpsRet, nil
}

// TODO: move to secgrp.go or security_group.go
func (a *AWS) GetNotUsedSecGrpsFromENI(secGrps SecurityGroups) (used *SecurityGroups, unused *SecurityGroups, err error) {
	used = &SecurityGroups{}
	unused = &SecurityGroups{}

	// TODO: must use go routines
	for _, secGrp := range secGrps {

		filter := *secGrp.SecurityGroup.GroupName
		eslog.Logger.Debugf("GetNotUsedSecGrpsFromENI(): filter %s", filter)

		in := &ec2.DescribeNetworkInterfacesInput{
			Filters: []ec2Types.Filter{
				{
					Name:   aws.String("group-name"),
					Values: []string{filter},
				},
			},
		}

		out, err := a.ec2.DescribeNetworkInterfaces(context.TODO(), in)
		if nil != err {
			return used, unused, fmt.Errorf("error describing network interfaces: %w", err)
		}

		if len(out.NetworkInterfaces) > 0 {
			attachedIfaces := []string{}
			for _, iface := range out.NetworkInterfaces {
				attachedIfaces = append(attachedIfaces, *iface.NetworkInterfaceId)
			}
			secGrp.IsUsed = true
			secGrp.AttachedToNetIfaces = attachedIfaces
			err := used.AddOrUpdate(*secGrp)
			eslog.LogIfErrorf(err, eslog.Errorf, "GetNotUsedSecGrpFromENI() AddOrUpdate() of used SecGrp failed: %s")
		} else {
			eslog.Logger.Debugf("No ENI attached to group with Name: %s", *secGrp.SecurityGroup.GroupName)
			err := unused.AddOrUpdate(*secGrp)
			eslog.LogIfErrorf(err, eslog.Errorf, "GetNotUsedSecGrpFromENI() AddOrUpdate() of unused SecGrp failed: %s")
		}
	}
	return used, unused, nil
}

// TODO: move to secgrp.go
func (a AWS) GetCloudTrailForSecGroups(startTime, endTime time.Time) SecurityGroups {
	var nextToken string = "empty"

	secGrps := SecurityGroups{}

	// TODO: must use go routines
	for nextToken != "" {
		time.Sleep(5 * time.Second)
		lookup := &cloudtrail.LookupEventsInput{
			StartTime: aws.Time(startTime),
			EndTime:   aws.Time(endTime),
			LookupAttributes: []cloudtrailTypes.LookupAttribute{
				{
					AttributeKey:   cloudtrailTypes.LookupAttributeKeyEventName, // Todo: Need to get DeleteSecurityGroup as well and delete them from the list...
					AttributeValue: aws.String(string(SECURITYGROUP_CREATED)),
				},
			},
		}
		if nextToken != "empty" {
			lookup.NextToken = aws.String(nextToken)
		}

		// We only get CloudTrailEvents of the last 90d: https://docs.aws.amazon.com/sdk-for-go/api/service/cloudtrail/#CloudTrail.LookupEvents
		out, err := a.cloudtrail.LookupEvents(context.TODO(), lookup)
		eslog.LogIfErrorf(err, eslog.Errorf, "Error on LookupEvents(): %s")
		// out could be nil if rate is exceeded
		// TODO: needs unit test
		if out != nil && out.NextToken != nil {
			nextToken = *out.NextToken
		} else {
			nextToken = ""
		}

		if out != nil {
			secGrps.AppendAll(a.getDetailsForSecGrpsFromCloudTrail(out))
		}
	}
	return secGrps
}

func (a AWS) getDetailsForSecGrpsFromCloudTrail(out *cloudtrail.LookupEventsOutput) SecurityGroups {

	additionalDetails := SecurityGroups{}

	for _, ev := range out.Events {
		for _, res := range ev.Resources {
			// TODO: needs unit testing
			if *res.ResourceType == CLOUDTRAIL_RESOURCE_TYPE {
				additionalDetails[*res.ResourceName] = &SecurityGroup{
					Creator:      *ev.Username,
					CreationTime: ev.EventTime,
				}
			}
		}
	}

	return additionalDetails
}

func (a *AWS) DeleteSecurityGroup(secGrp SecurityGroup, dryrun bool) error {
	if secGrp.SecurityGroup == nil || secGrp.SecurityGroup.GroupId == nil {
		return fmt.Errorf("can not delete SecurityGroup without GroupId") // this should usually never happen
	}
	if secGrp.SecurityGroup.GroupName != nil {
		eslog.Logger.Debugf("DeleteSecurityGroup(%v - %v), drydrun: %t", *secGrp.SecurityGroup.GroupName, *secGrp.SecurityGroup.GroupId, dryrun)
	} else {
		eslog.Logger.Debugf("DeleteSecurityGroup(%v), drydrun: %t", *secGrp.SecurityGroup.GroupId, dryrun)
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
		eslog.LogIfErrorf(err, eslog.Errorf, "Error on DescribeInstances: %s")

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
	eslog.Logger.Debugf("UsedImages[] from EC2 %v", usedImages)
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
		eslog.LogIfError(err, eslog.Error, err)
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
	eslog.Logger.Debugf("UsedImages[] from Launch Templates %v", usedImages)
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
		eslog.LogIfError(err, eslog.Error, err)

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
