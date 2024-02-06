package internal

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

func TestAddOrUpdate(t *testing.T) {

	t.Run("Add One New", func(t *testing.T) {

		grps := SecurityGroups{}

		AddOrUpdate(grps, "Name", "ID", "creator", aws.Time(time.Now()), true, []string{})

		assert.Len(t, grps, 1)
		assert.Contains(t, grps, "Name")

	})

	t.Run("Update Existing", func(t *testing.T) {

		expectedName := "Existing"

		grps := SecurityGroups{
			expectedName: SecurityGroup{
				Name:                expectedName,
				ID:                  "ID",
				CreationTime:        aws.Time(time.Now()),
				IsUsed:              true,
				AttachedToNetIfaces: []string{},
			},
		}

		AddOrUpdate(grps, expectedName, "ID", "creator", aws.Time(time.Now()), true, []string{})
		assert.Len(t, grps, 1)
		assert.Equal(t, "creator", grps[expectedName].Creator)
	})
}

func TestAppendAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		src := SecurityGroups{
			"onlySrc": SecurityGroup{
				Name:                "name",
				ID:                  "ID",
				Creator:             "creator",
				CreationTime:        aws.Time(time.Now()),
				IsUsed:              false,
				AttachedToNetIfaces: []string{},
			},
		}
		target := SecurityGroups{
			"onlyTarget": SecurityGroup{
				Name:                "name2",
				ID:                  "ID",
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
