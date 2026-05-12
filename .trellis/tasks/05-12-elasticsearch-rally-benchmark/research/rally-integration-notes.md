# Rally Integration Notes

Date: 2026-05-12

This note is intentionally broader than a README summary. It is meant to be a developer reference for implementing a production-ready Elasticsearch benchmark in kubebench using Elastic Rally.

## Primary References

* Rally repository: https://github.com/elastic/rally
  * Rally is Elastic's macrobenchmarking framework for Elasticsearch.
  * The project manages benchmark data/specifications, runs races, records results, supports telemetry, and compares results.
  * The repository includes the Python implementation, Docker packaging, docs sources, tests, release scripts, and benchmark-related code.
* Rally Docker docs: https://esrally.readthedocs.io/en/stable/docker.html
* Rally command line reference: https://esrally.readthedocs.io/en/stable/command_line_reference.html
* Rally pipelines: https://esrally.readthedocs.io/en/stable/pipelines.html
* Rally track reference: https://esrally.readthedocs.io/en/stable/track.html
* Rally metrics reference: https://esrally.readthedocs.io/en/stable/metrics.html
* Rally summary report: https://esrally.readthedocs.io/en/stable/summary_report.html
* Rally offline usage: https://esrally.readthedocs.io/en/stable/offline.html
* Rally telemetry devices: https://esrally.readthedocs.io/en/stable/telemetry.html
* Default Rally tracks repository: https://github.com/elastic/rally-tracks

## Rally Concepts Developers Need

### Race

A race is one benchmark execution. The `race` subcommand is the core command kubebench should run. A race binds together:

* pipeline: how Rally gets to a benchmarkable cluster
* target hosts: the Elasticsearch endpoints
* track: the workload definition
* challenge: the specific scenario within a track
* task filters: optional subset of tasks
* track params: parameter values substituted into templated tracks
* client options: TLS/authentication/connection behavior
* telemetry and reporting configuration

### Pipeline

Rally pipelines define setup/teardown around benchmarking, not the benchmark workload itself.

Relevant pipelines:

* `benchmark-only`: assumes Elasticsearch already exists, runs a benchmark, reports results.
* `from-distribution`: downloads and provisions Elasticsearch, then runs the benchmark.
* `from-sources`: builds/provisions Elasticsearch from source, then runs the benchmark.

Kubebench should use `benchmark-only` because kubebench benchmarks databases already running in Kubernetes. The official Rally Docker image explicitly does not support pipelines other than `benchmark-only`, so production code should not expose arbitrary pipeline selection unless kubebench owns a custom non-Docker Rally runtime.

Important consequence: with `benchmark-only`, Rally did not provision the target cluster, so it cannot collect every setup-level metric and reproducibility depends on the user-provisioned cluster state.

### Docker Runtime

The official image is `elastic/rally`.

Docker-specific facts:

* The image entrypoint already wraps `esrally`; Docker examples call `elastic/rally race ...`, not `elastic/rally esrally race ...`.
* Docker mode supports regular Rally commands subject to Docker limitations.
* Unsupported in Docker: distributed load-driver mode and non-`benchmark-only` pipelines.
* `/rally/.rally` is Rally's home/config/cache/result area in the container.
* Rally recommends persisting `/rally/.rally` so downloaded tracks/data and race metadata can be reused.
* The image runs as uid `1000` with gid `0`, which matters for PVC permissions and OpenShift-like clusters.

Kubebench implication:

* The run container command/args should be compatible with the official image entrypoint.
* A production-ready resource should provide a writable `/rally/.rally` volume and documented storage behavior.
* Use `emptyDir` as a safe default, and support an optional PVC-backed Rally home for large tracks/repeated runs.

### Track

A track is the workload definition. It describes one or more benchmark scenarios and may include:

* indices or data streams
* index templates, composable templates, component templates
* corpora/data files
* operations
* schedule
* challenge(s)
* dependencies

Track source options:

* Built-in/default track repository: selected with `--track=<name>`.
* Custom track repository: configure a repository and use `--track-repository=<name>`.
* Ad-hoc local track path: use `--track-path=<file-or-directory>`.

Track repository branch selection is tied to Elasticsearch version compatibility. Rally chooses the most appropriate track branch for the target version, falling back through exact version, major/minor, nearest prior minor, then major branch. This matters for production docs and upgrade benchmarking.

Kubebench implication:

* Support built-in tracks with `track`.
* Support custom repositories via a configurable `rally.ini` or documented `trackRepository`.
* Support local/custom image tracks via `trackPath`.
* Support `trackParams` because many serious Rally tracks are parameterized.
* Provide `testMode` for smoke tests but clearly document that numbers from test mode are not meaningful.

