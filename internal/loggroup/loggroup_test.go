package loggroup

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/steffakasid/awsclean/internal"
	"github.com/steffakasid/awsclean/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewInstance(t *testing.T) {
	awsClient := internal.NewFromInterface(nil, nil, nil)
	olderThen := 24 * time.Hour
	dryrun := true
	onlyUnused := false

	logGrp := NewInstance(awsClient, &olderThen, dryrun, onlyUnused)

	assert.NotNil(t, logGrp)
	assert.Equal(t, awsClient, logGrp.awsClient)
	assert.Equal(t, &olderThen, logGrp.olderthen)
	assert.Equal(t, "cdaas-agent-aws-cdk-production", logGrp.exclude)
	assert.Equal(t, dryrun, logGrp.dryrun)
	assert.Equal(t, onlyUnused, logGrp.onlyUnused)
	assert.Empty(t, logGrp.used)
	assert.Empty(t, logGrp.unused)
}

func TestGetCloudWatchLogGroupsExclude(t *testing.T) {
	mockCloudWatchLogs := new(mocks.MockCloudWatchLogs)
	awsClient := internal.NewFromInterface(nil, nil, mockCloudWatchLogs)

	mockCloudWatchLogs.EXPECT().DescribeLogGroups(mock.Anything, mock.Anything).Return(&cloudwatchlogs.DescribeLogGroupsOutput{
		LogGroups: []types.LogGroup{
			{
				LogGroupName: aws.String("cdaas-agent-aws-cdk-production-log-group"),
				CreationTime: aws.Int64(time.Now().Add(-48 * time.Hour).UnixMilli()),
			},
		},
	}, nil)

	olderThen := 24 * time.Hour
	logGrp := NewInstance(awsClient, &olderThen, false, false)

	logGrp.GetCloudWatchLogGroups()

	assert.Len(t, logGrp.unused, 0)
	assert.Len(t, logGrp.used, 1)
	assert.Equal(t, "cdaas-agent-aws-cdk-production-log-group", *logGrp.used[0].LogGroupName)

	mockCloudWatchLogs.AssertExpectations(t)
}

func TestGetCloudWatchLogGroupsEmpty(t *testing.T) {
	mockCloudWatchLogs := new(mocks.MockCloudWatchLogs)
	awsClient := internal.NewFromInterface(nil, nil, mockCloudWatchLogs)

	mockCloudWatchLogs.EXPECT().DescribeLogGroups(mock.Anything, mock.Anything).Return(&cloudwatchlogs.DescribeLogGroupsOutput{
		LogGroups: []types.LogGroup{},
	}, nil)

	olderThen := 24 * time.Hour
	logGrp := NewInstance(awsClient, &olderThen, false, false)

	logGrp.GetCloudWatchLogGroups()

	assert.Len(t, logGrp.unused, 0)
	assert.Len(t, logGrp.used, 0)

	mockCloudWatchLogs.AssertExpectations(t)
}

func TestGetUnusedLogGroupsWithMock(t *testing.T) {
	mockCloudWatchLogs := new(mocks.MockCloudWatchLogs)
	awsClient := internal.NewFromInterface(nil, nil, mockCloudWatchLogs)

	olderThen := 24 * time.Hour
	logGrp := NewInstance(awsClient, &olderThen, false, false)

	logGrp.unused = []types.LogGroup{
		{
			LogGroupName: aws.String("test-log-group-1"),
		},
	}

	unusedLogGroups := logGrp.GetUnusedLogGroups()

	assert.Len(t, unusedLogGroups, 1)
	assert.Equal(t, "test-log-group-1", *unusedLogGroups[0].LogGroupName)
}

func TestDeleteUnusedWithMock(t *testing.T) {
	mockCloudWatchLogs := new(mocks.MockCloudWatchLogs)
	awsClient := internal.NewFromInterface(nil, nil, mockCloudWatchLogs)

	mockCloudWatchLogs.EXPECT().DescribeLogGroups(mock.Anything, mock.Anything).Return(&cloudwatchlogs.DescribeLogGroupsOutput{
		LogGroups: []types.LogGroup{
			{
				LogGroupName: aws.String("test-log-group-1"),
				CreationTime: aws.Int64(time.Now().Add(-48 * time.Hour).UnixMilli()),
			},
		},
	}, nil).Once()

	mockCloudWatchLogs.EXPECT().DeleteLogGroup(mock.Anything, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String("test-log-group-1"),
	}).Return(&cloudwatchlogs.DeleteLogGroupOutput{}, nil)

	olderThen := 24 * time.Hour
	logGrp := NewInstance(awsClient, &olderThen, true, false)

	logGrp.DeleteUnused()

	mockCloudWatchLogs.AssertExpectations(t)
}
