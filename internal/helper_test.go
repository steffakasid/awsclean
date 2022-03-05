package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsTrue(t *testing.T) {
	arr := []string{"a", "b"}
	result := contains(arr, "b")
	assert.True(t, result)
}

func TestContainsFalse(t *testing.T) {
	arr := []string{"a", "b"}
	result := contains(arr, "c")
	assert.False(t, result)
}

func TestUniqueAppendAdd(t *testing.T) {
	arr := []string{"a", "b"}
	result := uniqueAppend(arr, "c")
	assert.Contains(t, result, "c")
	assert.Len(t, result, 3)
}

func TestUniqueAppendAddNot(t *testing.T) {
	arr := []string{"a", "b"}
	result := uniqueAppend(arr, "b")
	assert.Len(t, result, 2)
}
