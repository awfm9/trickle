package model

type QC struct {
	BlockID   Hash
	SignerIDs []Hash
	Signature []byte
}
