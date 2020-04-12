package model

type QC struct {
	Height    uint64    // height is the height of the referenced vertex
	VertexID  Hash      // vertex is the identifier of the referenced vertex
	SignerIDs []Hash    // signers is a list of signers who voted for the vertex
	Signature Signature // signature is the aggregated signature of all signers
}
