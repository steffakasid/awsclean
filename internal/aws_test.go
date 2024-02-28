package internal

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	cloudtrailTypes "github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/steffakasid/awsclean/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSUT(t *testing.T) (*AWS, *mocks.MockEc2client, *mocks.MockCloudTrail) {
	ec2ClientMock := mocks.NewMockEc2client(t)
	cloudTrailMock := mocks.NewMockCloudTrail(t)
	SUT := NewFromInterface(ec2ClientMock, cloudTrailMock)
	return SUT, ec2ClientMock, cloudTrailMock
}

func TestGetSecurityGroups(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedToken := "expected next token"

		SUT, mock, _ := setupSUT(t)

		expectedOpts := &ec2.DescribeSecurityGroupsInput{
			MaxResults: aws.Int32(1000),
		}
		expectedOut := &ec2.DescribeSecurityGroupsOutput{
			NextToken: &expectedToken,
			SecurityGroups: []types.SecurityGroup{
				{
					GroupId:   aws.String("1234"),
					GroupName: aws.String("some name"),
				},
			},
		}
		mock.EXPECT().DescribeSecurityGroups(context.TODO(), expectedOpts).Return(expectedOut, nil).Once()
		expectedOpts2 := &ec2.DescribeSecurityGroupsInput{
			MaxResults: aws.Int32(1000),
			NextToken:  aws.String(expectedToken),
		}
		expectedOut2 := &ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []types.SecurityGroup{
				{
					GroupId:   aws.String("5678"),
					GroupName: aws.String("some name2"),
				},
			},
		}
		mock.EXPECT().DescribeSecurityGroups(context.TODO(), expectedOpts2).Return(expectedOut2, nil).Once()

		secGrps, err := SUT.GetSecurityGroups()
		require.NoError(t, err)
		assert.Len(t, secGrps, 2)
		mock.AssertExpectations(t)
	})

	t.Run("Error from AWS and we fail", func(t *testing.T) {
		expectedToken := "expected next token"

		SUT, mock, _ := setupSUT(t)

		expectedOpts := &ec2.DescribeSecurityGroupsInput{
			MaxResults: aws.Int32(1000),
		}
		expectedOut := &ec2.DescribeSecurityGroupsOutput{
			NextToken: &expectedToken,
			SecurityGroups: []types.SecurityGroup{
				{
					GroupId:   aws.String("1234"),
					GroupName: aws.String("some name"),
				},
			},
		}
		mock.EXPECT().DescribeSecurityGroups(context.TODO(), expectedOpts).Return(expectedOut, nil).Once()
		expectedOpts2 := &ec2.DescribeSecurityGroupsInput{
			MaxResults: aws.Int32(1000),
			NextToken:  aws.String(expectedToken),
		}
		mock.EXPECT().DescribeSecurityGroups(context.TODO(), expectedOpts2).Return(nil, fmt.Errorf("Something went wrong")).Once()

		out, err := SUT.GetSecurityGroups()
		require.Error(t, err)
		require.EqualError(t, err, "Something went wrong")

		// We get back a partial result here
		assert.Len(t, out, 1)

		mock.AssertExpectations(t)
	})
}

