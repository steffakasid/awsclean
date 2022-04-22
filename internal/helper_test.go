package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		arr := []string{"a", "b"}
		result := contains(arr, "b")
		assert.True(t, result)
	})
	t.Run("False", func(t *testing.T) {
		arr := []string{"a", "b"}
		result := contains(arr, "c")
		assert.False(t, result)
	})
}

func TestUniqueAppend(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		arr := []string{"a", "b"}
		result := uniqueAppend(arr, "c")
		assert.Contains(t, result, "c")
		assert.Len(t, result, 3)
	})
	t.Run("AddNot", func(t *testing.T) {
		arr := []string{"a", "b"}
		result := uniqueAppend(arr, "b")
		assert.Len(t, result, 2)
	})
}

func TestMatchAny(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		regExps := []string{".*"}
		result, err := matchAny("someString", regExps)
		assert.NoError(t, err)
		assert.True(t, result)
	})
	t.Run("No Match", func(t *testing.T) {
		regExps := []string{".*abc"}
		result, err := matchAny("someString", regExps)
		assert.NoError(t, err)
		assert.False(t, result)
	})
	t.Run("Error", func(t *testing.T) {
		regExps := []string{"(?blub).*]"}
		result, err := matchAny("someString", regExps)
		assert.Error(t, err)
		assert.EqualError(t, err, "error parsing regexp: invalid or unsupported Perl syntax: `(?b`")
		assert.False(t, result)
	})
}
