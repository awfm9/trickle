package graph

import (
	"fmt"

	"github.com/alvalor/consensus/model/base"
)

// Chain represents a simple blockchain as graph for our consensus algorithm.
type Chain struct {
	finalID       base.Hash // holds the highest finalized vertex
	tipID         base.Hash // holds the best pending vertex
	vertices      map[base.Hash]*base.Vertex
	confirmations map[base.Hash]uint
}

// NewChain will create a new blockchain as a graph to back our consensus
// algorithm..
func NewChain(root *base.Vertex) *Chain {

	// create second vertex in state to streamline logic
	rootID := root.ID()
	child := base.Vertex{
		Height:     root.Height + 1,
		ParentID:   rootID,
		ProposerID: base.ZeroHash,
		ArcID:      base.ZeroHash,
	}
	childID := child.ID()

	// initialize chain with root as final and child as tip
	c := Chain{
		finalID:       rootID,
		tipID:         childID,
		vertices:      make(map[base.Hash]*base.Vertex),
		confirmations: make(map[base.Hash]uint),
	}
	c.vertices[rootID] = root
	c.vertices[childID] = &child
	return &c
}

// Extend will try to extend the current state graph with the given vertex and
// will fail if it would conflict with any part of the finalized state.
func (c *Chain) Extend(vertex *base.Vertex) error {

	// trace back from the vertex vertex until we:
	// 1) fail: find a missing link between vertex and finalized state
	// 2) fail: find a link that does not go through the latest finalized vertex
	// 3) pass: find a direct link to the latest finalized vertex
	ancestorID := vertex.ParentID
	final := c.vertices[c.finalID]
	for ancestorID != c.finalID {
		ancestor, found := c.vertices[ancestorID]
		if !found {
			return fmt.Errorf("no link to finalized state (ancestor: %x)", ancestorID)
		}
		if ancestor.Height < final.Height {
			return fmt.Errorf("invalid height for finalization (ancestor: %d, final: %d)", ancestor.Height, final.Height)
		}
		ancestorID = ancestor.ParentID
	}

	// if we reach here, the vertex is a valid extension of immutable state
	c.vertices[vertex.ID()] = vertex

	return nil
}

// Confirm will add one confirmation to the vertex with the given ID and all of
// its children until the finalized state is reached.
func (c *Chain) Confirm(vertexID base.Hash) error {

	// first, check if this vertex is pending
	vertex, exists := c.vertices[vertexID]
	if !exists {
		return fmt.Errorf("could not find vertex (%x)", vertexID)
	}

	// increase the number of confirmations by one and check if it's finalized
	c.confirmations[vertexID]++
	if c.confirmations[vertexID] >= 3 {
		c.finalID = vertexID
	}

	// NOTE: this should never happen to finalized candidtes, but we should not
	// make assumptions about the outer workings of the protocol - we simply try
	// to find the best candidate with at least one confirmation here

	// if the candidate has higher height than the current tip and more
	// confirmations, it should take over
	tip := c.vertices[c.tipID]
	if vertex.Height > tip.Height && c.confirmations[vertexID] >= c.confirmations[c.tipID] {
		c.tipID = vertexID
	}

	// it should also take over if it has the same height and more confirmations
	if vertex.Height == tip.Height && c.confirmations[vertexID] > c.confirmations[c.tipID] {
		c.tipID = vertexID
	}

	return nil
}

// Contains simply checks if a given vertex is part of the graph.
func (c *Chain) Contains(vertexID base.Hash) (bool, error) {
	_, contains := c.vertices[vertexID]
	return contains, nil
}

// Tip returns the highest confirmed vertex, which is our best candidate to
// extend the state from.
func (c *Chain) Tip() (*base.Vertex, error) {
	tip, found := c.vertices[c.tipID]
	if !found {
		return nil, fmt.Errorf("could not find tip (%x)", c.tipID)
	}
	return tip, nil
}

// Final returns the highest finalized vertex, which represents the boundary of
// the finalized state.
func (c *Chain) Final() (*base.Vertex, error) {
	final, found := c.vertices[c.finalID]
	if !found {
		return nil, fmt.Errorf("could not find final (%x)", c.finalID)
	}
	return final, nil
}
