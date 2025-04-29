## mapMeGcp

## Description
This wasmcloud component is responsible for ...

## Disclaimer
1. Ment for public repositories

## Pre-requisits

1. wash-cli  
2. wit-deps  
3. wit-bindgen-go (wash-tools)  

A solution is to install `cargo` and `cargo-binstall`

1. cargo install binstall
2. cargo binstall wash-cli
3. cargo binstall wasm-tools
4. cargo binstall wit-deps-cli
5. go install go.bytecodealliance.org/cmd/wit-bindgen-go@latest


## First build
1. `wit-deps`  
2. `mkdir gen`  
3. `wit-bindgen-go generate --world map-me-gcp --out gen ./wit`  
4. `go mod tidy`  
5. wash build

## First run
1. wash up in one terminal (hereby refered to as terminal 1)  
2. NATS_URL=nats://localhost:4222 nats kv add my-bucket  (terminal 2)  
3. wash app deploy local.wadm.yaml (terminal 2)  
4. wash app list (terminal 2)  
5. wash app status mapMeGcp (terminal 2)  

## Development
1. Open component.go (some examples are shown on how to register request/reply, subscription or consumer depending on your choices for capabilities)  
2. Use logger for logging to wash  
3. Use nats variable for pub/sub or req/reply  
4. Use js variable for jetstream publish  
5. Use kv variable for key value store

## Problems?
1. Instead of `wash up`, run `WASMTIME_BACKTRACE_DETAILS=1 RUST_LOG=debug wash up`  
2. Make an issue at github.com/Mattilsynet/map-cli  

## First time git tag push, do this before any github actions run
1. Goto your github repository
2. Goto packages
3. Goto settings
4. Enable this repository itself to modify packages

## CI/CD
Uses github actions for CI/CD, check `.github/workflows`
## Authors

