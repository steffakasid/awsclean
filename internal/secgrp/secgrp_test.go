package secgrp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/steffakasid/amiclean/internal"
	"github.com/steffakasid/amiclean/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xhit/go-str2duration/v2"
)

func setupSUT(t *testing.T, olderthen time.Duration, dryrun bool) (*SecGrp, *mocks.Ec2client) {
	ec2ClientMock := &mocks.Ec2client{}
	awsClient := internal.NewFromInterface(ec2ClientMock)
	SUT := NewInstance(awsClient, &olderthen, dryrun, false)
	return SUT, ec2ClientMock
}

func TestGetSecurityGroups(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedGrpID := "6987698-1243"

		olderthen, err := str2duration.ParseDuration("8d")
		dryrun := false
		unused := true
		require.NoError(t, err)

		SUT, mock := setupSUT(t, olderthen, dryrun)

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

		mock.EXPECT().DescribeSecurityGroups(context.TODO(), expectedDescribeSecGrpsOpts).Return(expectedDescribeSecGrpsOut, nil).Once()

		expectedDescribeNetIfaceOpts := &ec2.DescribeNetworkInterfacesInput{
			DryRun: aws.Bool(dryrun),
			Filters: []types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", expectedGrpID)}},
			},
		}
		expectedDescribeNetIfaceOut := &ec2.DescribeNetworkInterfacesOutput{}
		mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedDescribeNetIfaceOpts).Return(expectedDescribeNetIfaceOut, nil).Once()

		secgrps, err := SUT.GetSecurityGroups(unused)
		require.NoError(t, err)
		mock.AssertExpectations(t)
		assert.Len(t, secgrps, 1)

	})
}

func TestDeleteUnusedSecurityGroups(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		expectedGrpID := "6987698-1243"

		olderthen, err := str2duration.ParseDuration("8d")
		dryrun := false
		require.NoError(t, err)

		SUT, mock := setupSUT(t, olderthen, dryrun)

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

		mock.EXPECT().DescribeSecurityGroups(context.TODO(), expectedDescribeSecGrpsOpts).Return(expectedDescribeSecGrpsOut, nil).Once()

		expectedDescribeNetIfaceOpts := &ec2.DescribeNetworkInterfacesInput{
			DryRun: aws.Bool(dryrun),
			Filters: []types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", expectedGrpID)}},
			},
		}
		expectedDescribeNetIfaceOut := &ec2.DescribeNetworkInterfacesOutput{}
		mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedDescribeNetIfaceOpts).Return(expectedDescribeNetIfaceOut, nil).Once()

		expectedDeleteSecGrpOpts := &ec2.DeleteSecurityGroupInput{
			DryRun:  &dryrun,
			GroupId: &expectedGrpID,
		}
		mock.EXPECT().DeleteSecurityGroup(context.TODO(), expectedDeleteSecGrpOpts).Return(&ec2.DeleteSecurityGroupOutput{}, nil).Once()

		err = SUT.DeleteUnusedSecurityGroups()
		require.NoError(t, err)
		mock.AssertExpectations(t)
	})

}
