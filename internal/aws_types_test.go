package internal

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	extendedslog "github.com/steffakasid/extended-slog"
	"github.com/stretchr/testify/assert"
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
			expectedName: SecurityGroup{
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
			"onlySrc": SecurityGroup{
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
			"onlyTarget": SecurityGroup{
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
