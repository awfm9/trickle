package fixture

import (
	"crypto/rand"

	"github.com/stretchr/testify/require"
)

func Sig(t require.TestingT) []byte {
	seed := make([]byte, 128)
	n, err := rand.Read(seed)
	require.NoError(t, err, "could not read random seed")
	require.Equal(t, len(seed), n, "could not read full seed")
	return seed
}
