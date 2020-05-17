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
	"github.com/awfm/consensus/model/message"
)

type Network interface {
	Broadcast(proposal *message.Proposal) error
	Transmit(vote *message.Vote, recipientID base.Hash) error
}
