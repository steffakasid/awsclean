package internal

import (
	"errors"
	"testing"

	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
)

func TestGetUsedAMIsFromEC2(t *testing.T) {

}

func TestGetUsedAMIsFromLaunchTpls(t *testing.T) {

}

func TestDescribeImages(t *testing.T) {

}

func TestDeregisterImage(t *testing.T) {

}

func TestGetAvailableEBSVolumes(t *testing.T) {

}

func TestDeleteVolume(t *testing.T) {

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
