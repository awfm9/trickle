package fixture

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	"github.com/awfm/consensus/model/base"
)

func Hash(t testing.TB) base.Hash {
	seed := make([]byte, 128)
	n, err := rand.Read(seed)
	require.NoError(t, err, "could not read random seed")
	require.Equal(t, len(seed), n, "could not read full seed")
	hash := sha3.Sum256(seed)
	return hash
}

func Hashes(t testing.TB, n uint) []base.Hash {
	hashes := make([]base.Hash, 0, n)
	for i := 0; i < int(n); i++ {
		hashes = append(hashes, Hash(t))
	}
	return hashes
}
