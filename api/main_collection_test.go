package api

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMainCollectionId(t *testing.T) {
	t.Run("mainCollId conversion should work properly:", func(t *testing.T) {
		expected := exWaterId
		actual, err := expected.AsBase58().ToMainCollectionId()
		require.NoError(t, err)
		require.Equal(t, string(expected[:]), string(actual[:]))
		actMcid, err := Base58Str("4M2TH9NsAMM").ToMainCollectionId()
		require.NoError(t, err)
		assert.Equal(t, string(expected[:]), string(actMcid[:]))
	})
}
