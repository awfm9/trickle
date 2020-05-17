# Simple BFT Consensus

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/awfm/consensus) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/awfm/consensus/master/LICENSE) [![Build Status](https://travis-ci.org/awfm/consensus.svg?branch=master)](https://travis-ci.org/awfm/consensus)

The consensus package provides a simple event-driven and stateless harness for BFT consensus algorithms.

## Roadmap

### Milestone 1 - POC

The first milestone is about standing up the core consensus logic. It should be able to keep a number of instances of the consensus algorithm synchronized in integration tests. Dependencies should be provided as mocks, simulating the full real behaviour where needed, or simply providing naive placeholder functionality. The integration harness should initialize a consensus participant with the core business logic and the related mocked dependencies.

#### Version 0.1.0: integration test

- [x] implement processor component
- [x] implement simulated network
- [x] implement simulated cache
- [x] implement simulated state
- [x] implement placeholder builder
- [x] implement placeholder signer
- [x] implement placeholder verifier
- [x] implement integration harness

### Milestone 2 - MVP

The second milestone should deliver a fully working consensus algorithm that is robust in a non-hostile environment. No liveness guarantees are provided, the leader selection is predictable, the participant set is immutable and there are no stakes.

#### Version 0.2.0: state verification

- [x] add strategy interface
- [x] add graph interface
- [x] implement cache components
- [x] implement simulated graph
- [x] implement naive selection strategy
- [x] remove state interface

#### Version 0.2.1: rich state

- [ ] implement graph component

#### Version 0.2.2: consensus committee

- [ ] add committee interface
- [ ] implement commitee component
- [ ] implement signer component
- [ ] implement verifier component

### Milestone 3 - Alpha

The third milestone is all about implementing the missing parts needed for a production system. We need to add liveness to the system, we need to include a reliable and cryptographically secure source of randomness and we need a ledger to create the necessary economic incentives.

#### Version 0.3.0: theoretical liveness

- [ ] add vertex depth concept
- [ ] add leader timeout mechanism

#### Version 0.3.1: true randomness

- [ ] add distributed key generation
- [ ] add vertex theshold signatures
- [ ] implement standard selection strategy

#### Version 0.3.2: economic incentives

- [ ] add native token
- [ ] add transaction payloads
- [ ] add validator rewards
- [ ] add slashing challenges

### Milestone 4 - Beta

The fourth milestone is about adding features to set our consensus algorithm apart. It should contain the full feature set of what we will release.

#### Version 0.4.0: validator modifications

- [ ] add staking mechanism
- [ ] add unstaking mechanism
- [ ] add checkpoint mechanism

#### Version 0.4.1: validator voting

- [ ] add voting mechanism
- [ ] add stake delegation

#### Version 0.4.2: advanced cryptoeconomics

- [ ] add reward pool
- [ ] add distribution scheme
