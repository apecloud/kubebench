# Add Production-Ready Elasticsearch Rally Benchmark

## Goal

Add a production-ready Elasticsearch benchmark to kubebench by integrating Elastic Rally as a first-class benchmark custom resource. This must reach the same maturity bar as existing kubebench benchmarks: full API/controller/job/exporter/generated-manifest/docs/test coverage, not a thin or MVP-only Rally wrapper.

## What I Already Know

* The user wants this benchmark to be based on Elastic Rally: https://github.com/elastic/rally.
* The user clarified that the task must target production readiness and parity with other kubebench benchmarks, not just an MVP.
* The user also wants durable Rally knowledge captured in task documentation so implementation can consult it later without rereading only the GitHub README.
* Kubebench currently models each benchmark as a Kubebuilder API type, reconciler, Job builder, generated CRD/RBAC, sample YAML, docs, and sometimes exporter metrics.
* Existing benchmark Jobs use `/var/log` emptyDir for logs, apply CR labels plus `kubebench.apecloud.io/name` and `kubebench.apecloud.io/type`, and propagate tolerations/resources from `BenchCommon`.
* Existing reconcilers run Jobs sequentially, update `status.succeeded`, `status.total`, `status.completions`, and record pod logs into `status.conditions`.
* Rally's official Docker image supports `benchmark-only` for existing clusters; Docker mode does not support non-`benchmark-only` pipelines or distributed load-driver mode.
* Rally tracks can come from the default track repository, custom track repositories, or local `--track-path`.
* Rally can produce a CSV summary report via `--report-format=csv --report-file=...`, which is the best stable artifact for kubebench status and exporter parsing.

## Requirements

### API

* Add a new namespaced benchmark CRD named `Esrally` in group `benchmark.apecloud.io/v1alpha1`.
* Add `EsrallySpec`, `EsrallyStatus`, `Esrally`, and `EsrallyList` under `api/v1alpha1`, following existing benchmark status fields and embedding `BenchCommon`.
* Add `elasticsearch` as a supported target driver value where target driver validation is maintained.
* Provide Rally-specific spec fields:
  * `track` string, default `geonames`.
  * `trackRepository` string, optional.
  * `trackPath` string, optional alternative to `track`.
  * `challenge` string, optional.
  * `includeTasks` string slice, optional.
  * `trackParams` map of string to string, optional.
  * `targetHosts` string slice, optional; when empty use `spec.target.host:spec.target.port`.
  * `clientOptions` string, optional raw Rally client options.
  * `onError` enum `abort|continue|continue-on-network`, default `abort`.
  * `offline` bool, optional.
  * `testMode` bool, optional.
  * `telemetry` string slice, optional.
  * `telemetryParams` string, optional raw Rally telemetry params or mounted JSON path.
  * `reportFormat` enum `csv|markdown`, default `csv`.
  * `reportFile` string, default `/var/log/esrally-report.csv`.
  * `metrics` bool, default true.
  * A production storage/config surface for Rally home/config: default writable `emptyDir`, plus optional existing PVC claim name for track-data reuse.
* Add validation/default markers so generated CRDs are self-documenting and reject invalid enum values.
* Add print columns matching existing benchmark resources: status, completions, age.

### Constants and Images

* Add constants for `EsrallyType`, `ElasticsearchDriver`, and `KUBEBENCH_ESRALLY_IMAGE`.
* Default Rally image should be configurable by `KUBEBENCH_ESRALLY_IMAGE`.
* Use an official pinned `elastic/rally` tag by default for reproducibility unless maintainers choose an internal mirrored image.
* Update Helm/deployment image environment plumbing wherever other benchmark image env vars are surfaced.

### Precheck

* Add a production Elasticsearch precheck Job for parity with other benchmarks.
* Implement `tools elasticsearch ping` under `pkg/tools`/`cmd/tools`.
* Precheck should call Elasticsearch root or `/_cluster/health` using `spec.target`.
* Support basic auth from `Target.User`/`Target.Password`.
* Keep precheck logs concise and useful.
* Advanced TLS/client-options can be verified by Rally itself if duplicating the full Rally client-options parser in the precheck would add risky complexity; document that boundary.

