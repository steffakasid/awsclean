package amiclean

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/uuid"
	"github.com/steffakasid/amiclean/internal"
	"github.com/steffakasid/amiclean/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/xhit/go-str2duration/v2"
)

func TestNewInstance(t *testing.T) {
	confFunc := func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (cfg aws.Config, err error) {
		return aws.Config{}, nil
	}
	ec2InitFunc := func(cfg aws.Config, optFns ...func(*ec2.Options)) *ec2.Client {
		return &ec2.Client{}
	}
	awsClient := internal.NewAWSClient(confFunc, ec2InitFunc)

	amiclean := NewInstance(awsClient, 1, "1234", false, false, []string{})

	assert.NotNil(t, amiclean)
}

func initAMIClean(ec2Mock *mocks.Ec2client) *AmiClean {
	return &AmiClean{
		awsClient: internal.NewFromInterface(ec2Mock),
	}
}

func TestGetUsedAmis(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ec2Mock := &mocks.Ec2client{}
		expectedAMIIDs := mockDescribeInstances(1, ec2Mock)

		amiclean := initAMIClean(ec2Mock)
		amiclean.GetUsedAMIs()
		assert.ElementsMatch(t, expectedAMIIDs, amiclean.usedAMIs)
	})

	t.Run("With Paging", func(t *testing.T) {
		ec2Mock := &mocks.Ec2client{}
		expectedAMIIDs := mockDescribeInstances(4, ec2Mock)

		amiclean := initAMIClean(ec2Mock)
		amiclean.GetUsedAMIs()
		assert.ElementsMatch(t, expectedAMIIDs, amiclean.usedAMIs)
	})

	t.Run("Additionally from Launch Templates", func(t *testing.T) {
		ec2Mock := &mocks.Ec2client{}
		expectedAMIIDs := mockDescribeInstances(2, ec2Mock)
		expectedAMIIDs = append(expectedAMIIDs, mockDescribeLaunchTemplateVersions(2, ec2Mock)...)

		amiclean := initAMIClean(ec2Mock)
		amiclean.useLaunchTpls = true

		amiclean.GetUsedAMIs()
		assert.ElementsMatch(t, expectedAMIIDs, amiclean.usedAMIs)
	})

	t.Run("Error DescribeInstances", func(t *testing.T) {
		ec2Mock := &mocks.Ec2client{}
		mockDescribeInstances(2, ec2Mock, 2)

		amiclean := initAMIClean(ec2Mock)

		amiclean.GetUsedAMIs()
		assert.Len(t, amiclean.usedAMIs, 3)
	})
}

