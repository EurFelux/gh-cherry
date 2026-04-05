package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindTypeByName(t *testing.T) {
	types := []Type{
		{ID: "1", Name: "Bug"},
		{ID: "2", Name: "Feature"},
		{ID: "3", Name: "Task"},
	}

	t.Run("found", func(t *testing.T) {
		got, err := FindTypeByName(types, "Bug")
		require.NoError(t, err)
		assert.Equal(t, "1", got.ID)
		assert.Equal(t, "Bug", got.Name)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := FindTypeByName(types, "Epic")
		assert.ErrorContains(t, err, `issue type "Epic" not found`)
		assert.ErrorContains(t, err, "Bug")
		assert.ErrorContains(t, err, "Feature")
		assert.ErrorContains(t, err, "Task")
	})

	t.Run("empty list", func(t *testing.T) {
		_, err := FindTypeByName(nil, "Bug")
		assert.ErrorContains(t, err, `issue type "Bug" not found`)
	})
}
