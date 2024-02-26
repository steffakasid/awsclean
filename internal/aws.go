/*
Copyright © 2023 steffakasid
*/
package internal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	cloudtrailTypes "github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	extendedslog "github.com/steffakasid/extended-slog"
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

type CloudTrailEventType string

const (
	SECURITYGROUP_CREATED CloudTrailEventType = "CreateSecurityGroup"
	SECURITYGROUP_DELETED CloudTrailEventType = "DeleteSecurityGroup"
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
	extendedslog.Logger.Fatalf("aws.LoadDefaultConfig() failed: %w", err)

	aws.ec2 = ec2.NewFromConfig(cfg)
	aws.cloudtrail = cloudtrail.NewFromConfig(cfg)
	return aws
}

func (a *AWS) GetSecurityGroups(secGrpNames, secGrpIDs []string) (SecurityGroups, error) {
	secGrpsRet := SecurityGroups{}

	in := &ec2.DescribeSecurityGroupsInput{
		MaxResults: aws.Int32(100),
	}

	if len(secGrpNames) > 0 && len(secGrpIDs) > 0 {
		return secGrpsRet, errors.New("You must specify either SecuritGroupIDs or SecurityGroupNames. Not both at once.")
	}

	if len(secGrpNames) > 0 {
		in.GroupNames = secGrpNames
	}
	if len(secGrpIDs) > 0 {
		in.GroupIds = secGrpIDs
	}

	for {
		out, err := a.ec2.DescribeSecurityGroups(context.TODO(), in)
		extendedslog.Logger.Error(err)

		if nil != err {
			return secGrpsRet, err
		}

		for _, secGrp := range out.SecurityGroups {
			err := secGrpsRet.AddOrUpdate(SecurityGroup{
				SecurityGroup: &secGrp,
			})
			extendedslog.Logger.Errorf("GetSecurityGroups() AddOrUpdate failed %s", err)
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

		filter := *secGrp.SecurityGroup.GroupName
		extendedslog.Logger.Debugf("GetNotUsedSecGrpsFromENI(): filter %s", filter)

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

		if len(out.NetworkInterfaces) == 0 {
			extendedslog.Logger.Debugf("No ENI attached to group with Name: %s", *secGrp.SecurityGroup.GroupName)
			err := unused.AddOrUpdate(*secGrp)
			extendedslog.Logger.Errorf("GetNotUsedSecGrpFromENI() AddOrUpdate() of unused SecGrp failed: %s", err)
		}
		if len(out.NetworkInterfaces) > 0 {
			attachedIfaces := []string{}
			for _, iface := range out.NetworkInterfaces {
				attachedIfaces = append(attachedIfaces, *iface.NetworkInterfaceId)
			}
			secGrp.IsUsed = true
			secGrp.AttachedToNetIfaces = attachedIfaces
			err := used.AddOrUpdate(*secGrp)
			extendedslog.Logger.Errorf("GetNotUsedSecGrpFromENI() AddOrUpdate() of used SecGrp failed: %s", err)
		}
	}
	return used, unused, nil
}

func (a AWS) GetCloudTrailForSecGroups(eventType CloudTrailEventType, startTime, endTime time.Time) SecurityGroups {
	var nextToken string = "empty"

	secGrps := SecurityGroups{}

	for nextToken != "" {
		time.Sleep(5 * time.Second)
		lookup := &cloudtrail.LookupEventsInput{
			StartTime: aws.Time(startTime),
			EndTime:   aws.Time(endTime),
			LookupAttributes: []cloudtrailTypes.LookupAttribute{
				{
					AttributeKey:   cloudtrailTypes.LookupAttributeKeyEventName, // Todo: Need to get DeleteSecurityGroup as well and delete them from the list...
					AttributeValue: aws.String(string(eventType)),
				},
			},
		}
		if nextToken != "empty" {
			lookup.NextToken = aws.String(nextToken)
		}

		// We only get CloudTrailEvents of the last 90d: https://docs.aws.amazon.com/sdk-for-go/api/service/cloudtrail/#CloudTrail.LookupEvents
		out, err := a.cloudtrail.LookupEvents(context.TODO(), lookup)
		extendedslog.Logger.Errorf("Error on LookupEvents(): %s", err)
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
	var groupNames, groupIDs []string
	additionalDetails := SecurityGroups{}
	secGrps := SecurityGroups{}

	for _, ev := range out.Events {
		for _, res := range ev.Resources {
			// TODO: needs unit testing
			if *res.ResourceType == CLOUDTRAIL_RESOURCE_TYPE {
				// sg- indicates this is a GroupID
				if strings.HasPrefix(*res.ResourceName, "sg-") {
					groupIDs = UniqueAppend(groupIDs, *res.ResourceName)
				} else {
					groupNames = UniqueAppend(groupNames, *res.ResourceName)
				}
				additionalDetails[*res.ResourceName] = &SecurityGroup{
					Creator:      *ev.Username,
					CreationTime: ev.EventTime,
				}
			}
		}
	}
	// aws-sdk-go can't query with IDs and Names at once
	grpsWithName, err := a.GetSecurityGroups(groupNames, []string{})
	extendedslog.Logger.Errorf("Error getting details: %s", err)
	secGrps.AppendAll(grpsWithName)
	extendedslog.Logger.Errorf("GetCloudTrailForSecGroups() AddOrUpdate() failed: %s", err)
	grpsWithID, err := a.GetSecurityGroups([]string{}, groupIDs)
	extendedslog.Logger.Errorf("Error getting details: %s", err)
	secGrps.AppendAll(grpsWithID)

	secGrps.UpdateIfExists(additionalDetails)

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
		extendedslog.Logger.Errorf("Error on DescribeInstances: %s", err)
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
		extendedslog.Logger.Error(err)
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
		extendedslog.Logger.Error(err)

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