### Job Construction

* Add `NewEsrallyJobs` and `NewEsrallyRunJobs`, following existing Job builder conventions.
* The Job list must include precheck first, then one Rally run Job.
* Always include `race`.
* Always include `--pipeline=benchmark-only`.
* Always derive or render `--target-hosts`.
* Include track/challenge/include-tasks/track-params/client-options/on-error/offline/test-mode/telemetry/report flags according to spec.
* Write console output to `/var/log/esrally.log`.
* Write summary report to the configured report file, defaulting to `/var/log/esrally-report.csv`.
* Mount `/var/log` for shared benchmark/exporter access.
* Mount a writable `/rally/.rally` Rally home; default to `emptyDir`, with optional PVC-backed reuse for production runs.
* If a Rally config ConfigMap is supported, mount `rally.ini`/`logging.json` into the official image's expected path.
* Apply CR labels, kubebench labels, tolerations, resource requests/limits, owner references, restart policy, and image pull policy consistently with existing benchmark Jobs.
* Add exporter sidecar when `metrics` is true.
* Do not write raw `clientOptions` or passwords into CR status summaries.

### Controller

* Add `EsrallyReconciler`.
* Match existing sequential reconcile/status behavior.
* Status must move through Running, Completed, and Failed consistently.
* `status.total`, `status.succeeded`, `status.completions`, `status.conditions`, and `completionTimestamp` must be maintained.
* Successful run conditions should store a concise Rally result summary, not unbounded raw logs.
* Failed run conditions should retain enough pod log context for troubleshooting.
* Register the reconciler in `cmd/manager/main.go`.

### Exporter and Parsing

* Add production Rally exporter support, not an optional follow-up.
* Add `internal/exporter/esrally.go`.
* Register init and metrics in `internal/exporter/register.go`.
* Dispatch `esrally` in `internal/exporter/scrape.go`.
* Parse CSV reports using `encoding/csv`.
* Publish numeric Rally report rows through a generic gauge:
  `kubebench_esrally_metric_value{benchmark,name,metric,task,unit}`.
* Ignore empty, unavailable, or non-numeric values.
* Preserve enough labels to distinguish track tasks while avoiding raw log text or credentials.
* Add parser/status-summary unit tests and fixtures under `internal/exporter/testdata`.

### Generated and Deployment Artifacts

* Update `PROJECT` resource list.
* Regenerate `api/v1alpha1/zz_generated.deepcopy.go`.
* Regenerate CRDs under `config/crd/bases`.
* Regenerate RBAC roles under `config/rbac`.
* Add Helm CRD for Esrally under `deploy/helm/crds`.
* Update Helm templates/values for any new image env var or RBAC artifact expected by project convention.
* Add `config/samples/benchmark_v1alpha1_esrally.yaml`.
* Add sample to `config/samples/kustomization.yaml`.
* Add docs at `docs/esrally.md`.
* Update `README.md` benchmark table.

### Documentation

* Document Rally concepts enough for future developers and users:
  * race, pipeline, benchmark-only, track, challenge, task, track params, target hosts, client options, report file, metrics, telemetry, offline usage.
* Explain why kubebench uses `benchmark-only`.
* Include basic, auth/TLS, track/challenge, include-tasks, offline/custom-track, and metrics examples.
* Explain `testMode` as smoke-only and not meaningful for benchmark numbers.
* Explain Rally Docker limitations.
* Explain track data size and Rally home persistence behavior.
* Include troubleshooting for missing track data, auth/TLS failures, target not reachable, report file missing, nonzero error rate, and unsupported telemetry.

## Acceptance Criteria

