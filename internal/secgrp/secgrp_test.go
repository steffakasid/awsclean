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
	extendedslog "github.com/steffakasid/extended-slog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xhit/go-str2duration/v2"
)

var ninetyDayOffset time.Duration

func setupSUT(t *testing.T, olderthen *time.Duration, dryrun, onlyUnused bool) (*SecGrp, *mocks.MockEc2client, *mocks.MockCloudTrail) {
	extendedslog.InitLogger()
	var err error
	ninetyDayOffset, err = str2duration.ParseDuration("90d")
	require.NoError(t, err)

	ec2ClientMock := mocks.NewMockEc2client(t)
	cloudTrailMock := mocks.NewMockCloudTrail(t)
	awsClient := internal.NewFromInterface(ec2ClientMock, cloudTrailMock)
	SUT := NewInstance(awsClient, olderthen, dryrun, onlyUnused)
	return SUT, ec2ClientMock, cloudTrailMock
}

func TestGetSecurityGroups(t *testing.T) {
	t.Run("Success Get All", func(t *testing.T) {
		expectedSecGrpID := "6987698-1243"
		expectedSecGrpName := "abcde-secgrp"
		expectedEndtime, err := time.Parse(time.DateTime, "2006-01-02 15:04:05")
		require.NoError(t, err)
		expectedStarttime := expectedEndtime.Add(ninetyDayOffset * -1)

		dryrun := false
		unused := true

		SUT, ec2Mock, cloudTrailMock := setupSUT(t, nil, dryrun, unused)

		expectedLookupEventsIn := &cloudtrail.LookupEventsInput{
			StartTime: &expectedStarttime,
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
			MaxResults: aws.Int32(100),
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
			Filters: []ec2Types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", expectedSecGrpID)}},
			},
		}
		expectedDescribeNetIfaceOut := &ec2.DescribeNetworkInterfacesOutput{}
		ec2Mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedDescribeNetIfaceOpts).Return(expectedDescribeNetIfaceOut, nil).Once()

		err = SUT.GetSecurityGroups(expectedStarttime, expectedEndtime)
		require.NoError(t, err)
		ec2Mock.AssertExpectations(t)
		assert.Len(t, SUT.unusedSecGrps, 1)
		assert.Len(t, SUT.usedSecGrps, 0)
		assert.Contains(t, SUT.unusedSecGrps, expectedSecGrpName)
		assert.Equal(t, "username", SUT.unusedSecGrps[expectedSecGrpName].Creator)
	})
	t.Run("Success Get Created 8d Ago", func(t *testing.T) {
		expectedSecGrpID := "6987698-1243"
		expectedSecGrpName := "abcde-secgrp"
		// TODO: how to get close to time.Now???
		expectedEndtime := time.Now()
		eightDayOffset, err := str2duration.ParseDuration("8d")
		require.NoError(t, err)
		expectedStarttime := expectedEndtime.Add(eightDayOffset * -1)

		dryrun := false
		unused := true

		SUT, ec2Mock, cloudTrailMock := setupSUT(t, nil, dryrun, unused)

		expectedLookupEventsIn := &cloudtrail.LookupEventsInput{
			StartTime: &expectedStarttime,
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
			Filters: []ec2Types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", expectedSecGrpID)}},
			},
		}
		expectedDescribeNetIfaceOut := &ec2.DescribeNetworkInterfacesOutput{}
		ec2Mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedDescribeNetIfaceOpts).Return(expectedDescribeNetIfaceOut, nil).Once()

		err = SUT.GetSecurityGroups(expectedStarttime, expectedEndtime)
		require.NoError(t, err)
		ec2Mock.AssertExpectations(t)
		assert.Len(t, SUT.unusedSecGrps, 1)
		assert.Len(t, SUT.usedSecGrps, 0)
		assert.Contains(t, SUT.unusedSecGrps, expectedSecGrpName)
		assert.Equal(t, "username", SUT.unusedSecGrps[expectedSecGrpName].Creator)

	})
}

