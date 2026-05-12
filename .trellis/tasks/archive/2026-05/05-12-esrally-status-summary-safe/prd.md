# Make Esrally Status Summaries Concise And Secret-Safe

## Problem

`SummarizeEsrallyCSV` currently falls back to returning the original log message when CSV parsing finds no numeric metrics. The Esrally controller writes that return value to successful status conditions.

That means a successful run with a missing or unparsable report can store noisy raw output in status instead of a concise Rally result summary. It also increases the chance of accidentally surfacing sensitive values printed by Rally or shell output.

Relevant review references:

- `internal/exporter/esrally.go`
- `internal/controller/esrally_controller.go`
- `internal/utils/job.go`
- `.trellis/spec/backend/controller-reconcile-guidelines.md`
- `.trellis/spec/backend/exporter-metrics-guidelines.md`

## Goal

Successful Esrally status conditions should always be bounded, useful, and safe. Raw logs should remain available through pod logs and only be stored in failed conditions when needed for troubleshooting.

## Requirements

- Replace raw-log fallback for successful Esrally summaries with a bounded explanatory message.
- Prefer high-value Rally CSV rows when available.
- If no parseable CSV metrics exist, condition text should say that no numeric Rally CSV summary was found and point users to pod logs/report file.
- Do not include raw `clientOptions`, passwords, API keys, or connection strings in status.
- Keep failed Job behavior useful for troubleshooting, while still relying on existing log trimming.

## Acceptance Criteria

- Successful Esrally condition contains concise parsed metrics when CSV is valid.
- Successful Esrally condition does not copy full logs when CSV is missing, markdown, empty, or unparsable.
- Tests cover valid CSV, missing CSV marker, unparsable CSV, and log text containing credential-like strings.
- Failed Esrally Jobs still record useful trimmed logs.
- `go test ./internal/exporter ./internal/controller` passes.