* [ ] `api/v1alpha1.Esrally` and `EsrallyList` are registered in the scheme and have generated DeepCopy methods.
* [ ] `make manifests generate` produces an `esrallies.benchmark.apecloud.io` CRD with expected defaults, enums, status subresource, print columns, and RBAC.
* [ ] Manager registers `EsrallyReconciler`.
* [ ] `tools elasticsearch ping` exists and is used by the precheck Job.
* [ ] A sample `Esrally` resource renders a precheck Job followed by a Rally run Job.
* [ ] The Rally run Job is equivalent to official Docker usage: `elastic/rally race --pipeline=benchmark-only --target-hosts=<hosts> ...`.
* [ ] Job labels, owner references, tolerations, resource requests/limits, image pull policy, restart policy, `/var/log` volume, and Rally home volume match project conventions.
* [ ] Successful runs update phase to `Completed`, update completion counters, set completion timestamp, and record a concise Rally result summary in conditions.
* [ ] Failed runs update phase to `Failed` and record useful failure logs in conditions.
* [ ] Raw credentials/client options are not copied into CR status conditions or metric labels.
* [ ] Exporter parser unit tests cover representative Rally CSV rows, unavailable/non-numeric values, and labels for metric/task/unit.
* [ ] Controller/Job builder tests cover command construction for default, auth, track path, target hosts, offline, telemetry, and extra args scenarios where practical.
* [ ] Existing tests still pass.
* [ ] `docs/esrally.md` is complete enough for users and future developers to operate and extend the benchmark.
* [ ] Helm CRDs and generated config CRDs are consistent.

## Definition of Done

* Go code is formatted with `gofmt`/`go fmt`.
* `make manifests` and `make generate` have been run after API/marker changes.
* Relevant unit tests pass, including exporter parser tests, status parser tests, tools precheck tests where feasible, and Job builder tests.
* `go test ./...` or the closest feasible project test command has been run; any environment limitation is documented.
* Generated CRDs and Helm CRDs are consistent with API markers.
* Documentation and sample YAML match the implemented API exactly.
* The final worktree does not leave generated manifests stale.

## Technical Approach

Implement this as a new benchmark kind following the existing kubebench pattern rather than adding Elasticsearch-specific branches to an existing benchmark kind.

Resource name: `Esrally`, because current kubebench resources are mostly tool-centric (`Pgbench`, `Sysbench`, `Redisbench`) and Rally is the Elasticsearch-specific benchmark tool being exposed.

Production command shape:

```sh
race \
  --pipeline=benchmark-only \
  --target-hosts="${TARGET_HOSTS}" \
  --track="${TRACK}" \
  --challenge="${CHALLENGE}" \
  --include-tasks="${TASKS}" \
  --track-params="${TRACK_PARAMS}" \
  --client-options="${CLIENT_OPTIONS}" \
  --on-error="${ON_ERROR}" \
  --report-format=csv \
  --report-file=/var/log/esrally-report.csv
```

The official Rally Docker image entrypoint already invokes `esrally`, so Kubernetes container args should start with `race`. If kubebench chooses a custom image without that entrypoint, document and implement the alternate command explicitly.

Use the Rally CSV report for stable parsing. The CR condition parser should summarize high-value rows, while the exporter should publish numeric rows generically with labels.

## Decision (ADR-lite)

**Context**: Kubebench needs production-quality Elasticsearch benchmark support. Rally is the upstream, purpose-built macrobenchmarking framework for Elasticsearch, and the operator runs benchmark workloads as Kubernetes Jobs against existing services.

**Decision**: Add a dedicated `Esrally` custom resource that runs Rally's Docker `benchmark-only` mode against `spec.target` or `spec.targetHosts`, includes an Elasticsearch precheck, stores console/report artifacts under `/var/log`, uses a writable Rally home, summarizes results in status conditions, and exposes Prometheus metrics through a CSV parser.

**Consequences**:

* The implementation works with existing Elasticsearch clusters and aligns with Rally Docker constraints.
* Because `benchmark-only` does not provision the cluster, users remain responsible for cluster state/reproducibility.
* Track data can be large, so docs and Rally home persistence are important production concerns.
* Authentication/TLS support remains flexible through Rally `clientOptions`, but implementation must not echo secrets into status or metrics.
* A generic metric gauge keeps exporter implementation maintainable despite Rally's dynamic metric names and tasks.