func TestGetNotUsedSecGrpsFromENI(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		expectedGrpID1 := "1234"
		expectedGrpName1 := "groupname1"
		expectedGrpID2 := "5678"
		expectedGrpName2 := "groupname2"

		SUT, mock, _ := setupSUT(t)

		expectedOpts1 := &ec2.DescribeNetworkInterfacesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("group-name"),
					Values: []string{expectedGrpName1}},
			},
		}
		expectedOut1 := ec2.DescribeNetworkInterfacesOutput{
			NetworkInterfaces: []types.NetworkInterface{
				{
					NetworkInterfaceId: aws.String("asdf-234"),
					MacAddress:         aws.String("1"),
				},
			},
		}
		mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedOpts1).Return(&expectedOut1, nil).Once()
		expectedOpts2 := &ec2.DescribeNetworkInterfacesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("group-name"),
					Values: []string{expectedGrpName2}},
			},
		}
		expectedOut2 := ec2.DescribeNetworkInterfacesOutput{
			NetworkInterfaces: []types.NetworkInterface{},
		}
		mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedOpts2).Return(&expectedOut2, nil).Once()

		secGrps := SecurityGroups{
			"Group1": &SecurityGroup{
				SecurityGroup: &types.SecurityGroup{
					GroupName: &expectedGrpName1,
					GroupId:   &expectedGrpID1,
				},
			},
			"Group2": &SecurityGroup{
				SecurityGroup: &types.SecurityGroup{
					GroupName: &expectedGrpName2,
					GroupId:   &expectedGrpID2,
				},
			},
		}
		usedSecGrps, notUsedSecGrps, err := SUT.GetNotUsedSecGrpsFromENI(secGrps)
		require.NoError(t, err)
		assert.Len(t, *notUsedSecGrps, 1)
		assert.Len(t, *usedSecGrps, 1)
		mock.AssertExpectations(t)
	})

	t.Run("Error from AWS", func(t *testing.T) {
		expectedGrpName1 := "groupname1"
		expectedGrpName2 := "groupname2"

		SUT, mock, _ := setupSUT(t)

		expectedOpts1 := &ec2.DescribeNetworkInterfacesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("group-name"),
					Values: []string{expectedGrpName1}},
			},
		}
		expectedOut1 := ec2.DescribeNetworkInterfacesOutput{
			NetworkInterfaces: []types.NetworkInterface{
				{
					NetworkInterfaceId: aws.String("asdf-234"),
					MacAddress:         aws.String("1"),
				},
			},
		}
		mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedOpts1).Return(&expectedOut1, nil).Once()
		expectedOpts2 := &ec2.DescribeNetworkInterfacesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("group-name"),
					Values: []string{expectedGrpName2}},
			},
		}
		mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedOpts2).Return(nil, fmt.Errorf("Something went wrong")).Once()

		secGrps := SecurityGroups{
			"Group1": &SecurityGroup{
				SecurityGroup: &types.SecurityGroup{
					GroupName: &expectedGrpName1,
				},
			},
			"Group2": &SecurityGroup{
				SecurityGroup: &types.SecurityGroup{
					GroupName: &expectedGrpName2,
				},
			},
		}
		usedSecGrps, notUsedSecGrps, err := SUT.GetNotUsedSecGrpsFromENI(secGrps)
		require.Error(t, err)
		require.EqualError(t, err, "error describing network interfaces: Something went wrong")
		assert.Len(t, *notUsedSecGrps, 0)
		assert.Len(t, *usedSecGrps, 1)

		mock.AssertExpectations(t)
	})
}

func TestGetCloudTrailForSecGroups(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		starttime, err := time.Parse(time.DateTime, "2006-01-02 15:04:05")
		require.NoError(t, err)

		endtime, err := time.Parse(time.DateTime, "2006-01-30 15:04:05")
		require.NoError(t, err)

		SUT, ec2Mock, cloudtrailMOck := setupSUT(t)

		lookupEventsIn := &cloudtrail.LookupEventsInput{
			StartTime: &starttime,
			EndTime:   &endtime,
			LookupAttributes: []cloudtrailTypes.LookupAttribute{
				{
					AttributeKey:   cloudtrailTypes.LookupAttributeKeyEventName,
					AttributeValue: aws.String("CreateSecurityGroup"),
				},
			},
		}
		cloudtrailMOck.EXPECT().LookupEvents(context.TODO(), lookupEventsIn).Return(&cloudtrail.LookupEventsOutput{
			Events: []cloudtrailTypes.Event{
				{
					EventTime: aws.Time(time.Now()),
					Username:  aws.String("someuser"),
					Resources: []cloudtrailTypes.Resource{
						{
							ResourceName: aws.String("somename"),
							ResourceType: aws.String(CLOUDTRAIL_RESOURCE_TYPE),
						},
					},
				},
			},
		}, nil)

		SUT.GetCloudTrailForSecGroups(starttime, endtime)

		ec2Mock.AssertExpectations(t)
		cloudtrailMOck.AssertExpectations(t)
	})

}

