package secgrp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/steffakasid/awsclean/internal"
	"github.com/steffakasid/awsclean/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xhit/go-str2duration/v2"
)

func setupSUT(t *testing.T, olderthen time.Duration, dryrun bool) (*SecGrp, *mocks.MockEc2client, *mocks.MockCloudTrail) {
	ec2ClientMock := &mocks.MockEc2client{}
	cloudTrailMock := &mocks.MockCloudTrail{}
	awsClient := internal.NewFromInterface(ec2ClientMock, cloudTrailMock)
	SUT := NewInstance(awsClient, &olderthen, dryrun, false)
	return SUT, ec2ClientMock, cloudTrailMock
}

func TestGetSecurityGroups(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedGrpID := "6987698-1243"

		olderthen, err := str2duration.ParseDuration("8d")
		dryrun := false
		unused := true
		require.NoError(t, err)

		// TODO: we might use the CloudTrailMock later too
		SUT, ec2Mock, _ := setupSUT(t, olderthen, dryrun)

		expectedDescribeSecGrpsOpts := &ec2.DescribeSecurityGroupsInput{
			DryRun:     aws.Bool(dryrun),
			MaxResults: aws.Int32(100),
		}
		expectedDescribeSecGrpsOut := &ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []types.SecurityGroup{
				{
					GroupId:   aws.String(expectedGrpID),
					GroupName: aws.String("some name"),
				},
			},
		}

		ec2Mock.EXPECT().DescribeSecurityGroups(context.TODO(), expectedDescribeSecGrpsOpts).Return(expectedDescribeSecGrpsOut, nil).Once()

		expectedDescribeNetIfaceOpts := &ec2.DescribeNetworkInterfacesInput{
			DryRun: aws.Bool(dryrun),
			Filters: []types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", expectedGrpID)}},
			},
		}
		expectedDescribeNetIfaceOut := &ec2.DescribeNetworkInterfacesOutput{}
		ec2Mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedDescribeNetIfaceOpts).Return(expectedDescribeNetIfaceOut, nil).Once()

		secgrps, err := SUT.GetSecurityGroups(unused)
		require.NoError(t, err)
		ec2Mock.AssertExpectations(t)
		assert.Len(t, secgrps, 1)

	})
}

func TestDeleteUnusedSecurityGroups(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		expectedGrpID := "6987698-1243"

		olderthen, err := str2duration.ParseDuration("8d")
		dryrun := false
		require.NoError(t, err)

		// TODO: we might use the CloudTrailMock later too
		SUT, ec2Mock, _ := setupSUT(t, olderthen, dryrun)

		expectedDescribeSecGrpsOpts := &ec2.DescribeSecurityGroupsInput{
			DryRun:     aws.Bool(dryrun),
			MaxResults: aws.Int32(100),
		}
		expectedDescribeSecGrpsOut := &ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []types.SecurityGroup{
				{
					GroupId:   aws.String(expectedGrpID),
					GroupName: aws.String("some name"),
				},
			},
		}

		ec2Mock.EXPECT().DescribeSecurityGroups(context.TODO(), expectedDescribeSecGrpsOpts).Return(expectedDescribeSecGrpsOut, nil).Once()

		expectedDescribeNetIfaceOpts := &ec2.DescribeNetworkInterfacesInput{
			DryRun: aws.Bool(dryrun),
			Filters: []types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", expectedGrpID)}},
			},
		}
		expectedDescribeNetIfaceOut := &ec2.DescribeNetworkInterfacesOutput{}
		ec2Mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedDescribeNetIfaceOpts).Return(expectedDescribeNetIfaceOut, nil).Once()

		expectedDeleteSecGrpOpts := &ec2.DeleteSecurityGroupInput{
			DryRun:  &dryrun,
			GroupId: &expectedGrpID,
		}
		ec2Mock.EXPECT().DeleteSecurityGroup(context.TODO(), expectedDeleteSecGrpOpts).Return(&ec2.DeleteSecurityGroupOutput{}, nil).Once()

		err = SUT.DeleteUnusedSecurityGroups()
		require.NoError(t, err)
		ec2Mock.AssertExpectations(t)
	})

}