func TestDeleteOlderUnusedAMIs(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		ec2Mock := &mocks.Ec2client{}
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
		ec2Mock.EXPECT().DescribeImages(context.TODO(), input).Return(response, nil).Once()
		derregisterInput := &ec2.DeregisterImageInput{ImageId: aws.String("my-image-12345"), DryRun: aws.Bool(false)}
		ec2Mock.EXPECT().DeregisterImage(context.TODO(), derregisterInput).Return(nil, nil)

		olderthen, err := str2duration.ParseDuration("7d")
		assert.NoError(t, err)
		amiclean := initAMIClean(ec2Mock)
		amiclean.olderthen = olderthen
		amiclean.usedAMIs = []string{}

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
	})

	t.Run("With AWS Account", func(t *testing.T) {
		ec2Mock := &mocks.Ec2client{}
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
		ec2Mock.EXPECT().DescribeImages(context.TODO(), input).Return(response, nil).Once()
		derregisterInput := &ec2.DeregisterImageInput{ImageId: aws.String("my-image-12345"), DryRun: aws.Bool(false)}
		ec2Mock.EXPECT().DeregisterImage(context.TODO(), derregisterInput).Return(nil, nil)

		olderthen, err := str2duration.ParseDuration("7h")
		assert.NoError(t, err)
		amiclean := initAMIClean(ec2Mock)
		amiclean.awsaccount = "1234568"
		amiclean.olderthen = olderthen
		amiclean.usedAMIs = []string{}

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
	})

	t.Run("Dry Run", func(t *testing.T) {
		ec2Mock := &mocks.Ec2client{}
		input := &ec2.DescribeImagesInput{Owners: []string{"self", "123456"}}
		response := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
			},
		}
		ec2Mock.EXPECT().DescribeImages(context.TODO(), input).Return(response, nil).Once()
		derregisterInput := &ec2.DeregisterImageInput{ImageId: aws.String("my-image-12345"), DryRun: aws.Bool(true)}
		ec2Mock.EXPECT().DeregisterImage(context.TODO(), derregisterInput).Return(nil, nil)

		olderthen, err := str2duration.ParseDuration("7h")
		assert.NoError(t, err)

		amiclean := initAMIClean(ec2Mock)
		amiclean.awsaccount = "123456"
		amiclean.olderthen = olderthen
		amiclean.usedAMIs = []string{}
		amiclean.dryrun = true

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
	})

	t.Run("With Duration", func(t *testing.T) {
		ec2Mock := &mocks.Ec2client{}
		input := &ec2.DescribeImagesInput{Owners: []string{"self", "123456"}}
		sixdays, err := str2duration.ParseDuration("6d")
		creationDate := time.Now().Add(sixdays * -1).Format("2006-01-02T15:04:05.000Z")
		response := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
				{
					ImageId:      aws.String("filtered-out-id"),
					Name:         aws.String("filtered-out"),
					CreationDate: aws.String(creationDate),
				},
			},
		}
		ec2Mock.EXPECT().DescribeImages(context.TODO(), input).Return(response, nil).Once()
		derregisterInput := &ec2.DeregisterImageInput{ImageId: aws.String("my-image-12345"), DryRun: aws.Bool(true)}
		ec2Mock.EXPECT().DeregisterImage(context.TODO(), derregisterInput).Return(nil, nil)

		olderthen, err := str2duration.ParseDuration("7d")
		assert.NoError(t, err)
		amiclean := initAMIClean(ec2Mock)
		amiclean.awsaccount = "123456"
		amiclean.olderthen = olderthen
		amiclean.usedAMIs = []string{}
		amiclean.dryrun = true

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
	})

	t.Run("With Used AMIs", func(t *testing.T) {
		ec2Mock := &mocks.Ec2client{}
		input := &ec2.DescribeImagesInput{Owners: []string{"self", "123456"}}
		response := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
				{
					ImageId:      aws.String("still-in-use-id"),
					Name:         aws.String("in-use"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
			},
		}
		ec2Mock.EXPECT().DescribeImages(context.TODO(), input).Return(response, nil).Once()
		derregisterInput := &ec2.DeregisterImageInput{ImageId: aws.String("my-image-12345"), DryRun: aws.Bool(true)}
		ec2Mock.EXPECT().DeregisterImage(context.TODO(), derregisterInput).Return(nil, nil)

		olderthen, err := str2duration.ParseDuration("7h")
		assert.NoError(t, err)

		amiclean := initAMIClean(ec2Mock)
		amiclean.awsaccount = "123456"
		amiclean.olderthen = olderthen
		amiclean.usedAMIs = []string{"still-in-use-id"}
		amiclean.dryrun = true

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
	})

	t.Run("With Filter Patterns", func(t *testing.T) {
		ec2Mock := &mocks.Ec2client{}
		input := &ec2.DescribeImagesInput{Owners: []string{"self", "123456"}}
		response := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
				{
					ImageId:      aws.String("to-be-deleted-id"),
					Name:         aws.String("to-be-deleted"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
				{
					ImageId:      aws.String("still-in-use-id"),
					Name:         aws.String("filtered-out"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
			},
		}
		ec2Mock.EXPECT().DescribeImages(context.TODO(), input).Return(response, nil).Once()
		derregisterInput := &ec2.DeregisterImageInput{ImageId: aws.String("to-be-deleted-id"), DryRun: aws.Bool(true)}
		ec2Mock.EXPECT().DeregisterImage(context.TODO(), derregisterInput).Return(nil, nil)

		olderthen, err := str2duration.ParseDuration("7h")
		assert.NoError(t, err)
		amiclean := initAMIClean(ec2Mock)
		amiclean.awsaccount = "123456"
		amiclean.olderthen = olderthen
		amiclean.usedAMIs = []string{}
		amiclean.dryrun = true
		amiclean.ignorePatterns = []string{"^my-image.*", ".*ed.*"}

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
	})

	t.Run("Complete Filter Logic", func(t *testing.T) {
		ec2Mock := &mocks.Ec2client{}
		input := &ec2.DescribeImagesInput{Owners: []string{"self", "123456"}}
		sixdays, err := str2duration.ParseDuration("6d")
		creationDate := time.Now().Add(sixdays * -1).Format("2006-01-02T15:04:05.000Z")
		response := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("to-young-id"),
					Name:         aws.String("to-young"),
					CreationDate: aws.String(creationDate),
				},
				{
					ImageId:      aws.String("in-use-12345"),
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
		ec2Mock.EXPECT().DescribeImages(context.TODO(), input).Return(response, nil).Once()
		derregisterInput := &ec2.DeregisterImageInput{ImageId: aws.String("to-be-deleted-id"), DryRun: aws.Bool(false)}
		ec2Mock.EXPECT().DeregisterImage(context.TODO(), derregisterInput).Return(nil, nil)

		olderthen, err := str2duration.ParseDuration("7d")
		assert.NoError(t, err)
		amiclean := initAMIClean(ec2Mock)
		amiclean.awsaccount = "123456"
		amiclean.olderthen = olderthen
		amiclean.usedAMIs = []string{"in-use"}
		amiclean.dryrun = false
		amiclean.ignorePatterns = []string{"^filtered.*"}

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
	})

	t.Run("Error DescribeImages", func(t *testing.T) {
		ec2Mock := &mocks.Ec2client{}
		input := &ec2.DescribeImagesInput{Owners: []string{"self", "123456"}}
		ec2Mock.EXPECT().DescribeImages(context.TODO(), input).Return(nil, errors.New("Some error")).Once()

		olderthen, err := str2duration.ParseDuration("7h")
		assert.NoError(t, err)

		amiclean := initAMIClean(ec2Mock)
		amiclean.awsaccount = "123456"
		amiclean.olderthen = olderthen
		amiclean.usedAMIs = []string{}
		amiclean.dryrun = true

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "Some error")
	})

	t.Run("Error DeregeisterImage", func(t *testing.T) {
		ec2Mock := &mocks.Ec2client{}
		input := &ec2.DescribeImagesInput{Owners: []string{"self", "123456"}}
		response := &ec2.DescribeImagesOutput{
			Images: []types.Image{
				{
					ImageId:      aws.String("my-image-12345"),
					Name:         aws.String("my-image-name"),
					CreationDate: aws.String("2006-01-02T15:04:05.000Z"),
				},
			},
		}
		ec2Mock.EXPECT().DescribeImages(context.TODO(), input).Return(response, nil).Once()
		derregisterInput := &ec2.DeregisterImageInput{ImageId: aws.String("my-image-12345"), DryRun: aws.Bool(true)}
		ec2Mock.EXPECT().DeregisterImage(context.TODO(), derregisterInput).Return(nil, errors.New("Some Error"))

		olderthen, err := str2duration.ParseDuration("7h")
		assert.NoError(t, err)

		amiclean := initAMIClean(ec2Mock)
		amiclean.awsaccount = "123456"
		amiclean.olderthen = olderthen
		amiclean.usedAMIs = []string{}
		amiclean.dryrun = true

		err = amiclean.DeleteOlderUnusedAMIs()
		assert.NoError(t, err)
	})
}

