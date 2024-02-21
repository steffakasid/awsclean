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

	tblTest := map[string]struct {
		src            SecurityGroups
		tgt            SecurityGroups
		assertExpected func(*testing.T, SecurityGroups, SecurityGroups)
		length         int
	}{
		"simple": {
			src: SecurityGroups{
				"onlySrc": &SecurityGroup{
					SecurityGroup: &types.SecurityGroup{
						GroupId:   aws.String("ID"),
						GroupName: aws.String("name"),
					},
				},
			},
			tgt: SecurityGroups{
				"onlyTarget": &SecurityGroup{
					SecurityGroup: &types.SecurityGroup{
						GroupId:   aws.String("ID"),
						GroupName: aws.String("name2"),
					},
				},
			},
			assertExpected: func(t *testing.T, src SecurityGroups, tgt SecurityGroups) {
				assert.Contains(t, tgt, "onlySrc")
			},
			length: 2,
		},
		"do not add but update": {
			src: SecurityGroups{
				"onBoth": &SecurityGroup{
					SecurityGroup: &types.SecurityGroup{
						GroupId:   aws.String("ID"),
						GroupName: aws.String("name"),
					},
					Creator:             "creator",
					CreationTime:        aws.Time(time.Now()),
					IsUsed:              true,
					AttachedToNetIfaces: []string{"someENI"},
				},
			},
			tgt: SecurityGroups{
				"onBoth": &SecurityGroup{
					SecurityGroup: &types.SecurityGroup{
						GroupId:   aws.String("ID"),
						GroupName: aws.String("name"),
					},
				},
			},
			assertExpected: func(t *testing.T, src SecurityGroups, tgt SecurityGroups) {
				expectedKey := "onBoth"
				assert.Contains(t, tgt, expectedKey)
				assert.Equal(t, tgt[expectedKey].Creator, "creator")
				assert.True(t, tgt[expectedKey].IsUsed)
				assert.Contains(t, tgt[expectedKey].AttachedToNetIfaces, "someENI")
			},
			length: 1,
		},
	}

	for name, test := range tblTest {
		t.Run(fmt.Sprint("Success ", name), func(t *testing.T) {

			AppendAll(test.src, test.tgt)
			assert.Len(t, test.tgt, test.length)
			test.assertExpected(t, test.src, test.tgt)
		})
	}
}

