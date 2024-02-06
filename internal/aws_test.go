package internal

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	cloudtrailTypes "github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
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
		dryRun := false
		expectedToken := "expected next token"

		SUT, mock, _ := setupSUT(t)

		expectedOpts := &ec2.DescribeSecurityGroupsInput{
			DryRun:     aws.Bool(dryRun),
			MaxResults: aws.Int32(100),
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
			DryRun:     aws.Bool(dryRun),
			MaxResults: aws.Int32(100),
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

		secGrps, err := SUT.GetSecurityGroups(dryRun, SecurityGroups{})
		require.NoError(t, err)
		assert.Len(t, secGrps, 2)
		mock.AssertExpectations(t)
	})

	t.Run("Error from AWS and we fail", func(t *testing.T) {
		dryRun := false
		expectedToken := "expected next token"

		SUT, mock, _ := setupSUT(t)

		expectedOpts := &ec2.DescribeSecurityGroupsInput{
			DryRun:     aws.Bool(dryRun),
			MaxResults: aws.Int32(100),
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
			DryRun:     aws.Bool(dryRun),
			MaxResults: aws.Int32(100),
			NextToken:  aws.String(expectedToken),
		}
		mock.EXPECT().DescribeSecurityGroups(context.TODO(), expectedOpts2).Return(nil, fmt.Errorf("Something went wrong")).Once()

		out, err := SUT.GetSecurityGroups(dryRun, SecurityGroups{})
		require.Error(t, err)
		require.EqualError(t, err, "Something went wrong")

		// We get back a partial result here
		assert.Len(t, out, 1)

		mock.AssertExpectations(t)
	})
}

func TestGetNotUsedSecGrpsFromENI(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		dryRun := false
		expectedGrpID1 := "1234"
		expectedGrpName1 := "groupname1"
		expectedGrpID2 := "5678"
		expectedGrpName2 := "groupname2"

		SUT, mock, _ := setupSUT(t)

		expectedOpts1 := &ec2.DescribeNetworkInterfacesInput{
			DryRun: aws.Bool(dryRun),
			Filters: []types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", expectedGrpID1)}},
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
			DryRun: aws.Bool(dryRun),
			Filters: []types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", expectedGrpID2)}},
			},
		}
		expectedOut2 := ec2.DescribeNetworkInterfacesOutput{
			NetworkInterfaces: []types.NetworkInterface{
				{
					NetworkInterfaceId: aws.String("asdf-234"),
					MacAddress:         aws.String("2"),
				},
			},
		}
		mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedOpts2).Return(&expectedOut2, nil).Once()

		secGrps := SecurityGroups{
			"Group1": SecurityGroup{
				Name: expectedGrpName1,
				ID:   expectedGrpID1,
			},
			"Group2": SecurityGroup{
				Name: expectedGrpName2,
				ID:   expectedGrpID2,
			},
		}
		notUsedSecGrps, err := SUT.GetNotUsedSecGrpsFromENI(secGrps, dryRun)
		require.NoError(t, err)
		// TODO: this is not two unused they are used by networkifacea
		assert.Len(t, notUsedSecGrps, 2)
		mock.AssertExpectations(t)
	})

	t.Run("Yes it's used", func(t *testing.T) {
		dryRun := false
		expectedGrpID1 := "1234"

		SUT, mock, _ := setupSUT(t)

		expectedOpts1 := &ec2.DescribeNetworkInterfacesInput{
			DryRun: aws.Bool(dryRun),
			Filters: []types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", expectedGrpID1)}},
			},
		}
		expectedOut1 := ec2.DescribeNetworkInterfacesOutput{}
		mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedOpts1).Return(&expectedOut1, nil).Once()

		notUsedSecGrps, err := SUT.GetNotUsedSecGrpsFromENI(SecurityGroups{"Group1": SecurityGroup{ID: expectedGrpID1}}, dryRun)
		require.NoError(t, err)
		assert.Len(t, notUsedSecGrps, 1)

		contained := false
		for _, secGrp := range notUsedSecGrps {
			if secGrp.ID == expectedGrpID1 {
				contained = true
			}
		}
		assert.True(t, contained)

		mock.AssertExpectations(t)
	})

	t.Run("Error from AWS", func(t *testing.T) {
		dryRun := false
		expectedGrpID1 := "1234"
		expectedGrpName1 := "groupname1"
		expectedGrpID2 := "5678"
		expectedGrpName2 := "groupname2"

		SUT, mock, _ := setupSUT(t)

		expectedOpts1 := &ec2.DescribeNetworkInterfacesInput{
			DryRun: aws.Bool(dryRun),
			Filters: []types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", expectedGrpID1)}},
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
			DryRun: aws.Bool(dryRun),
			Filters: []types.Filter{
				{
					Values: []string{fmt.Sprintf("Name=%s", expectedGrpID2)}},
			},
		}
		mock.EXPECT().DescribeNetworkInterfaces(context.TODO(), expectedOpts2).Return(nil, fmt.Errorf("Something went wrong")).Once()

		secGrps := SecurityGroups{
			"Group1": SecurityGroup{ID: expectedGrpID1, Name: expectedGrpName1},
			"Group2": SecurityGroup{ID: expectedGrpID2, Name: expectedGrpName2},
		}
		notUsedSecGrps, err := SUT.GetNotUsedSecGrpsFromENI(secGrps, dryRun)
		require.Error(t, err)
		require.EqualError(t, err, "Something went wrong")
		// TODO: this is not one unused it is used by networkiface
		assert.Len(t, notUsedSecGrps, 1)

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
							ResourceType: aws.String("SecurityGroup"),
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

		err := SUT.DeleteSecurityGroup(SecurityGroup{ID: expectedSecGrpID}, dryRun)
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

		err := SUT.DeleteSecurityGroup(SecurityGroup{ID: expectedSecGrpID}, dryRun)
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

func TestCheckError(t *testing.T) {
	t.Run("without error", func(t *testing.T) {
		CheckError(nil, func(tpl string, args ...interface{}) {
			t.Log("shouldn't be called")
			t.Fail()
		})
	})

	t.Run("with error", func(t *testing.T) {
		err := errors.New("error")
		CheckError(err, func(tpl string, args ...interface{}) {
			assert.Equal(t, err.Error(), tpl)
		})
	})

	t.Run("with smithy error", func(t *testing.T) {
		err := &smithy.GenericAPIError{
			Code:    "1234",
			Message: "message",
			Fault:   smithy.FaultServer,
		}
		CheckError(err, func(tpl string, args ...interface{}) {
			assert.Equal(t, "code: %s, message: %s, fault: %s", tpl)
			assert.Equal(t, err.Code, args[0])
			assert.Equal(t, err.Message, args[1])
			assert.Equal(t, err.Fault.String(), args[2])
		})
	})
}
