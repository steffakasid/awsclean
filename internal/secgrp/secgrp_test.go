package secgrp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	cloudtrailTypes "github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/steffakasid/awsclean/internal"
	"github.com/steffakasid/awsclean/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSUT(t *testing.T, olderthen, createdAgo *time.Duration, dryrun, onlyUnused, showTags bool) (*SecGrp, *mocks.MockEc2client, *mocks.MockCloudTrail) {
	ec2ClientMock := mocks.NewMockEc2client(t)
	cloudTrailMock := mocks.NewMockCloudTrail(t)
	awsClient := internal.NewFromInterface(ec2ClientMock, cloudTrailMock)
	SUT := NewInstance(awsClient, olderthen, createdAgo, dryrun, onlyUnused, showTags)
	return SUT, ec2ClientMock, cloudTrailMock
}

func TestGetSecurityGroups(t *testing.T) {
	t.Run("Success Get All", func(t *testing.T) {
		expectedSecGrpID := "6987698-1243"
		expectedSecGrpName := "abcde-secgrp"

		dryrun := false
		unused := true
		showTags := false

		SUT, ec2Mock, cloudTrailMock := setupSUT(t, nil, nil, dryrun, unused, showTags)

		expectedEndtime, err := time.Parse(time.DateTime, "2006-01-02 15:04:05")
		require.NoError(t, err)

		timeMock := &mocks.MockTime{}
		timeMock.EXPECT().GetTimeP().Return(&expectedEndtime).Once()
		SUT.EndTime = timeMock

		expectedStartTime := time.Time{}

		expectedLookupEventsIn := &cloudtrail.LookupEventsInput{
			StartTime: &expectedStartTime,
			EndTime:   &expectedEndtime,
			LookupAttributes: []cloudtrailTypes.LookupAttribute{
				{
					AttributeKey:   cloudtrailTypes.LookupAttributeKeyEventName,
					AttributeValue: aws.String("CreateSecurityGroup"),
				},
			},
		}
		cloudTrailMock.EXPECT().LookupEvents(context.TODO(), expectedLookupEventsIn).Return(&cloudtrail.LookupEventsOutput{
			Events: []cloudtrailTypes.Event{
				{
					EventTime: aws.Time(time.Now()),
					Username:  aws.String("username"),
					Resources: []cloudtrailTypes.Resource{
						{
							ResourceName: aws.String(expectedSecGrpName),
							ResourceType: aws.String("SecurityGroup"),
						},
					},
				},
			},
		}, nil).Once()

		expectedDescribeSecGrpsOpts := &ec2.DescribeSecurityGroupsInput{
			DryRun:     aws.Bool(dryrun),
			MaxResults: aws.Int32(100),
			GroupNames: []string{expectedSecGrpName},
		}
		expectedDescribeSecGrpsOut := &ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []ec2Types.SecurityGroup{
				{
					GroupId:   aws.String(expectedSecGrpID),
					GroupName: aws.String(expectedSecGrpName),
				},
			},
		}

		ec2Mock.EXPECT().DescribeSecurityGroups(context.TODO(), expectedDescribeSecGrpsOpts).Return(expectedDescribeSecGrpsOut, nil).Once()

		expectedDescribeNetIfaceOpts := &ec2.DescribeNetworkInterfacesInput{
			DryRun: aws.Bool(dryrun),
			Filters: []ec2Types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", expectedSecGrpID)}},
			},
		}
		expectedDescribeNetIfaceOut := &ec2.DescribeNetworkInterfacesOutput{}
		ec2Mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedDescribeNetIfaceOpts).Return(expectedDescribeNetIfaceOut, nil).Once()

		secgrps, err := SUT.GetSecurityGroups()
		require.NoError(t, err)
		ec2Mock.AssertExpectations(t)
		assert.Len(t, secgrps, 1)

	})
	t.Run("Success Get Created 8d Ago", func(t *testing.T) {})
}