### Challenge and Tasks

A track can contain multiple challenges. A challenge is the scenario schedule: for example, append-only indexing, mixed indexing/search, update-heavy workload, or a query-only workload.

Rally can filter tasks inside a challenge with `--include-tasks`. This is useful when a track contains setup/index/search phases but users want only part of the schedule.

Kubebench implication:

* Support `challenge`.
* Support `includeTasks`.
* Status and Prometheus labels must include challenge/task when available.

### Target Hosts

For `benchmark-only`, Rally needs `--target-hosts`. It accepts comma-separated `host:port` entries and supports optional URL prefix paths such as `host:port/path`.

Kubebench implication:

* Reuse `BenchCommon.target` for parity with existing benches.
* Add an optional `targetHosts []string` for production ES clusters that should be hit through several nodes or a prefixed proxy endpoint.
* If `targetHosts` is omitted, render `spec.target.host:spec.target.port`.

### Client Options, Auth, and TLS

`--client-options` configures Rally's Elasticsearch client.

Supported usage patterns include:

* API key auth: `api_key:'...'`
* Basic auth: `basic_auth_user:'user',basic_auth_password:'password'`
* HTTP compression: `http_compress:true`
* TLS: `use_ssl:true`
* Certificate verification: `verify_certs:true|false`
* CA path: `ca_certs:'/path/to/cacert.pem'`
* Timeout and other Python Elasticsearch client options

Rally's command line parser for basic auth is simple; quotes, commas, and colons in username/password are problematic.

Kubebench implication:

* Provide a raw `clientOptions` escape hatch.
* When `clientOptions` is empty and `spec.target.user/password` are set, synthesize basic auth client options.
* Avoid echoing client options into status conditions, logs generated by kubebench, or metric labels.
* Document YAML quoting carefully.
* For production robustness, prefer mounting certificates/config files and passing paths via `clientOptions`.

### Error Strategy

Rally supports `--on-error`:

* `continue`: record errors; final report includes error rate.
* `abort`: stop on first request error.
* `continue-on-network`: continue through transient network errors, but can hide a down target.

Kubebench implication:

* Expose `onError` enum with a conservative default of `abort` for CI/smoke correctness or `continue` for benchmark completeness. The PRD recommends default `abort` because kubebench Job status should fail loudly when the benchmark target is unhealthy.
* Document when to use `continue` for long-running workloads where error rate is the measurement.

### Reporting and Summary Metrics

Rally prints a summary report at the end of each race. `--report-format` can be `markdown` or `csv`; `--report-file` writes the report to a file.

Important summary report metrics:

* throughput: min/mean/median/max per task
* latency: percentile response latency including wait time
* service time: percentile request processing time excluding wait time
* processing time: client-side processing-inclusive timing when enabled
* error rate: ratio of failed responses
* index stats: indexing, merge, refresh, flush, segment, store/dataset/translog size
* ingest pipeline stats when present
* disk usage metrics when `disk-usage-stats` telemetry is enabled

Kubebench implication:

* CSV report should be the canonical parser input.
* Console log is still needed for troubleshooting and CR conditions.
* Exporter should publish numeric CSV rows generically, not hardcode every possible metric into separate Go variables.
* CR status should summarize high-value rows and point to pod logs for full output.

### Metrics Store

Rally can keep metrics in memory or write detailed records to a dedicated Elasticsearch metrics store. It should not write these records to the target cluster being benchmarked unless users intentionally configure that.

Metric records include race metadata such as race timestamp/id, track, challenge, car, sample type, metric name/value/unit, task, operation, operation type, and meta fields.

Kubebench implication:

* The production kubebench integration should not require a Rally metrics store.
* It can expose a raw `rallyIniConfigMap` or documented configuration path for advanced users who want Rally's native metrics store.
* Built-in kubebench metrics should come from CSV parsing to avoid managing a second Elasticsearch dependency.

### Telemetry

Telemetry devices add extra insight but may skew results.

For `benchmark-only`, only runtime-level telemetry devices with `-stats` suffix are supported. Setup-level telemetry such as `jit`, `gc`, `jfr`, and `heapdump` requires Rally-provisioned clusters and therefore should not be enabled for kubebench's official Docker flow.

Kubebench implication:

* Expose `telemetry []string` and `telemetryParams` only for runtime-compatible devices, or document that unsupported telemetry will fail.
* Production default should be no telemetry to avoid changing benchmark behavior.
* Consider safe examples: `node-stats`, `segment-stats`, `ingest-pipeline-stats`, `disk-usage-stats`.

### Offline Usage

Rally may update tracks/teams repositories unless `--offline` is passed. Built-in tracks normally need network access to download track data unless the data has been preloaded.

