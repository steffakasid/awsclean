package amiclean

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/uuid"
	"github.com/steffakasid/awsclean/internal"
	"github.com/steffakasid/awsclean/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xhit/go-str2duration/v2"
)

func setupSUT(t *testing.T,
	olderthen time.Duration,
	awsaccount string,
	dryrun bool,
	onlyUnused bool,
	useLaunchTpls bool,
	ignorePatterns []string) (*AmiClean, *mocks.MockEc2client, *mocks.MockCloudTrail) {
	ec2ClientMock := &mocks.MockEc2client{}
	cloudTrailMock := &mocks.MockCloudTrail{}

	awsClient := internal.NewFromInterface(ec2ClientMock, cloudTrailMock)
	return NewInstance(awsClient, olderthen, awsaccount, dryrun, onlyUnused, useLaunchTpls, ignorePatterns), ec2ClientMock, cloudTrailMock
}

const (
	noAWSAccount      = ""
	noDryrun          = false
	notOnlyUnused     = false
	dontUseLaunchTpls = false
)

var noFilterPatterns = []string{}

func TestGetUsedAmis(t *testing.T) {
	defaultOlderthen, err := str2duration.ParseDuration("7d")
	assert.NoError(t, err)

	t.Run("Success", func(t *testing.T) {

		amiclean, ec2ClientMock, _ := setupSUT(t, defaultOlderthen, noAWSAccount, noDryrun, notOnlyUnused, dontUseLaunchTpls, noFilterPatterns)

		mockDescribeInstances(1, ec2ClientMock)

		expectedImgIn := &ec2.DescribeImagesInput{Owners: []string{"self"}}
		expectedImgOut := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
			},
		}
		ec2ClientMock.EXPECT().DescribeImages(context.TODO(), expectedImgIn).Return(expectedImgOut, nil).Once()

		err := amiclean.GetAMIs()
		require.NoError(t, err)
		assert.Len(t, amiclean.GetAllAMIs(), 1)
		ec2ClientMock.AssertExpectations(t)
	})

	t.Run("With Paging", func(t *testing.T) {
		amiclean, ec2ClientMock, _ := setupSUT(t, defaultOlderthen, noAWSAccount, noDryrun, notOnlyUnused, dontUseLaunchTpls, noFilterPatterns)

		mockDescribeInstances(4, ec2ClientMock)

		expectedImgIn := &ec2.DescribeImagesInput{Owners: []string{"self"}}
		expectedImgOut := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
			},
		}
		ec2ClientMock.EXPECT().DescribeImages(context.TODO(), expectedImgIn).Return(expectedImgOut, nil).Once()

		err := amiclean.GetAMIs()
		require.NoError(t, err)
		assert.Len(t, amiclean.GetAllAMIs(), 1)

		ec2ClientMock.AssertExpectations(t)
	})

	t.Run("Additionally from Launch Templates", func(t *testing.T) {
		const useLaunchTpls = true
		amiclean, ec2ClientMock, _ := setupSUT(t, defaultOlderthen, noAWSAccount, noDryrun, notOnlyUnused, useLaunchTpls, noFilterPatterns)

		mockDescribeInstances(2, ec2ClientMock)
		mockDescribeLaunchTemplateVersions(2, ec2ClientMock)

		expectedImgIn := &ec2.DescribeImagesInput{Owners: []string{"self"}}
		expectedImgOut := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
			},
		}
		ec2ClientMock.EXPECT().DescribeImages(context.TODO(), expectedImgIn).Return(expectedImgOut, nil).Once()

		err := amiclean.GetAMIs()
		require.NoError(t, err)
		assert.Len(t, amiclean.GetAllAMIs(), 1)
	})

	t.Run("Error DescribeInstances", func(t *testing.T) {
		amiclean, ec2ClientMock, _ := setupSUT(t, defaultOlderthen, noAWSAccount, noDryrun, notOnlyUnused, dontUseLaunchTpls, noFilterPatterns)

		mockDescribeInstances(2, ec2ClientMock, 2)

		expectedImgIn := &ec2.DescribeImagesInput{Owners: []string{"self"}}
		expectedImgOut := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
			},
		}
		ec2ClientMock.EXPECT().DescribeImages(context.TODO(), expectedImgIn).Return(expectedImgOut, nil).Once()

		err := amiclean.GetAMIs()
		require.NoError(t, err)
		assert.Len(t, amiclean.GetAllAMIs(), 1)
	})
}

