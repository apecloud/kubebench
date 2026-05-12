# Harden Esrally Metrics And Report Contract

## Problem

The current API allows `reportFormat: markdown` while `metrics` defaults to true. The exporter sidecar is still added, but it only parses Rally CSV reports, so metrics can be silently absent.

The current API also allows arbitrary `reportFile` values even though the workload and exporter only share the `/var/log` volume. If users point `reportFile` elsewhere, the run container may write a report the exporter cannot read.

Relevant review references:

- `api/v1alpha1/esrally_types.go`
- `internal/controller/esrally_job.go`
- `internal/exporter/esrally.go`
- `docs/esrally.md`

## Goal

Make metrics behavior explicit and hard to misconfigure. Users should either get working metrics or an obvious validation/runtime signal explaining why metrics are unavailable.

## Requirements

- Define the supported `metrics`, `reportFormat`, and `reportFile` combinations.
- Ensure `metrics: true` requires a CSV report path that the exporter can read, preferably under `/var/log`.
- Avoid silently starting an exporter that can never parse the configured report.
- Consider API validation markers for report path/format if they fit Kubernetes CRD constraints.
- If non-CSV reports remain supported, document that metrics are unavailable and ensure the Job behavior reflects that.
- Ensure status summaries and docs match the final contract.

## Acceptance Criteria

- `metrics: true` with CSV report under the shared log volume works.
- `metrics: true` with markdown does not silently succeed without metrics; it is rejected, normalized, or documented with explicit Job behavior.
- `metrics: true` with a report file outside the shared volume does not silently lose metrics.
- Job builder tests cover default metrics, disabled metrics, markdown format, and non-default report paths.
- Exporter tests cover the chosen behavior for unavailable/non-CSV reports.
- `docs/esrally.md`, sample YAML, and CRD markers match the implemented contract.