func TestDeleteSecurityGroup(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedSecGrpID := "13210-41231-21-23212-3123"
		dryRun := false

		SUT, mock, _ := setupSUT(t)
		expectedOpts := &ec2.DeleteSecurityGroupInput{
			DryRun:  &dryRun,
			GroupId: &expectedSecGrpID,
		}
		mock.EXPECT().DeleteSecurityGroup(context.TODO(), expectedOpts).Return(nil, nil).Once()

		err := SUT.DeleteSecurityGroup(SecurityGroup{
			SecurityGroup: &types.SecurityGroup{
				GroupId: &expectedSecGrpID,
			},
		}, dryRun)
		require.NoError(t, err)
	})

	t.Run("Error from AWS", func(t *testing.T) {
		expectedSecGrpID := "13210-41231-21-23212-3123"
		dryRun := false

		SUT, mock, _ := setupSUT(t)
		expectedOpts := &ec2.DeleteSecurityGroupInput{
			DryRun:  &dryRun,
			GroupId: &expectedSecGrpID,
		}

		mock.EXPECT().DeleteSecurityGroup(context.TODO(), expectedOpts).Return(nil, fmt.Errorf("Something went wrong")).Once()

		err := SUT.DeleteSecurityGroup(SecurityGroup{SecurityGroup: &types.SecurityGroup{GroupId: &expectedSecGrpID}}, dryRun)
		require.Error(t, err)
		require.EqualError(t, err, "Something went wrong")

		mock.AssertExpectations(t)
	})
}

func TestGetUsedAMIsFromEC2(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		expectedToken := "expected next token"

		SUT, mock, _ := setupSUT(t)

		expectedOpts1 := &ec2.DescribeInstancesInput{}
		expectedOutput1 := &ec2.DescribeInstancesOutput{
			NextToken: &expectedToken,
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							ImageId: aws.String("1234"),
						},
					},
				},
			},
		}
		mock.EXPECT().DescribeInstances(context.TODO(), expectedOpts1).Return(expectedOutput1, nil).Once()
		expectedOpts2 := &ec2.DescribeInstancesInput{
			NextToken: &expectedToken,
		}
		expectedOutput2 := &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							ImageId: aws.String("5678"),
						},
					},
				},
			},
		}
		mock.EXPECT().DescribeInstances(context.TODO(), expectedOpts2).Return(expectedOutput2, nil).Once()

		usedAMIs := SUT.GetUsedAMIsFromEC2()
		assert.Len(t, usedAMIs, 2)

		mock.AssertExpectations(t)
	})
}

func TestGetUsedAMIsFromLaunchTpls(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		expectedNextToken := "expected next token"

		SUT, mock, _ := setupSUT(t)

		expectedOpts1 := &ec2.DescribeLaunchTemplateVersionsInput{
			Versions: []string{"$Latest"},
		}
		expectedOutput1 := &ec2.DescribeLaunchTemplateVersionsOutput{
			NextToken: &expectedNextToken,
			LaunchTemplateVersions: []types.LaunchTemplateVersion{
				{
					LaunchTemplateData: &types.ResponseLaunchTemplateData{ImageId: aws.String("1234")},
				},
			},
		}
		mock.EXPECT().DescribeLaunchTemplateVersions(context.TODO(), expectedOpts1).Return(expectedOutput1, nil).Once()
		expectedOpts2 := &ec2.DescribeLaunchTemplateVersionsInput{
			Versions:  []string{"$Latest"},
			NextToken: &expectedNextToken,
		}
		expectedOutput2 := &ec2.DescribeLaunchTemplateVersionsOutput{
			LaunchTemplateVersions: []types.LaunchTemplateVersion{
				{
					LaunchTemplateData: &types.ResponseLaunchTemplateData{ImageId: aws.String("5678")},
				},
			},
		}
		mock.EXPECT().DescribeLaunchTemplateVersions(context.TODO(), expectedOpts2).Return(expectedOutput2, nil).Once()

		usedAmis := SUT.GetUsedAMIsFromLaunchTpls()
		assert.Len(t, usedAmis, 2)

		mock.AssertExpectations(t)
	})
}

