# Alvalor Consensus

[![Build Status](https://travis-ci.com/alvalor/consensus.svg?token=Hm9YEiz4aAfKiFLFh2sr&branch=master)](https://travis-ci.com/alvalor/consensus)

## Description

Alvalor consensus implements a byzantine fault-tolerant (BFT) consensus algorithm, based on HotStuff.

## Roadmap

### Milestone 1 - PoC

Version 0.0.1

- event-driven consensus logic
- no liveness mechanism for leader
- no crypto-economic incentive scheme
- no cryptographic primitives
- no verification against graph

### Milestone 2 - MVP

Version 0.0.2

- implement buffer component
- implement chain component
- add state extension check
- add finalization of vertices

### Milestone 3 - Cryptography

Version 0.0.3

- implement signature component
- implement verification component
- add identity set for participants
- add signature creation & checking

### Milestone 4 - Randomness

Version 0.0.4

- add threshold key share generation
- add threshold signature shares to votes
- add threshold signature to vertices

### Milestone 5 - Liveness

Version 0.1.0

- add depth concept to vertices
- add timeout mechanism for leader

### Milestone X - Incentives

Version 0.2.0

- add native economic token ledger
- add transaction fee distribution
- add slashing challenges

### Milestone X - Committee

Version 0.3.0

- add checkpoints / epochs
- add staking / unstaking
- add validator token auctions
- add validator token buybacks

### Miscellaneous

Version x.x.x

- add stake delegation
