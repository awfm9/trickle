package model

// Hash represents a 256-bit (32 bytes) hash that can be used to uniquely
// identify entities in our system.
type Hash [32]byte

// ZeroHash represens an empty hash, to be used as zero value.
var ZeroHash = [32]byte{}