func TestDeleteSecurityGroups(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		expectedSecGrpID := "6987698-1243"
		expectedSecGrpName := "abcde-secgrp"

		dryrun := false
		onlyUnused := false
		showTags := false

		SUT, ec2Mock, cloudTrailMock := setupSUT(t, nil, nil, dryrun, onlyUnused, showTags)

		expectedEndtime, err := time.Parse(time.DateTime, "2006-01-02 15:04:05")
		require.NoError(t, err)

		timeMock := &mocks.MockTime{}
		timeMock.EXPECT().GetTimeP().Return(&expectedEndtime).Once()
		SUT.EndTime = timeMock

		expectedStartTime := time.Time{}

		expectedLookupEventsIn := &cloudtrail.LookupEventsInput{
			StartTime: &expectedStartTime,
			EndTime:   &expectedEndtime,
			LookupAttributes: []cloudtrailTypes.LookupAttribute{
				{
					AttributeKey:   cloudtrailTypes.LookupAttributeKeyEventName,
					AttributeValue: aws.String("CreateSecurityGroup"),
				},
			},
		}
		cloudTrailMock.EXPECT().LookupEvents(context.TODO(), expectedLookupEventsIn).Return(&cloudtrail.LookupEventsOutput{
			Events: []cloudtrailTypes.Event{
				{
					EventTime: aws.Time(time.Now()),
					Username:  aws.String("username"),
					Resources: []cloudtrailTypes.Resource{
						{
							ResourceName: aws.String(expectedSecGrpName),
							ResourceType: aws.String("SecurityGroup"),
						},
					},
				},
			},
		}, nil).Once()
		expectedDescribeSecGrpsOpts := &ec2.DescribeSecurityGroupsInput{
			DryRun:     aws.Bool(dryrun),
			MaxResults: aws.Int32(100),
			GroupNames: []string{expectedSecGrpName},
		}
		expectedDescribeSecGrpsOut := &ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []ec2Types.SecurityGroup{
				{
					GroupId:   aws.String(expectedSecGrpID),
					GroupName: aws.String(expectedSecGrpName),
				},
			},
		}
		ec2Mock.EXPECT().DescribeSecurityGroups(context.TODO(), expectedDescribeSecGrpsOpts).Return(expectedDescribeSecGrpsOut, nil).Once()

		expectedDeleteSecGrpOpts := &ec2.DeleteSecurityGroupInput{
			DryRun:  &dryrun,
			GroupId: &expectedSecGrpID,
		}
		ec2Mock.EXPECT().DeleteSecurityGroup(context.TODO(), expectedDeleteSecGrpOpts).Return(&ec2.DeleteSecurityGroupOutput{}, nil).Once()

		err = SUT.DeleteSecurityGroups()
		require.NoError(t, err)
		ec2Mock.AssertExpectations(t)
		cloudTrailMock.AssertExpectations(t)
		timeMock.AssertExpectations(t)
	})

	t.Run("Success OnlyUnused", func(t *testing.T) {
		expectedSecGrpID := "6987698-1243"
		expectedSecGrpName := "abcde-secgrp"

		dryrun := false
		onlyUnused := true
		showTags := false

		SUT, ec2Mock, cloudTrailMock := setupSUT(t, nil, nil, dryrun, onlyUnused, showTags)

		expectedEndtime, err := time.Parse(time.DateTime, "2006-01-02 15:04:05")
		require.NoError(t, err)

		timeMock := &mocks.MockTime{}
		timeMock.EXPECT().GetTimeP().Return(&expectedEndtime).Once()
		SUT.EndTime = timeMock

		expectedStartTime := time.Time{}

		expectedLookupEventsIn := &cloudtrail.LookupEventsInput{
			StartTime: &expectedStartTime,
			EndTime:   &expectedEndtime,
			LookupAttributes: []cloudtrailTypes.LookupAttribute{
				{
					AttributeKey:   cloudtrailTypes.LookupAttributeKeyEventName,
					AttributeValue: aws.String("CreateSecurityGroup"),
				},
			},
		}
		cloudTrailMock.EXPECT().LookupEvents(context.TODO(), expectedLookupEventsIn).Return(&cloudtrail.LookupEventsOutput{
			Events: []cloudtrailTypes.Event{
				{
					EventTime: aws.Time(time.Now()),
					Username:  aws.String("username"),
					Resources: []cloudtrailTypes.Resource{
						{
							ResourceName: aws.String(expectedSecGrpName),
							ResourceType: aws.String("SecurityGroup"),
						},
					},
				},
			},
		}, nil).Once()

		expectedNetIfaceIn := &ec2.DescribeNetworkInterfacesInput{
			DryRun: aws.Bool(dryrun),
			Filters: []ec2Types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", expectedSecGrpID)},
				},
			},
		}
		ec2Mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedNetIfaceIn).Return(&ec2.DescribeNetworkInterfacesOutput{}, nil)

		expectedDescribeSecGrpsOpts := &ec2.DescribeSecurityGroupsInput{
			DryRun:     aws.Bool(dryrun),
			MaxResults: aws.Int32(100),
			GroupNames: []string{expectedSecGrpName},
		}
		expectedDescribeSecGrpsOut := &ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []ec2Types.SecurityGroup{
				{
					GroupId:   aws.String(expectedSecGrpID),
					GroupName: aws.String(expectedSecGrpName),
				},
			},
		}
		ec2Mock.EXPECT().DescribeSecurityGroups(context.TODO(), expectedDescribeSecGrpsOpts).Return(expectedDescribeSecGrpsOut, nil).Once()

		expectedDeleteSecGrpOpts := &ec2.DeleteSecurityGroupInput{
			DryRun:  &dryrun,
			GroupId: &expectedSecGrpID,
		}
		ec2Mock.EXPECT().DeleteSecurityGroup(context.TODO(), expectedDeleteSecGrpOpts).Return(&ec2.DeleteSecurityGroupOutput{}, nil).Once()

		err = SUT.DeleteSecurityGroups()
		require.NoError(t, err)

		ec2Mock.AssertExpectations(t)
		cloudTrailMock.AssertExpectations(t)
		timeMock.AssertExpectations(t)
	})

}
