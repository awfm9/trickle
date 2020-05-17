// Consensus is a general purpose event-driven BFT consensus harness.
// Copyright (C) 2020 Max Wolter

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package consensus

import (
	"github.com/awfm/consensus/model/base"
)

type Graph interface {
	Extend(vertex *base.Vertex) error
	Confirm(vertexID base.Hash) error
	Contains(vertexID base.Hash) (bool, error)
	Tip() (*base.Vertex, error)
	Final() (*base.Vertex, error)
}