## Out of Scope

* Starting or tearing down Elasticsearch clusters with Rally.
* Rally daemon mode or distributed load-driver mode.
* Non-`benchmark-only` pipelines with the official Docker image.
* Managing a dedicated Rally metrics Elasticsearch store as a kubebench-managed dependency.
* Full secret-ref redesign across all kubebench benchmarks. Esrally may add narrowly scoped config/secret support if it fits project conventions, but this task should not refactor all benchmark credential handling.
* Building or maintaining custom Rally tracks inside this task, beyond supporting built-in tracks, custom repositories/config, and `trackPath`.

## Technical Notes

* Research reference: [`research/rally-integration-notes.md`](research/rally-integration-notes.md).
* Relevant existing patterns:
  * `api/v1alpha1/redisbench_types.go`
  * `api/v1alpha1/sysbench_types.go`
  * `internal/controller/redisbench_controller.go`
  * `internal/controller/redisbench_job.go`
  * `internal/controller/sysbench_job.go`
  * `internal/exporter/sysbench.go`
  * `internal/exporter/register.go`
  * `internal/exporter/scrape.go`
  * `cmd/manager/main.go`
  * `cmd/exporter/main.go`
  * `cmd/tools/main.go`
  * `pkg/constants/constants.go`
  * `pkg/constants/env.go`
  * `internal/utils/job.go`
* Relevant Trellis specs:
  * `.trellis/spec/backend/crd-api-guidelines.md`
  * `.trellis/spec/backend/controller-reconcile-guidelines.md`
  * `.trellis/spec/backend/job-construction-guidelines.md`
  * `.trellis/spec/backend/exporter-metrics-guidelines.md`
  * `.trellis/spec/backend/config-deployment-guidelines.md`
  * `.trellis/spec/backend/cli-tools-guidelines.md`
  * `.trellis/spec/backend/quality-guidelines.md`

## Implementation Plan

1. API and scaffold
   * Add `api/v1alpha1/esrally_types.go`.
   * Register types in scheme.
   * Update constants/env defaults and target driver enum.
   * Update `PROJECT`.
   * Run code generation.
2. Elasticsearch precheck
   * Add `pkg/tools/elasticsearch.go`.
   * Register `tools elasticsearch ping`.
   * Route `NewPreCheckJob` for `ElasticsearchDriver`.
   * Add focused tests where practical.
3. Controller and Job builder
   * Add `internal/controller/esrally_job.go`.
   * Add `internal/controller/esrally_controller.go`.
   * Register reconciler in `cmd/manager/main.go`.
   * Add tests for command construction, labels/resources, precheck ordering, target hosts, auth/client options, and optional flags.
4. Exporter and status parser
   * Add `internal/exporter/esrally.go`.
   * Update exporter registration and scrape dispatch.
   * Add parser fixtures/tests for Rally CSV output.
   * Add `ParseEsrally` for concise CR status output.
5. Manifests, samples, docs
   * Regenerate CRDs/RBAC/DeepCopy.
   * Copy/generated Helm CRD according to project convention.
   * Add sample YAML and docs.
   * Update README benchmark table.
6. Verification
   * Run `make manifests generate fmt vet`.
   * Run relevant `go test` packages, preferably `go test ./...`.
   * Inspect generated CRD schema for expected defaults/enums.

## Implementation Defaults

* Rally image: use a pinned official `elastic/rally:<stable-version>` tag and keep it overridable through `KUBEBENCH_ESRALLY_IMAGE`. The implementer must verify the current official Docker tag before coding, record the exact tag in docs/sample output, and avoid `latest` as the default.
* Rally home persistence: default to `emptyDir` for compatibility with existing benchmark Jobs, and add an optional existing PVC claim name for production track-cache reuse.
* Credentials: support existing `Target.User`/`Target.Password` conventions and raw `clientOptions`; do not introduce a broad cross-benchmark credential refactor in this task. Do not copy secrets into status, metric labels, or generated summaries.
