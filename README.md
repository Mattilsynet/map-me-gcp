## mapMeGcp
Responsible for reading map-managed-gcp-environment bucket and setting up a cloudrun job for each environment. The cloudrunjob in-turn is responsible for how the google project is set-up according to its environment variables. map-me-gcp also needs to know which service-account to use for the running cloudrunjob which today is hard-coded. 

## Pre-requisits

1. wash-cli  
2. wit-deps  
3. wit-bindgen-go (wash-tools)  
4. tinygo >= 0.33.0

A solution is to install via `cargo` and `cargo-binstall`

1. `cargo install binstall`
2. `cargo binstall wash-cli`
3. `cargo binstall wasm-tools`
4. `cargo binstall wit-deps-cli`
5. `go install go.bytecodealliance.org/cmd/wit-bindgen-go@latest` 

## First build
1. `wash build`

## First run
1. wash up in one terminal (hereby refered to as terminal 1)  
2. `NATS_URL=nats://localhost:4222 nats kv map-managed-gcp-environment`  (terminal 2)  
3. `wash app deploy local.wadm.yaml` (terminal 2)  
4. `wash app list` (terminal 2)  
5. `wash app status mapMeGcp` (terminal 2)  

## Development
1. Open component.go (some examples are shown on how to register request/reply, subscription or consumer depending on your choices for capabilities)  
2. Use logger for logging to wash  
3. Use nats variable for pub/sub or req/reply  
4. Use js variable for jetstream publish  
5. Use kv variable for key value store

## Problems?
1. Instead of `wash up`, run `WASMTIME_BACKTRACE_DETAILS=1 RUST_LOG=debug wash up`  
2. Make an issue at github.com/Mattilsynet/map-me-gcp