Offline flow:

* Download track data on a connected machine.
* Copy the track data archive into Rally's home for the user/container.
* Run with `--offline`.

Kubebench implication:

* Support `offline`.
* Support PVC-backed `/rally/.rally` and `trackPath` for preloaded tracks/data.
* Docs should explain that `offline=true` does not magically provide missing data.

## Production-Ready Kubebench Scope

The Esrally benchmark should be as complete as existing kubebench benchmarks, not a minimal wrapper.

Required surfaces:

* API type and status.
* Generated DeepCopy.
* CRD schema/defaults/enums/print columns.
* Controller registration.
* Sequential reconcile/status/conditions.
* Job builder with labels, owner references, tolerations, resources, image env constant, log and Rally home volumes.
* Elasticsearch precheck Job.
* Exporter sidecar and Prometheus metrics.
* Parser unit tests and fixtures.
* Sample YAML.
* Docs page and README benchmark table.
* Kustomize/RBAC/Helm CRD artifacts.
* Verification commands and documented limitations.

## Recommended API Shape

Resource kind: `Esrally`.

Core fields:

* `track string`
* `trackRepository string`
* `trackPath string`
* `challenge string`
* `includeTasks []string`
* `trackParams map[string]string`
* `targetHosts []string`
* `clientOptions string`
* `onError string`
* `offline bool`
* `testMode bool`
* `telemetry []string`
* `telemetryParams string`
* `reportFormat string`
* `reportFile string`
* `metrics bool`
* `rallyHomeVolume` / `rallyHomePVC` or an equivalent volume configuration
* `rallyConfigConfigMap` optional, for custom `rally.ini` / `logging.json`
* embedded `BenchCommon`

## Recommended Command Shape

Using the official Docker image entrypoint:

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

Add optional flags only when configured:

* `--track-repository`
* `--track-path`
* `--offline`
* `--test-mode`
* `--telemetry`
* `--telemetry-params`
* user `extraArgs`

Implementation note: prefer env vars and argument arrays over shell string interpolation when possible. If the existing project pattern uses `/bin/sh -c` for `tee`, keep credentials out of echoed commands and status summaries.

## Exporter Design

Rally has dynamic metric and task names. A production exporter should avoid one Go variable per possible Rally metric.

Recommended metric:

```text
kubebench_esrally_metric_value{
  benchmark="<cr name>",
  name="<job name>",
  metric="<Rally Metric column>",
  task="<Rally Task column>",
  unit="<Rally Unit column>"
}
```

Parser behavior:

* Parse CSV with `encoding/csv`, not string splitting.
* Detect columns by header name.
* Trim spaces.
* Skip empty, unavailable, or non-numeric `Value` cells.
* Preserve metric/task/unit label values exactly enough to be useful, but avoid high-cardinality free-form logs.
* Test representative throughput, latency, service time, error rate, index stats, and unavailable rows.

## Elasticsearch Precheck

Production-ready parity with existing benches means Esrally should have a precheck Job.

Recommended implementation:

* Add `pkg/tools/elasticsearch.go`.
* Add `tools elasticsearch ping`.
* Perform HTTP GET against `http(s)://host:port/` or `/_cluster/health`.
* Support basic auth from `Target.User/Password`.
* Support optional TLS/insecure/CA only if this can be expressed cleanly without duplicating Rally's full `clientOptions`; otherwise document that advanced TLS is verified by Rally itself.
* Add `constants.ElasticsearchDriver` and route `NewPreCheckJob` to this tool.

## Main Risks

* Track data can be huge. Defaults and docs must not surprise users with large downloads.
* `testMode` is only for sanity checks; benchmark numbers are meaningless.
* `benchmark-only` cannot collect all metrics and is less reproducible than Rally-provisioned pipelines.
* Credentials in `clientOptions` can leak through Kubernetes pod specs if passed as args. Avoid writing them to CR status or logs.
* Telemetry can skew results and setup-level telemetry is incompatible with `benchmark-only`.
* The official image entrypoint differs from a plain CLI image. Tests should assert the intended command shape.

## Production Documentation Checklist

The final `docs/esrally.md` should include:

* What Rally is and why kubebench uses `benchmark-only`.
* Full CR example for basic unauthenticated Elasticsearch.
* Auth/TLS examples with `clientOptions`.
* Track/challenge/includeTasks examples.
* Custom track and offline examples.
* PVC-backed Rally home example or explanation.
* Metrics exposed by kubebench exporter.
* How to read status conditions and pod logs.
* Limitations: Docker pipeline, distributed load driver, telemetry, track data size, test mode.
* Troubleshooting: missing track data, auth errors, cluster not green, report file missing, nonzero error rate.
