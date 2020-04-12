# Alvalor Consensus

## Description

Alvalor consensus implements a byzantine fault-tolerant (BFT) consensus algorithm, based on HotStuff.

## Roadmap

### Milestone 1 - PoC

- event-driven consensus logic
- no liveness mechanism for leader
- no crypto-economic incentive scheme
- no cryptographic primitives
- no verification against state

### Milestone 2 - MVP

- implement buffer component
- implement chain component
- add state extension check
- add finalization of vertices

### Milestone 3 - Cryptography

- implement signature component
- implement verification component
- add identity set for participants
- add signature creation & checking

### Milestone X - Liveness

- add depth concept to vertices
- add timeout mechanism for leader

### Milestone X - Incentives

- add native economic token ledger
- add transaction fee distribution
- add slashing challenges

### Milestone X - Dynamic Validator Set

- add checkpoints / epochs
- add staking / unstaking
- add validator token auctions
- add validator token buybacks