func TestDeleteOlderUnusedAMIs(t *testing.T) {
	defaultOlderthen, err := str2duration.ParseDuration("7d")
	assert.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		amiclean, ec2ClientMock, _ := setupSUT(t, defaultOlderthen, noAWSAccount, noDryrun, notOnlyUnused, dontUseLaunchTpls, noFilterPatterns)

		mockDescribeInstances(1, ec2ClientMock)

		input := &ec2.DescribeImagesInput{Owners: []string{"self"}}
		response := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
			},
		}
		ec2ClientMock.EXPECT().DescribeImages(context.TODO(), input).Return(response, nil).Once()
		derregisterInput := &ec2.DeregisterImageInput{ImageId: aws.String("my-image-12345"), DryRun: aws.Bool(false)}
		ec2ClientMock.EXPECT().DeregisterImage(context.TODO(), derregisterInput).Return(nil, nil)

		err := amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
		ec2ClientMock.AssertExpectations(t)
	})

	t.Run("With AWS Account", func(t *testing.T) {
		const awsaccount = "1234568"
		amiclean, ec2ClientMock, _ := setupSUT(t, defaultOlderthen, awsaccount, noDryrun, notOnlyUnused, dontUseLaunchTpls, noFilterPatterns)

		mockDescribeInstances(2, ec2ClientMock, 2)

		input := &ec2.DescribeImagesInput{Owners: []string{"self", "1234568"}}
		response := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
			},
		}
		ec2ClientMock.EXPECT().DescribeImages(context.TODO(), input).Return(response, nil).Once()
		derregisterInput := &ec2.DeregisterImageInput{ImageId: aws.String("my-image-12345"), DryRun: aws.Bool(false)}
		ec2ClientMock.EXPECT().DeregisterImage(context.TODO(), derregisterInput).Return(nil, nil)

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
		ec2ClientMock.AssertExpectations(t)
	})

	t.Run("Dry Run", func(t *testing.T) {
		amiclean, ec2ClientMock, _ := setupSUT(t, defaultOlderthen, noAWSAccount, noDryrun, notOnlyUnused, dontUseLaunchTpls, noFilterPatterns)
		input := &ec2.DescribeImagesInput{Owners: []string{"self", "123456"}}
		olderthen, err := str2duration.ParseDuration("7h")
		assert.NoError(t, err)
		amiclean.awsaccount = "123456"
		amiclean.olderthen = olderthen
		amiclean.dryrun = true

		mockDescribeInstances(2, ec2ClientMock, 2)

		response := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
			},
		}
		ec2ClientMock.EXPECT().DescribeImages(context.TODO(), input).Return(response, nil).Once()
		derregisterInput := &ec2.DeregisterImageInput{ImageId: aws.String("my-image-12345"), DryRun: aws.Bool(true)}
		ec2ClientMock.EXPECT().DeregisterImage(context.TODO(), derregisterInput).Return(nil, nil)

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
		ec2ClientMock.AssertExpectations(t)
	})

	// TODO: validate that this really works
	t.Run("With Duration", func(t *testing.T) {
		amiclean, ec2ClientMock, _ := setupSUT(t, defaultOlderthen, noAWSAccount, noDryrun, notOnlyUnused, dontUseLaunchTpls, noFilterPatterns)
		mockDescribeInstances(1, ec2ClientMock)

		input := &ec2.DescribeImagesInput{Owners: []string{"self"}}
		sixdays, err := str2duration.ParseDuration("6d")
		require.NoError(t, err)
		creationDate := time.Now().Add(sixdays * -1).Format("2006-01-02T15:04:05.000Z")
		response := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
				{
					ImageId:      aws.String("filtered-out"),
					Name:         aws.String("filtered-out"),
					CreationDate: aws.String(creationDate),
				},
			},
		}
		ec2ClientMock.EXPECT().DescribeImages(context.TODO(), input).Return(response, nil).Once()
		derregisterInput := &ec2.DeregisterImageInput{ImageId: aws.String("my-image-12345"), DryRun: aws.Bool(noDryrun)}
		ec2ClientMock.EXPECT().DeregisterImage(context.TODO(), derregisterInput).Return(nil, nil)

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
		ec2ClientMock.AssertExpectations(t)
	})

	t.Run("With Used AMIs", func(t *testing.T) {
		amiclean, ec2ClientMock, _ := setupSUT(t, defaultOlderthen, noAWSAccount, noDryrun, notOnlyUnused, dontUseLaunchTpls, noFilterPatterns)

		usedAMIs := mockDescribeInstances(1, ec2ClientMock)

		input := &ec2.DescribeImagesInput{Owners: []string{"self"}}
		response := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
				{
					ImageId:      aws.String(usedAMIs[0]),
					Name:         aws.String("in-use"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
			},
		}
		ec2ClientMock.EXPECT().DescribeImages(context.TODO(), input).Return(response, nil).Once()

		derregisterInput2 := &ec2.DeregisterImageInput{ImageId: aws.String("my-image-12345"), DryRun: aws.Bool(noDryrun)}
		ec2ClientMock.EXPECT().DeregisterImage(context.TODO(), derregisterInput2).Return(nil, nil)

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
		ec2ClientMock.AssertExpectations(t)
	})

	t.Run("With Filter Patterns", func(t *testing.T) {
		ignorePatterns := []string{"^my-image.*", ".*ed.*"}
		amiclean, ec2ClientMock, _ := setupSUT(t, defaultOlderthen, noAWSAccount, noDryrun, notOnlyUnused, dontUseLaunchTpls, ignorePatterns)

		mockDescribeInstances(2, ec2ClientMock, 2)

		input := &ec2.DescribeImagesInput{Owners: []string{"self"}}
		response := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
				{
					ImageId:      aws.String("to-be-deleted-id"),
					Name:         aws.String("delete-it"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
				{
					ImageId:      aws.String("still-in-use-id"),
					Name:         aws.String("filtered-out"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
			},
		}
		ec2ClientMock.EXPECT().DescribeImages(context.TODO(), input).Return(response, nil).Once()
		derregisterInput := &ec2.DeregisterImageInput{ImageId: aws.String("to-be-deleted-id"), DryRun: aws.Bool(noDryrun)}
		ec2ClientMock.EXPECT().DeregisterImage(context.TODO(), derregisterInput).Return(nil, nil)

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
		ec2ClientMock.AssertExpectations(t)
	})

	t.Run("Complete Filter Logic", func(t *testing.T) {
		const awsaccount = "123456"
		ignorePatterns := []string{"^filtered.*"}
		amiclean, ec2ClientMock, _ := setupSUT(t, defaultOlderthen, awsaccount, noDryrun, notOnlyUnused, dontUseLaunchTpls, ignorePatterns)

		usedAMIs := mockDescribeInstances(1, ec2ClientMock)

		expectedDescribeImgIn := &ec2.DescribeImagesInput{Owners: []string{"self", "123456"}}
		sixdays, err := str2duration.ParseDuration("6d")
		require.NoError(t, err)
		creationDate := time.Now().Add(sixdays * -1).Format("2006-01-02T15:04:05.000Z")
		expectedDescribeImgOut := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("to-young-id"),
					Name:         aws.String("to-young"),
					CreationDate: aws.String(creationDate),
				},
				{
					ImageId:      aws.String(usedAMIs[0]),
					Name:         aws.String("in-use"),
					CreationDate: aws.String(creationDate),
				},
				{
					ImageId:      aws.String("filtered-out-id"),
					Name:         aws.String("filtered-out"),
					CreationDate: aws.String(creationDate),
				},
				{
					ImageId:      aws.String("to-be-deleted-id"),
					Name:         aws.String("to-be-deleted"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
			},
		}
		ec2ClientMock.EXPECT().DescribeImages(context.TODO(), expectedDescribeImgIn).Return(expectedDescribeImgOut, nil).Once()
		derregisterInput := &ec2.DeregisterImageInput{ImageId: aws.String("to-be-deleted-id"), DryRun: aws.Bool(false)}
		ec2ClientMock.EXPECT().DeregisterImage(context.TODO(), derregisterInput).Return(nil, nil)

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
		ec2ClientMock.AssertExpectations(t)
	})

	t.Run("Error DescribeImages", func(t *testing.T) {
		amiclean, ec2ClientMock, _ := setupSUT(t, defaultOlderthen, noAWSAccount, noDryrun, notOnlyUnused, dontUseLaunchTpls, noFilterPatterns)

		mockDescribeInstances(1, ec2ClientMock)

		input := &ec2.DescribeImagesInput{Owners: []string{"self"}}
		ec2ClientMock.EXPECT().DescribeImages(context.TODO(), input).Return(nil, errors.New("Some error")).Once()

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "Some error")
	})

	t.Run("Error DeregeisterImage", func(t *testing.T) {
		amiclean, ec2ClientMock, _ := setupSUT(t, defaultOlderthen, noAWSAccount, noDryrun, notOnlyUnused, dontUseLaunchTpls, noFilterPatterns)

		mockDescribeInstances(2, ec2ClientMock)

		input := &ec2.DescribeImagesInput{Owners: []string{"self"}}
		response := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
			},
		}
		ec2ClientMock.EXPECT().DescribeImages(context.TODO(), input).Return(response, nil).Once()
		derregisterInput := &ec2.DeregisterImageInput{ImageId: aws.String("my-image-12345"), DryRun: aws.Bool(noDryrun)}
		ec2ClientMock.EXPECT().DeregisterImage(context.TODO(), derregisterInput).Return(nil, errors.New("Some Error"))

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
	})
}

