package ebsclean

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/steffakasid/amiclean/internal"
	"github.com/steffakasid/amiclean/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/xhit/go-str2duration/v2"
)

func initEBSClean(ec2Mock *mocks.Ec2client, t *testing.T) *EBSClean {
	olderthenDuration, err := str2duration.ParseDuration("7d")
	assert.NoError(t, err)
	return &EBSClean{
		awsClient: internal.NewFromInterface(ec2Mock),
		olderthen: olderthenDuration,
	}
}

func TestNewInstance(t *testing.T) {
	mock := &mocks.Ec2client{}
	awsClient := internal.NewFromInterface(mock)
	ebsclean := NewInstance(awsClient, time.Duration(1), false)
	assert.NotNil(t, ebsclean)
	assert.Equal(t, time.Duration(1), ebsclean.olderthen)
	assert.False(t, ebsclean.dryrun)
}

func TestDeleteUnusedEBSVolumes(t *testing.T) {
	deleteWhenOlder, err := str2duration.ParseDuration("7d1h")
	assert.NoError(t, err)
	t.Run("success", func(t *testing.T) {
		mock := &mocks.Ec2client{}
		toDelete := mockDescribeVolumes(2, 2, deleteWhenOlder, mock)
		mockDeleteVolume(toDelete, false, mock)
		SUT := initEBSClean(mock, t)

		SUT.DeleteUnusedEBSVolumes()
		mock.AssertExpectations(t)
	})
}

func mockDescribeVolumes(numCalls int, numDelete int, before time.Duration, mock *mocks.Ec2client) (toDelete []string) {
	for i := 1; i <= numCalls; i++ {
		opts := ec2.DescribeVolumesInput{}
		volumes := []types.Volume{
			{
				VolumeId:   aws.String(fmt.Sprintf("i-abcde%d", rand.Int())),
				CreateTime: aws.Time(time.Now()), // is in use don't care about createtime
				State:      types.VolumeStateInUse,
			},
			{
				VolumeId:   aws.String(fmt.Sprintf("i-abcde%d", rand.Int())),
				CreateTime: aws.Time(time.Now()), // is in use don't care about createtime
				State:      types.VolumeStateInUse,
			},
			{
				VolumeId:   aws.String(fmt.Sprintf("i-abcde%d", rand.Int())),
				CreateTime: aws.Time(time.Now()), // is pretty new so will not be deleted even if state available
				State:      types.VolumeStateAvailable,
			},
		}
		for i := 1; i <= numDelete; i++ {
			id := fmt.Sprintf("i-abcde%d", rand.Int())
			createTime := time.Now().Add(before * time.Duration(rand.Intn(numDelete)) * -1)
			vol := types.Volume{
				VolumeId:   aws.String(id),
				CreateTime: aws.Time(createTime),
				State:      types.VolumeStateAvailable,
			}
			toDelete = append(toDelete, id)
			volumes = append(volumes, vol)
		}

		out := ec2.DescribeVolumesOutput{
			Volumes: volumes,
		}
		if i < numCalls {
			out.NextToken = aws.String(strconv.Itoa(i))
		}

		if i > 1 {
			opts.NextToken = aws.String(strconv.Itoa(i - 1))
		}
		mock.EXPECT().DescribeVolumes(context.TODO(), &opts).Return(&out, nil).Once()
	}
	return toDelete
}

func mockDeleteVolume(volumeIds []string, dryrun bool, mock *mocks.Ec2client) {
	for _, volumeId := range volumeIds {
		opts := ec2.DeleteVolumeInput{
			VolumeId: aws.String(volumeId),
			DryRun:   aws.Bool(dryrun),
		}
		mock.EXPECT().DeleteVolume(context.TODO(), &opts).Return(&ec2.DeleteVolumeOutput{}, nil)
	}
}
