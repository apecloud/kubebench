# Fix Esrally Precheck For Production Auth And TLS

## Problem

The current Esrally controller always creates an Elasticsearch precheck Job before the Rally run Job, but the precheck only uses `spec.target.host`, `spec.target.port`, `spec.target.user`, and `spec.target.password` over HTTP/basic auth.

This blocks valid production Rally configurations before Rally can run:

- HTTPS clusters configured through `clientOptions`.
- API key or certificate based auth configured through `clientOptions`.
- Multi-host or proxy/prefix configurations provided through `targetHosts`.
- Any advanced Rally client behavior intentionally delegated to `--client-options`.

Relevant review references:

- `internal/controller/esrally_job.go`
- `internal/utils/job.go`
- `pkg/tools/elasticsearch.go`
- `.trellis/tasks/05-12-elasticsearch-rally-benchmark/prd.md`
- `.trellis/tasks/05-12-elasticsearch-rally-benchmark/research/rally-integration-notes.md`

## Goal

Make the precheck useful without rejecting production-valid Rally configurations that only Rally itself can fully validate.

## Requirements

- Decide and implement a production-safe precheck contract:
  - Either make precheck understand enough Esrally fields to reach the same endpoint Rally will use, or
  - allow advanced `clientOptions`/TLS/API-key cases to skip or soften precheck intentionally while preserving useful diagnostics.
- If precheck remains mandatory, it must support HTTPS and target host selection for common production Elasticsearch endpoints.
- If precheck is bypassable, the bypass must be explicit, documented, and tested.
- Avoid parsing full Rally `clientOptions` if doing so would be fragile; document the boundary clearly.
- Do not log passwords, API keys, full client options, or generated credential-bearing command strings.
- Keep parity with existing benchmark precheck behavior where practical.

## Acceptance Criteria

- Esrally against a basic HTTP Elasticsearch target still creates and passes a precheck Job.
- Esrally configured for HTTPS/TLS or API-key-style `clientOptions` is not blocked by an HTTP-only basic-auth precheck.
- `targetHosts` behavior is considered in the precheck contract.
- New or updated tests cover basic HTTP, HTTPS/advanced client-options behavior, and any explicit skip/soft-fail mode.
- Docs explain what precheck validates and what Rally validates during the run Job.
- `go test ./internal/controller ./pkg/tools` passes.

