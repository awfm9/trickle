package fixture

import (
	"crypto/rand"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	"github.com/alvalor/consensus/model"
)

func Hash(t require.TestingT) model.Hash {
	seed := make([]byte, 128)
	n, err := rand.Read(seed)
	require.NoError(t, err, "could not read random seed")
	require.Equal(t, len(seed), n, "could not read full seed")
	hash := sha3.Sum256(seed)
	return hash
}

func Hashes(t require.TestingT, n uint) []model.Hash {
	hashes := make([]model.Hash, 0, n)
	for i := 0; i < int(n); i++ {
		hashes = append(hashes, Hash(t))
	}
	return hashes
}
