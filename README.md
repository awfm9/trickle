# Alvalor Consensus

## Description

Alvalor consensus implements a byzantine fault-tolerant (BFT) consensus algorithm, based on HotStuff.

## Roadmap

### Milestone 1 - MVP

- static participant information
- no liveness or leader timeouts
- no economic incentives or stakes
- majority based on participant count
- no cryptography or signatures
- discard blocks for non-current height
- discard votes for unknown blocks
- no timestamps

### Milestone 2 - Cryptography

- simple signature on vote
- aggregated signature on block
- valid timestamp range
- buffer entities for future view or time

### Milestone 2 - Liveness

- timeout mechanism with node-local parameters
- deterministic multi-leader election per round

### Milestone 3 - Incentives

- ledger with staking balances
- secondary state for currency balances
- simple transaction selection mechanism