func mockDescribeInstances(numCalls int, ec2ClientMock *mocks.MockEc2client, errCalls ...int) (imageIds []string) {
	var nextToken string = ""
	for i := 1; i <= numCalls; i++ {
		previousToken := nextToken

		opts := ec2.DescribeInstancesInput{}
		if previousToken != "" {
			opts.NextToken = &previousToken
		}

		if isErrorCall(i, errCalls) {
			ec2ClientMock.EXPECT().DescribeInstances(context.TODO(), &opts).Return(nil, errors.New("some error")).Once()
		} else {
			id1 := uuid.NewString()
			imageIds = append(imageIds, id1)
			id2 := uuid.NewString()
			imageIds = append(imageIds, id2)
			id3 := uuid.NewString()
			imageIds = append(imageIds, id3)
			result := ec2.DescribeInstancesOutput{
				Reservations: []types.Reservation{
					{
						Instances: []types.Instance{
							{
								ImageId: aws.String(id1),
							},
						},
					},
					{
						Instances: []types.Instance{
							{
								ImageId: aws.String(id2),
							},
							{
								ImageId: aws.String(id3),
							},
						},
					},
				},
			}
			if i < numCalls {
				result.NextToken = aws.String(uuid.NewString())
				nextToken = *result.NextToken
			} else {
				nextToken = ""
			}
			ec2ClientMock.EXPECT().DescribeInstances(context.TODO(), &opts).Return(&result, nil).Once()
		}

	}
	return imageIds
}

