# Esrally

Esrally runs [Elastic Rally](https://github.com/elastic/rally) against an existing Elasticsearch cluster.
Kubebench uses Rally's `race` command with `--pipeline=benchmark-only` for benchmark execution and reporting, but it does not use Rally's remote track repositories or downloadable corpora. ESRally data and the local Rally track are generated inside the `run` Job.
Kubebench's default ESRally image is pinned to Rally 2.12.0.

## Generated Logs Example

```yaml
apiVersion: benchmark.apecloud.io/v1alpha1
kind: Esrally
metadata:
  name: esrally-generated-logs
spec:
  step: all
  target:
    driver: elasticsearch
    host: elasticsearch.default.svc
    port: 9200
    database: kubebench-logs
  targetVersion: 8.12.2
  dataProfile: logs
  documentCount: 100000
  workload: all
```

## Generated Metrics Example

```yaml
apiVersion: benchmark.apecloud.io/v1alpha1
kind: Esrally
metadata:
  name: esrally-generated-metrics
spec:
  step: all
  target:
    driver: elasticsearch
    host: elasticsearch.default.svc
    port: 9200
    database: kubebench-metrics
  targetVersion: 8.12.2
  dataProfile: metricbeat
  documentCount: 100000
  workload: mixed
```

For generated data, Kubebench supports the basic target fields `spec.target.host`, `spec.target.port`, `spec.target.tls`, `spec.target.user`, and `spec.target.password`.

The run step generates a local Rally track and local JSON corpus in the run Pod before invoking Rally. Kubebench chooses the track path, challenge, included tasks, corpus file, and track parameters internally, and always passes `--offline` to Rally. It does not expose API fields for remote Rally tracks, track repositories, corpus downloads, or local track internals.

`targetVersion` identifies the target Elasticsearch version for kubebench compatibility decisions. Kubebench does not pass it as Rally `--distribution-version` because ESRally uses Rally's `benchmark-only` pipeline against an already-running cluster.

## Concepts

Rally calls one benchmark execution a race. In Kubebench ESRally, the race owns generated-data indexing, refresh, search, aggregation, and mixed workload tasks so write and read metrics appear in Rally output.

Kubebench runs a local Rally track against the generated index. Track internals such as challenge names, task names, and template parameters are not part of the ESRally user API.

## Steps

`spec.step` controls ESRally Jobs in the same shape as other Kubebench benchmarks:

| `step` | Jobs |
|--------|------|
| `cleanup` | Delete `spec.target.database`; missing index is success. |
| `prepare` | Validate generated workload configuration. |
| `run` | Generate the local track/corpus and run Rally against the target. |
| `all` | Cleanup, validate, then run Rally. |

When Kubebench can use the basic target fields directly, it adds a precheck Job before the selected work Jobs.

## Spec Fields

| Field | Description |
|-------|-------------|
| `step` | `cleanup`, `prepare`, `run`, or `all`. Defaults to `all`. |
| `targetVersion` | Optional Elasticsearch target version, such as `7.17.0` or `8.12.2`, used for version-aware kubebench behavior. |
| `onError` | `abort` or `continue`. Defaults to `abort`. |
| `telemetry` | Optional Rally runtime telemetry devices supported by `--pipeline=benchmark-only`: `node-stats`, `recovery-stats`, `ccr-stats`, `segment-stats`, `transform-stats`, `searchable-snapshots-stats`, `shard-stats`, `data-stream-stats`, `ingest-pipeline-stats`, `disk-usage-stats`, or `geoip-stats`. |
| `dataProfile` | Generated dataset profile. One of `logs`, `metrics`, `http_logs`, `metricbeat`, `geonames`, `nyc_taxis`, `noaa`, `nested`, `pmc`, `so`, or `dense_vector`. Defaults to `logs`. |
| `documentCount` | Number of generated documents. Defaults to `10000`. |
| `workload` | Generated Rally workload profile: `index`, `search`, `mixed`, or `all`. Defaults to `all`. |

`spec.target.database` is the generated Elasticsearch index name. When omitted, it defaults to `kubebench`.

Kubebench passes the generated target index and, when set, `targetVersion` to its generated local Rally track as internal `--track-params`. Generated-data cleanup, prepare validation, and run generation support Elasticsearch 6 and newer; for Elasticsearch 6, the generated Rally corpus uses `_doc` target type compatibility, while Elasticsearch 7 and newer use typeless corpus metadata. Unsupported target versions fail before cleanup deletes the target index.

## Workloads

`index` creates the target index, bulk indexes generated documents, and refreshes the index.

`search` creates and bulk loads the generated corpus, then runs match-all, term/keyword, range, and aggregation searches.

`mixed` combines write and read tasks in one Rally race.

`all` runs the full generated workflow: indexing, refresh, searches, aggregations, and mixed write/read tasks.

## Auth And TLS

When both `spec.target.user` and `spec.target.password` are set, Kubebench uses those credentials for generated cleanup requests and synthesizes Rally basic auth client options internally for the run step. If only one of the two fields is set, Kubebench does not send partial basic auth credentials.

```yaml
spec:
  target:
    driver: elasticsearch
    host: elasticsearch.default.svc
    port: 9200
    user: elastic
    password: secret
```

`spec.target.tls` defaults to `false`. Set it to `true` when the Elasticsearch HTTP endpoint uses TLS. Kubebench then uses HTTPS for the precheck, cleanup, and Rally run paths, skips TLS certificate verification for precheck and cleanup, and passes Rally `verify_certs:false` for the run step.

```yaml
spec:
  target:
    driver: elasticsearch
    host: elasticsearch.default.svc
    port: 9200
    tls: true
    user: elastic
    password: secret
```

Kubebench runs `tools elasticsearch ping` against `spec.target.host:spec.target.port` before selected work Jobs, using `spec.target.tls` to choose HTTP or HTTPS and optional paired basic auth.

## Generated Data Shapes

`logs` creates compact service log documents with timestamp, service, host, HTTP method/path/status, latency, bytes, and message fields.

`metrics` creates generic infrastructure metric documents with timestamp, host/pod/container labels, and CPU, memory, disk, and network numeric fields.

`http_logs` creates HTTP access log documents with client IP, method, URL, status, bytes, user agent, referer, and response latency fields.

`metricbeat` creates Metricbeat-like host and Kubernetes metric documents with event dataset, host, pod, namespace, CPU, memory, and filesystem utilization fields.

`geonames` creates location documents with place names, country and feature codes, population/elevation fields, and a `geo_point` location.

`nyc_taxis` creates trip documents with pickup/dropoff timestamps, pickup/dropoff `geo_point` locations, passenger count, distance, fare, payment type, and vendor fields.

`noaa` creates weather-observation documents with station metadata, `geo_point` location, temperature, precipitation, wind speed, and weather type fields.

`nested` creates order documents with nested line items for nested-query style workloads.

`pmc` creates full-text publication-style documents with title, abstract, body, journal, year, and authors fields.

`so` creates Stack Overflow-style question documents with title, body, tags, scores, answer counts, and nested accepted-answer content.

`dense_vector` creates small semantic-search documents with an 8-dimensional `dense_vector` field. When `targetVersion` is set, Kubebench rejects this profile for Elasticsearch versions below 7 before the run step.

The generated Rally track creates the target index with profile-specific mappings before bulk indexing. Cleanup is the supported way to remove any existing generated index before a fresh run.

## Storage

Kubebench mounts an `emptyDir` at `/rally/.rally` so Rally can write local state during the run Job. The generated track and corpus are written to the run container local filesystem under `/tmp/kubebench-esrally-track`; they are not shared with prepare and are not used as a persistent corpus cache.

## Metrics

Kubebench always adds an ESRally exporter sidecar. Numeric rows from Rally's internal CSV summary report are exposed as:

```text
kubebench_esrally_metric_value{benchmark,name,metric,task,unit}
```

Empty, unavailable, and non-numeric CSV values are ignored. Metric labels come from Rally's summary report fields and do not include raw logs or client options.

The report format and path are not user-facing spec fields. They are part of the kubebench ESRally integration contract so the workload container, status summarizer, and exporter sidecar stay aligned.

## Troubleshooting

If a basic HTTP target is unreachable, inspect the precheck Job logs first. It calls `tools elasticsearch ping` against `/_cluster/health`.

If the run step exits before Rally starts, inspect the run Job logs for Rally errors from the packaged generated-data track.

If auth fails during the run step, check `spec.target.user/password` and the Rally run Job logs. Advanced Rally client options such as TLS and API keys are not exposed in the generated-data ESRally API.

If the report file is missing, verify Rally completed successfully and inspect `/var/log/esrally.log` in the run Job.

If error rate is nonzero, use `onError: continue` only when measuring errors is intentional. The default `abort` fails fast for unhealthy targets.

Unsupported setup telemetry devices are rejected by the CRD. With `--pipeline=benchmark-only`, Rally 2.12.0 supports only runtime telemetry devices whose names end in `-stats`.