func mockDescribeInstances(numCalls int, ec2Mock *mocks.Ec2client, errCalls ...int) (imageIds []string) {
	var nextToken string = ""
	for i := 1; i <= numCalls; i++ {
		previousToken := nextToken

		opts := ec2.DescribeInstancesInput{}
		if previousToken != "" {
			opts.NextToken = &previousToken
		}

		if isErrorCall(i, errCalls) {
			ec2Mock.EXPECT().DescribeInstances(context.TODO(), &opts).Return(nil, errors.New("some error")).Once()
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
			ec2Mock.EXPECT().DescribeInstances(context.TODO(), &opts).Return(&result, nil).Once()
		}

	}
	return imageIds
}

func mockDescribeLaunchTemplateVersions(numCalls int, ec2Mock *mocks.Ec2client, errCalls ...int) (imageIds []string) {
	var nextToken string = ""
	for i := 1; i <= numCalls; i++ {
		previousToken := nextToken

		opts := ec2.DescribeLaunchTemplateVersionsInput{Versions: []string{"$Latest"}}
		if previousToken != "" {
			opts.NextToken = &previousToken
		}

		if isErrorCall(i, errCalls) {
			ec2Mock.EXPECT().DescribeLaunchTemplateVersions(context.TODO(), &opts).Return(nil, errors.New("some error")).Once()
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
			ec2Mock.EXPECT().DescribeLaunchTemplateVersions(context.TODO(), &opts).Return(&result, nil).Once()
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