func TestDescribeImages(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		expetedAccountID := "1234567890"

		SUT, mock, _ := setupSUT(t)
		expectedOpts := &ec2.DescribeImagesInput{
			Owners: []string{"self", expetedAccountID},
		}
		mock.EXPECT().DescribeImages(context.TODO(), expectedOpts).Return(&ec2.DescribeImagesOutput{Images: []types.Image{}}, nil).Once()

		out, err := SUT.DescribeImages(expetedAccountID)
		require.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("Error from AWS", func(t *testing.T) {
		expetedAccountID := "1234567890"

		SUT, mock, _ := setupSUT(t)
		expectedOpts := &ec2.DescribeImagesInput{
			Owners: []string{"self", expetedAccountID},
		}
		mock.EXPECT().DescribeImages(context.TODO(), expectedOpts).Return(nil, fmt.Errorf("Something went wrong")).Once()

		out, err := SUT.DescribeImages(expetedAccountID)
		assert.Nil(t, out)
		require.Error(t, err)
		require.EqualError(t, err, "Something went wrong")

		mock.AssertExpectations(t)
	})
}

func TestDeregisterImage(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		expectedImageID := "1234-543-23ffs"
		dryRun := false

		SUT, mock, _ := setupSUT(t)

		expectedOpts := &ec2.DeregisterImageInput{
			ImageId: &expectedImageID,
			DryRun:  &dryRun,
		}
		mock.EXPECT().DeregisterImage(context.TODO(), expectedOpts).Return(&ec2.DeregisterImageOutput{}, nil).Once()

		err := SUT.DeregisterImage(expectedImageID, dryRun)
		require.NoError(t, err)

		mock.AssertExpectations(t)
	})

	t.Run("Error from AWS", func(t *testing.T) {
		expectedImageID := "1234-543-23ffs"
		dryRun := false

		SUT, mock, _ := setupSUT(t)

		expectedOpts := &ec2.DeregisterImageInput{
			ImageId: &expectedImageID,
			DryRun:  &dryRun,
		}
		mock.EXPECT().DeregisterImage(context.TODO(), expectedOpts).Return(nil, fmt.Errorf("Something went wrong")).Once()

		err := SUT.DeregisterImage(expectedImageID, dryRun)
		require.Error(t, err)
		require.EqualError(t, err, "Something went wrong")

		mock.AssertExpectations(t)
	})
}

func TestGetAvailableEBSVolumes(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		SUT, mock, _ := setupSUT(t)

		expectedNextToken := "12345"
		expectedOpts1 := &ec2.DescribeVolumesInput{}
		expectedOutput1 := &ec2.DescribeVolumesOutput{
			NextToken: &expectedNextToken,
			Volumes: []types.Volume{
				{},
				{},
			},
		}
		mock.EXPECT().DescribeVolumes(context.TODO(), expectedOpts1).Return(expectedOutput1, nil).Once()
		expectedOpts2 := &ec2.DescribeVolumesInput{
			NextToken: &expectedNextToken,
		}
		expectedOutput2 := &ec2.DescribeVolumesOutput{}
		mock.EXPECT().DescribeVolumes(context.TODO(), expectedOpts2).Return(expectedOutput2, nil).Once()

		volumes := SUT.GetAvailableEBSVolumes()
		assert.Len(t, volumes, 2)

		mock.AssertExpectations(t)
	})

}

func TestDeleteVolume(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		volumeID := "1234-44555-23456"
		dryRun := false

		SUT, mock, _ := setupSUT(t)

		expectedOpts := &ec2.DeleteVolumeInput{
			VolumeId: &volumeID,
			DryRun:   &dryRun,
		}
		mock.EXPECT().DeleteVolume(context.TODO(), expectedOpts).Return(&ec2.DeleteVolumeOutput{}, nil).Once()

		err := SUT.DeleteVolume(volumeID, dryRun)
		require.NoError(t, err)

		mock.AssertExpectations(t)
	})

	t.Run("Error from AWS", func(t *testing.T) {
		volumeID := "1234-44555-23456"
		dryRun := false

		SUT, mock, _ := setupSUT(t)

		expectedOpts := &ec2.DeleteVolumeInput{
			VolumeId: &volumeID,
			DryRun:   &dryRun,
		}
		mock.EXPECT().DeleteVolume(context.TODO(), expectedOpts).Return(nil, fmt.Errorf("Something went wrong")).Once()

		err := SUT.DeleteVolume(volumeID, dryRun)
		require.Error(t, err)
		require.EqualError(t, err, "Something went wrong")

		mock.AssertExpectations(t)
	})

}
