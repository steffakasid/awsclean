package internal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/steffakasid/amiclean/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/xhit/go-str2duration/v2"
)

func TestNewInstance(t *testing.T) {
	confFunc := func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (cfg aws.Config, err error) {
		return aws.Config{}, nil
	}
	initFunc := func(cfg aws.Config, optFns ...func(*ec2.Options)) *ec2.Client {
		return &ec2.Client{}
	}

	amiclean := NewInstance(confFunc, initFunc, 1, "1234", false)

	assert.NotNil(t, amiclean)
	assert.Implements(t, (*Ec2client)(nil), amiclean.ec2client)
}

func TestGetUsedAmisSuccess(t *testing.T) {
	ec2Mock := &mocks.Ec2client{}
	opts := &ec2.DescribeInstancesInput{NextToken: nil}
	result := &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						ImageId: aws.String("1234"),
					},
				},
			},
			{
				Instances: []types.Instance{
					{
						ImageId: aws.String("53453"),
					},
					{
						ImageId: aws.String("acb"),
					},
				},
			},
		},
		NextToken: nil,
	}
	ec2Mock.EXPECT().DescribeInstances(context.TODO(), opts).Return(result, nil).Once()

	amiclean := &AmiClean{
		ec2client: ec2Mock,
	}
	amiclean.GetUsedAMIs()
	assert.Contains(t, amiclean.usedAMIs, "1234")
	assert.Contains(t, amiclean.usedAMIs, "53453")
	assert.Contains(t, amiclean.usedAMIs, "acb")
}

func TestGetUsedAmisWithPaging(t *testing.T) {
	ec2Mock := &mocks.Ec2client{}
	opts := &ec2.DescribeInstancesInput{NextToken: nil}
	result := &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						ImageId: aws.String("1234"),
					},
				},
			},
		},
		NextToken: aws.String("next"),
	}
	ec2Mock.EXPECT().DescribeInstances(context.TODO(), opts).Return(result, nil).Once()
	opts2 := &ec2.DescribeInstancesInput{NextToken: aws.String("next")}
	result2 := &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						ImageId: aws.String("4321"),
					},
				},
			},
		},
		NextToken: nil,
	}
	ec2Mock.EXPECT().DescribeInstances(context.TODO(), opts2).Return(result2, nil).Once()
	amiclean := &AmiClean{
		ec2client: ec2Mock,
	}
	amiclean.GetUsedAMIs()
	assert.Contains(t, amiclean.usedAMIs, "1234")
	assert.Contains(t, amiclean.usedAMIs, "4321")
}

func TestGetUsedAmisError(t *testing.T) {
	ec2Mock := &mocks.Ec2client{}
	opts := &ec2.DescribeInstancesInput{NextToken: nil}
	ec2Mock.EXPECT().DescribeInstances(context.TODO(), opts).Return(nil, errors.New("An Error")).Once()

	amiclean := &AmiClean{
		ec2client: ec2Mock,
	}
	amiclean.GetUsedAMIs()
	assert.Len(t, amiclean.usedAMIs, 0)
}

func TestDeleteOlderUnusedAmisSuccess(t *testing.T) {
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
	amiclean := &AmiClean{
		ec2client: ec2Mock,
		olderthen: olderthen,
		usedAMIs:  []string{},
	}
	err = amiclean.DeleteOlderUnusedAMIs()
	assert.NoError(t, err)
}

func TestDeleteOlderUnusedAmisWithAwsAccount(t *testing.T) {
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
	amiclean := &AmiClean{
		ec2client:  ec2Mock,
		awsaccount: "1234568",
		olderthen:  olderthen,
		usedAMIs:   []string{},
	}
	err = amiclean.DeleteOlderUnusedAMIs()
	assert.NoError(t, err)
}

func TestDeleteOlderUnusedAmisDryrun(t *testing.T) {
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
	amiclean := &AmiClean{
		ec2client:  ec2Mock,
		awsaccount: "123456",
		olderthen:  olderthen,
		usedAMIs:   []string{},
		dryrun:     true,
	}
	err = amiclean.DeleteOlderUnusedAMIs()
	assert.NoError(t, err)
}

func TestDeleteOlderUnusedAmisError(t *testing.T) {
	ec2Mock := &mocks.Ec2client{}
	input := &ec2.DescribeImagesInput{Owners: []string{"self", "123456"}}
	ec2Mock.EXPECT().DescribeImages(context.TODO(), input).Return(nil, errors.New("Some error")).Once()

	olderthen, err := str2duration.ParseDuration("7h")
	assert.NoError(t, err)
	amiclean := &AmiClean{
		ec2client:  ec2Mock,
		awsaccount: "123456",
		olderthen:  olderthen,
		usedAMIs:   []string{},
		dryrun:     true,
	}
	err = amiclean.DeleteOlderUnusedAMIs()
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "Some error")
}

func TestDeleteOlderUnusedAmisError2(t *testing.T) {
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
	amiclean := &AmiClean{
		ec2client:  ec2Mock,
		awsaccount: "123456",
		olderthen:  olderthen,
		usedAMIs:   []string{},
		dryrun:     true,
	}
	err = amiclean.DeleteOlderUnusedAMIs()
	assert.NoError(t, err)
}

func TestDeleteOlderUnusedAmisFilterDuration(t *testing.T) {
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
	amiclean := &AmiClean{
		ec2client:  ec2Mock,
		awsaccount: "123456",
		olderthen:  olderthen,
		usedAMIs:   []string{},
		dryrun:     true,
	}
	err = amiclean.DeleteOlderUnusedAMIs()
	assert.NoError(t, err)
}

func TestDeleteOlderUnusedAmisInuse(t *testing.T) {
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
	amiclean := &AmiClean{
		ec2client:  ec2Mock,
		awsaccount: "123456",
		olderthen:  olderthen,
		usedAMIs:   []string{"still-in-use-id"},
		dryrun:     true,
	}
	err = amiclean.DeleteOlderUnusedAMIs()
	assert.NoError(t, err)
}
