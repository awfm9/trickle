package fixture

import (
	"crypto/rand"
	"testing"

	"github.com/awfm/consensus/model/base"

	"github.com/stretchr/testify/require"
)

func Sig(t testing.TB) base.Signature {
	seed := make([]byte, 128)
	n, err := rand.Read(seed)
	require.NoError(t, err, "could not read random seed")
	require.Equal(t, len(seed), n, "could not read full seed")
	return seed
}
