# Alvalor Consensus

[![Build Status](https://travis-ci.com/alvalor/consensus.svg?token=Hm9YEiz4aAfKiFLFh2sr&branch=master)](https://travis-ci.com/alvalor/consensus)

## Description

Alvalor consensus implements a byzantine fault-tolerant (BFT) consensus algorithm, based on HotStuff.

## Roadmap

### Milestone 1 - PoC

Version 0.0.1

- [x] event-driven consensus logic

### Milestone 2 - MVP

Version 0.0.2

- [x] implement cache components
- [ ] implement chain component
- [ ] add state extension check
- [ ] add vertex confirmation

### Milestone 3 - Cryptography

Version 0.0.3

- [ ] implement signature component
- [ ] implement verification component
- [ ] add participants identity set
- [ ] add signature creation & checking

### Milestone 4 - Randomness

Version 0.0.4

- [ ] add entropy key setup
- [ ] add entropy vote signatures
- [ ] add vertex random source

### Milestone 5 - Liveness

Version 0.1.0

- [ ] add vertex depth concept
- [ ] add leader timeout mechanism

### Milestone X - Incentives

- [ ] add native token ledger
- [ ] add direct reward structure
- [ ] add slashing challenges

### Milestone X - Committee

- [ ] add validator tokens
- [ ] add checkpoints / epochs
- [ ] add staking / unstaking

### Milestone X - Delegation

- [ ] add stake delegation

### Milestone X - Auctions

- [ ] add validator reward pool
- [ ] add native token auctions
- [ ] add native token buybacks
- [ ] add reward distribution scheme
