package internal

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	extendedslog "github.com/steffakasid/extended-slog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	extendedslog.InitLogger()
}

func TestAddOrUpdate(t *testing.T) {

	t.Run("Add One New", func(t *testing.T) {
		expectedName := "Existing"
		expectedID := "ID"
		expectedGrp := &types.SecurityGroup{
			GroupId:   &expectedID,
			GroupName: &expectedName,
		}

		grps := SecurityGroups{}

		AddOrUpdate(grps, expectedGrp, "creator", aws.Time(time.Now()), true, []string{})

		assert.Len(t, grps, 1)
		assert.Contains(t, grps, expectedName)

	})

	t.Run("Update Existing", func(t *testing.T) {
		expectedName := "Existing"
		expectedID := "ID"
		expectedGrp := &types.SecurityGroup{
			GroupId:   &expectedID,
			GroupName: &expectedName,
		}

		grps := SecurityGroups{
			expectedName: &SecurityGroup{
				SecurityGroup: &types.SecurityGroup{
					GroupId:   &expectedID,
					GroupName: &expectedName,
				},
				CreationTime:        aws.Time(time.Now()),
				IsUsed:              true,
				AttachedToNetIfaces: []string{},
			},
		}

		AddOrUpdate(grps, expectedGrp, "creator", aws.Time(time.Now()), true, []string{})
		assert.Len(t, grps, 1)
		assert.Equal(t, "creator", grps[expectedName].Creator)
	})
}

func TestAppendAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		src := SecurityGroups{
			"onlySrc": &SecurityGroup{
				SecurityGroup: &types.SecurityGroup{
					GroupId:   aws.String("ID"),
					GroupName: aws.String("name"),
				},
				Creator:             "creator",
				CreationTime:        aws.Time(time.Now()),
				IsUsed:              false,
				AttachedToNetIfaces: []string{},
			},
		}
		target := SecurityGroups{
			"onlyTarget": &SecurityGroup{
				SecurityGroup: &types.SecurityGroup{
					GroupId:   aws.String("ID"),
					GroupName: aws.String("name2"),
				},
				Creator:             "creator",
				CreationTime:        aws.Time(time.Now()),
				IsUsed:              false,
				AttachedToNetIfaces: []string{},
			},
		}
		AppendAll(src, target)
		assert.Len(t, target, 2)
		assert.Contains(t, target, "onlySrc")
	})
}

func TestMergeFields(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		expectedID := "1234"
		expectedName := "Name"
		expectedCreator := "Creator"
		expectedCreationTime := time.Now()
		src := &SecurityGroup{
			SecurityGroup: &types.SecurityGroup{
				GroupId:   aws.String(expectedID),
				GroupName: aws.String(expectedName),
			},
			Creator:      expectedCreator,
			CreationTime: aws.Time(expectedCreationTime),
		}
		tgt := &SecurityGroup{
			SecurityGroup: &types.SecurityGroup{
				GroupId:   aws.String(expectedID),
				GroupName: aws.String(expectedName),
			},
		}
		err := mergeFields(src, tgt)
		require.NoError(t, err)
		assert.Equal(t, expectedCreator, tgt.Creator)
		assert.Equal(t, expectedCreationTime, *tgt.CreationTime)
	})

	t.Run("Not the same obj", func(t *testing.T) {
		expectedID := "1234"
		expectedID2 := "1234"
		expectedName := "Name"
		expectedName2 := "Name2"
		src := &SecurityGroup{
			SecurityGroup: &types.SecurityGroup{
				GroupId:   aws.String(expectedID),
				GroupName: aws.String(expectedName),
			},
		}
		tgt := &SecurityGroup{
			SecurityGroup: &types.SecurityGroup{
				GroupId:   aws.String(expectedID2),
				GroupName: aws.String(expectedName2),
			},
		}
		err := mergeFields(src, tgt)
		require.Error(t, err)
		require.EqualError(t, err, fmt.Sprintf("error mergig SecurityGroups: %v != %v", expectedName, expectedName2))
	})

	t.Run("No SecurityGroup Details", func(t *testing.T) {
		src := &SecurityGroup{}
		tgt := &SecurityGroup{}
		err := mergeFields(src, tgt)
		require.Error(t, err)
		require.EqualError(t, err, "error mergig SecurityGroups. Both objects have obj.SecurityGroup = nil")
	})

	t.Run("Success - Src has no details", func(t *testing.T) {
		expectedID := "1234"
		expectedName := "Name"
		expectedCreator := "Creator"
		expectedCreationTime := time.Now()
		src := &SecurityGroup{
			Creator:      expectedCreator,
			CreationTime: aws.Time(expectedCreationTime),
		}
		tgt := &SecurityGroup{
			SecurityGroup: &types.SecurityGroup{
				GroupId:   aws.String(expectedID),
				GroupName: aws.String(expectedName),
			},
		}
		err := mergeFields(src, tgt)
		require.NoError(t, err)
		assert.Equal(t, expectedCreator, tgt.Creator)
		assert.Equal(t, expectedCreationTime, *tgt.CreationTime)
	})
}
