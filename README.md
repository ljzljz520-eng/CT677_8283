# Vocational Training Sync

This Go 1.22 command-line application validates vocational training records, stores them in memory, uploads valid records to an industry platform asynchronously, and exposes task progress, correction CSV output, and ordered task logs.

The project has no external services or third-party dependencies. Record IDs, task IDs, fixtures, clocks, and synchronization are deterministic.

## Requirements

- Go 1.22.12
- `GOTOOLCHAIN=local`
- `CGO_ENABLED=0` for static builds

## Run

From the module root:

```sh
GOTOOLCHAIN=local go run ./cmd/training-sync
```

The command reads `fixtures/enrollments.csv`, prints invalid rows as a correction CSV template, waits for the asynchronous upload task, then prints progress and task logs.

## Test

```sh
GOTOOLCHAIN=local go test -count=1 ./...
```

The suite covers fixture parsing, validation, correction output, successful asynchronous upload, progress, logs, and cancellation after a platform rejection. The cancellation scenario currently contains the intentionally injected regression requested for this project, so its business-chain test fails deterministically.

## Build

```sh
GOTOOLCHAIN=local CGO_ENABLED=0 go build ./...
```

The source uses only the Go standard library and supports `linux/amd64` and `linux/arm64`.
