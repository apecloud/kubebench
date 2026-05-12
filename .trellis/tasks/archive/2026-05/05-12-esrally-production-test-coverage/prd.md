# Expand Esrally Production Test Coverage

## Problem

The current Esrally tests cover the happy path and a few command options, but the PRD called for production-level maturity and parity with other benchmarks. Several important scenarios are not covered:

- `telemetry` and `telemetryParams`.
- `extraArgs`.
- `metrics: false`.
- `reportFormat` and `reportFile` edge cases.
- realistic Rally CSV shapes, including additional columns such as `Lap`.
- exporter done-file behavior.
- Prometheus label/value updates.
- controller status behavior around completion and failure.

Relevant review references:

- `internal/controller/esrally_job_test.go`
- `internal/exporter/esrally_test.go`
- `internal/exporter/testdata/esrally.csv`
- `.trellis/spec/backend/quality-guidelines.md`

## Goal

Raise Esrally coverage to the same production-readiness bar expected by the original task PRD, with focused tests around the highest-risk behavior.

## Requirements

- Add table-driven Job builder tests for default, auth, targetHosts, trackPath, PVC, offline, testMode, telemetry, extraArgs, metrics false, and report settings.
- Use a more realistic Rally CSV fixture, including extra columns when Rally emits them.
- Test parser behavior for numeric rows, empty values, `N/A`, non-numeric values, task/unit labels, and missing headers.
- Test `ScrapeEsrally` done-file behavior without slow sleeps where feasible.
- Test metric update labels or isolate registry setup to avoid duplicate collector panics.
- Add controller/status tests if a lightweight fake-client pattern is practical; otherwise document why Job builder/parser coverage is the appropriate scope.

## Acceptance Criteria

- Esrally parser tests use representative Rally CSV data, not only a minimal synthetic CSV.
- Esrally Job tests fail if command arguments, volume mounts, exporter args, or labels drift.
- Tests cover the metrics/report contract chosen by the related task.
- `go test ./internal/controller ./internal/exporter ./pkg/tools` passes.
- `go test ./...` passes.

