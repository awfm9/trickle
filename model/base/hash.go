package base

// Hash is a 256-bit (32 bytes) hash and can be used to uniquely identify
// entities.
type Hash [32]byte

// ZeroHash is a hash where all bits are zeros.
var ZeroHash = [32]byte{}
