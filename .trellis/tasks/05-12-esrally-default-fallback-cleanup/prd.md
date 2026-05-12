# Clean Up Esrally Default Handling And Fallback Logic

## Problem

Some Esrally helper logic compensates for Kubernetes API defaults not being applied when tests construct Go objects directly. The clearest example is:

```go
return cr.Spec.Metrics || cr.Spec.Metrics == false && cr.Spec.ReportFormat == ""
```

This makes explicit `metrics: false` indistinguishable from an omitted boolean in direct Go object tests when `reportFormat` is empty. It is understandable as a compatibility fallback, but it is brittle and can hide real behavior.

Relevant review references:

- `internal/controller/esrally_job.go`
- `api/v1alpha1/esrally_types.go`
- `internal/controller/esrally_job_test.go`

## Goal

Make Esrally defaults explicit, readable, and testable without relying on surprising boolean fallback expressions.

## Requirements

- Review Esrally helper defaults for track, onError, reportFormat, reportFile, metrics, and clientOptions.
- Make direct-Go-object defaulting behavior explicit in a small helper if needed.
- Avoid ambiguous `bool` default logic where omitted and false need different semantics; consider pointer bool only if API compatibility allows it.
- Keep CRD defaults and Job builder defaults consistent.
- Add tests that distinguish omitted defaults from explicit values where possible.
- Ensure cleanup does not regress existing sample YAML or generated CRD behavior.

## Acceptance Criteria

- Esrally defaulting logic is straightforward and covered by unit tests.
- Explicit metrics-disabled cases behave as expected in direct unit tests and CRD-defaulted runtime objects.
- No surprising fallback silently enables a sidecar after the user disables metrics.
- Generated CRDs remain consistent after any API marker changes.
- `make generate manifests` produces no unexpected drift after the change.

