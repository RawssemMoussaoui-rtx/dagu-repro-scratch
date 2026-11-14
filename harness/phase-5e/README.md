# Phase 5e synthetic harness

## DTrace

```sh
dtrace -q -s dtrace/trace.d -p <pid>
```

Output lines use: `target=<pid> event=<name> ... fd=<n> gen=<n> valid=<0|1>`.

## FD calibration

```sh
cd fd-calibration
GOTOOLCHAIN=go1.27.3 go build ./...
./fd-calibration
```

## Security check

```sh
C:\go\bin\gosec.exe -fmt=text ./fd-calibration/...
```

## Fixtures

`artifact-fixture/` contains sanitized transport-test inputs and `SHA256SUMS`.