func mockDescribeLaunchTemplateVersions(numCalls int, ec2ClientMock *mocks.MockEc2client, errCalls ...int) (imageIds []string) {
	var nextToken string = ""
	for i := 1; i <= numCalls; i++ {
		previousToken := nextToken

		opts := ec2.DescribeLaunchTemplateVersionsInput{Versions: []string{"$Latest"}}
		if previousToken != "" {
			opts.NextToken = &previousToken
		}

		if isErrorCall(i, errCalls) {
			ec2ClientMock.EXPECT().DescribeLaunchTemplateVersions(context.TODO(), &opts).Return(nil, errors.New("some error")).Once()
		} else {
			id1 := uuid.NewString()
			imageIds = append(imageIds, id1)
			id2 := uuid.NewString()
			imageIds = append(imageIds, id2)
			id3 := uuid.NewString()
			imageIds = append(imageIds, id3)
			result := ec2.DescribeLaunchTemplateVersionsOutput{
				LaunchTemplateVersions: []types.LaunchTemplateVersion{
					{
						LaunchTemplateData: &types.ResponseLaunchTemplateData{
							ImageId: &id1,
						},
					},
					{
						LaunchTemplateData: &types.ResponseLaunchTemplateData{
							ImageId: &id2,
						},
					},
					{
						LaunchTemplateData: &types.ResponseLaunchTemplateData{
							ImageId: &id3,
						},
					},
				},
			}
			if i < numCalls {
				result.NextToken = aws.String(uuid.NewString())
				nextToken = *result.NextToken
			} else {
				nextToken = ""
			}
			ec2ClientMock.EXPECT().DescribeLaunchTemplateVersions(context.TODO(), &opts).Return(&result, nil).Once()
		}

	}
	return imageIds
}

func isErrorCall(call int, errCallIDs []int) bool {
	for _, errCallID := range errCallIDs {
		if call == errCallID {
			return true
		}
	}
	return false
}