func TestMergeFields(t *testing.T) {

	expectedID := "1234"
	expectedName := "Name"
	expectedCreator := "Creator"
	expectedCreationTime := time.Now()
	tblTest := map[string]struct {
		src *SecurityGroup
		tgt *SecurityGroup
	}{
		"simple": {
			src: &SecurityGroup{
				SecurityGroup: &types.SecurityGroup{
					GroupId:   aws.String(expectedID),
					GroupName: aws.String(expectedName),
				},
				Creator:      expectedCreator,
				CreationTime: aws.Time(expectedCreationTime),
			},
			tgt: &SecurityGroup{
				SecurityGroup: &types.SecurityGroup{
					GroupId:   aws.String(expectedID),
					GroupName: aws.String(expectedName),
				},
			},
		},
		"src no id": {
			src: &SecurityGroup{
				SecurityGroup: &types.SecurityGroup{
					GroupName: aws.String(expectedName),
				},
				Creator:      expectedCreator,
				CreationTime: aws.Time(expectedCreationTime),
			},
			tgt: &SecurityGroup{
				SecurityGroup: &types.SecurityGroup{
					GroupId:   aws.String(expectedID),
					GroupName: aws.String(expectedName),
				},
			},
		},
		"tgt no id": {
			src: &SecurityGroup{
				SecurityGroup: &types.SecurityGroup{
					GroupName: aws.String(expectedName),
				},
				Creator:      expectedCreator,
				CreationTime: aws.Time(expectedCreationTime),
			},
			tgt: &SecurityGroup{
				SecurityGroup: &types.SecurityGroup{
					GroupName: aws.String(expectedName),
				},
			},
		},
		"src has no SecGrp detail": {
			src: &SecurityGroup{
				Creator:      expectedCreator,
				CreationTime: aws.Time(expectedCreationTime),
			},
			tgt: &SecurityGroup{
				SecurityGroup: &types.SecurityGroup{
					GroupName: aws.String(expectedName),
				},
			},
		},
	}
	for name, test := range tblTest {
		t.Run(fmt.Sprint("Success ", name), func(t *testing.T) {
			err := mergeFields(test.src, test.tgt)
			require.NoError(t, err)
			assert.Equal(t, expectedCreator, test.tgt.Creator)
			assert.Equal(t, expectedCreationTime, *test.tgt.CreationTime)
		})
	}

	t.Run("Do not overwrite tgt.Creator", func(t *testing.T) {
		expectedID := "1234"
		expectedName := "Name"
		expectedCreator := "Creator"
		src := &SecurityGroup{
			SecurityGroup: &types.SecurityGroup{
				GroupId:   aws.String(expectedID),
				GroupName: aws.String(expectedName),
			},
			Creator: *aws.String("do not use this"),
		}
		tgt := &SecurityGroup{
			SecurityGroup: &types.SecurityGroup{
				GroupId:   aws.String(expectedID),
				GroupName: aws.String(expectedName),
			},
			Creator: *aws.String(expectedCreator),
		}
		err := mergeFields(src, tgt)
		require.NoError(t, err)
		assert.Equal(t, expectedCreator, tgt.Creator)
	})

	t.Run("Do not overwrite tgt.CreationTime", func(t *testing.T) {
		expectedID := "1234"
		expectedName := "Name"
		doNotUse := time.Now().Add(24 * time.Hour * -1)
		expectedCreationTime := time.Now()
		src := &SecurityGroup{
			SecurityGroup: &types.SecurityGroup{
				GroupId:   aws.String(expectedID),
				GroupName: aws.String(expectedName),
			},
			CreationTime: &doNotUse,
		}
		tgt := &SecurityGroup{
			SecurityGroup: &types.SecurityGroup{
				GroupId:   aws.String(expectedID),
				GroupName: aws.String(expectedName),
			},
			CreationTime: &expectedCreationTime,
		}
		err := mergeFields(src, tgt)
		require.NoError(t, err)
		assert.Equal(t, expectedCreationTime, *tgt.CreationTime)
	})

	t.Run("Overwrite tgt.CreationTime if only default", func(t *testing.T) {
		expectedID := "1234"
		expectedName := "Name"
		defaultCreationtime := time.Time{}
		expectedCreationTime := time.Now()
		src := &SecurityGroup{
			SecurityGroup: &types.SecurityGroup{
				GroupId:   aws.String(expectedID),
				GroupName: aws.String(expectedName),
			},
			CreationTime: &expectedCreationTime,
		}
		tgt := &SecurityGroup{
			SecurityGroup: &types.SecurityGroup{
				GroupId:   aws.String(expectedID),
				GroupName: aws.String(expectedName),
			},
			CreationTime: &defaultCreationtime,
		}
		err := mergeFields(src, tgt)
		require.NoError(t, err)
		assert.Equal(t, expectedCreationTime, *tgt.CreationTime)
	})

	t.Run("No CreationTime so set default", func(t *testing.T) {
		expectedID := "1234"
		expectedName := "Name"
		src := &SecurityGroup{
			SecurityGroup: &types.SecurityGroup{
				GroupId:   aws.String(expectedID),
				GroupName: aws.String(expectedName),
			},
		}
		tgt := &SecurityGroup{
			SecurityGroup: &types.SecurityGroup{
				GroupId:   aws.String(expectedID),
				GroupName: aws.String(expectedName),
			},
		}
		err := mergeFields(src, tgt)
		require.NoError(t, err)
		assert.True(t, tgt.CreationTime.Compare(time.Time{}) == 0)
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
		require.EqualError(t, err, fmt.Sprintf("error mergin SecurityGroups: %v != %v", expectedName, expectedName2))
	})

	t.Run("No SecurityGroup Details", func(t *testing.T) {
		src := &SecurityGroup{}
		tgt := &SecurityGroup{}
		err := mergeFields(src, tgt)
		require.Error(t, err)
		require.EqualError(t, err, "error mergin SecurityGroups. Both objects have obj.SecurityGroup = nil")
	})
}
