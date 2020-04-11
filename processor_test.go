package consensus

import (
	"crypto/rand"
	"testing"

	"github.com/alvalor/consensus/mocks"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/sha3"
)

func TestProcessor(t *testing.T) {
	suite.Run(t, new(ProcessorSuite))
}

type ProcessorSuite struct {
	suite.Suite
	pro *Processor
}

func (ps *ProcessorSuite) SetupTest() {

	// generate identity
	seed := make([]byte, 256)
	_, _ = rand.Read(seed)
	self := sha3.Sum256(seed)

	// set up database mock
	db := &mocks.Database{}

	// set up network mock
	net := &mocks.Network{}

	// set up state mock
	state := &mocks.State{}

	// set up signer mock
	sign := &mocks.Signer{}

	// set up verify mock
	verify := &mocks.Verifier{}

	// set up the processor
	ps.pro = NewProcessor(db, net, state, sign, verify, self)
}