func TestDeleteSecurityGroups(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		expectedSecGrpID := "6987698-1243"
		expectedSecGrpName := "abcde-secgrp"
		expectedEndtime, err := time.Parse(time.DateTime, "2006-01-02 15:04:05")
		require.NoError(t, err)
		expectedStarttime := expectedEndtime.Add(ninetyDayOffset * -1)

		dryrun := false
		onlyUnused := true

		SUT, ec2Mock, cloudTrailMock := setupSUT(t, nil, dryrun, onlyUnused)

		expectedLookupEventsIn := &cloudtrail.LookupEventsInput{
			StartTime: &expectedStarttime,
			EndTime:   &expectedEndtime,
			LookupAttributes: []cloudtrailTypes.LookupAttribute{
				{
					AttributeKey:   cloudtrailTypes.LookupAttributeKeyEventName,
					AttributeValue: aws.String("CreateSecurityGroup"),
				},
			},
		}
		cloudTrailMock.EXPECT().LookupEvents(context.TODO(), expectedLookupEventsIn).Return(&cloudtrail.LookupEventsOutput{}, nil).Once()
		expectedDescribeSecGrpsOpts := &ec2.DescribeSecurityGroupsInput{
			MaxResults: aws.Int32(100),
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
			Filters: []ec2Types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", expectedSecGrpID)}},
			},
		}
		expectedDescribeNetIfaceOut := &ec2.DescribeNetworkInterfacesOutput{}
		ec2Mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedDescribeNetIfaceOpts).Return(expectedDescribeNetIfaceOut, nil).Once()

		expectedDeleteSecGrpOpts := &ec2.DeleteSecurityGroupInput{
			DryRun:  &dryrun,
			GroupId: &expectedSecGrpID,
		}
		ec2Mock.EXPECT().DeleteSecurityGroup(context.TODO(), expectedDeleteSecGrpOpts).Return(&ec2.DeleteSecurityGroupOutput{}, nil).Once()

		err = SUT.DeleteSecurityGroups(expectedStarttime, expectedEndtime)
		require.NoError(t, err)
		ec2Mock.AssertExpectations(t)
		cloudTrailMock.AssertExpectations(t)
	})

	t.Run("Success OnlyUnused", func(t *testing.T) {
		expectedSecGrpID := "6987698-1243"
		expectedSecGrpName := "abcde-secgrp"
		expectedEndtime, err := time.Parse(time.DateTime, "2006-01-02 15:04:05")
		require.NoError(t, err)
		expectedStarttime := expectedEndtime.Add(ninetyDayOffset * -1)

		dryrun := false
		onlyUnused := true
		olderthen, err := str2duration.ParseDuration("8d")
		require.NoError(t, err)

		SUT, ec2Mock, cloudTrailMock := setupSUT(t, &olderthen, dryrun, onlyUnused)

		expectedLookupEventsIn := &cloudtrail.LookupEventsInput{
			StartTime: &expectedStarttime,
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
					EventTime: aws.Time(time.Now().Add(olderthen * -1)),
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
			MaxResults: aws.Int32(100),
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
			Filters: []ec2Types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", expectedSecGrpID)},
				},
			},
		}
		expectedDescribeNetIfaceOut := &ec2.DescribeNetworkInterfacesOutput{}
		ec2Mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedDescribeNetIfaceOpts).Return(expectedDescribeNetIfaceOut, nil).Once()

		expectedDeleteSecGrpOpts := &ec2.DeleteSecurityGroupInput{
			DryRun:  &dryrun,
			GroupId: &expectedSecGrpID,
		}
		ec2Mock.EXPECT().DeleteSecurityGroup(context.TODO(), expectedDeleteSecGrpOpts).Return(&ec2.DeleteSecurityGroupOutput{}, nil).Once()

		err = SUT.DeleteSecurityGroups(expectedStarttime, expectedEndtime)
		require.NoError(t, err)

		ec2Mock.AssertExpectations(t)
		cloudTrailMock.AssertExpectations(t)
	})

}
