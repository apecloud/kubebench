# Esrally

Esrally runs [Elastic Rally](https://github.com/elastic/rally) against an existing Elasticsearch cluster.
Kubebench uses Rally's `race` command with `--pipeline=benchmark-only` because the operator benchmarks databases that already exist in Kubernetes. Rally's official Docker image supports this Docker flow; it does not support Rally-managed Elasticsearch provisioning or distributed load-driver mode.

## Basic Example

```yaml
apiVersion: benchmark.apecloud.io/v1alpha1
kind: Esrally
metadata:
  name: esrally-geonames
spec:
  target:
    driver: elasticsearch
    host: elasticsearch.default.svc
    port: 9200
  track: geonames
  challenge: append-no-conflicts
  reportFormat: csv
  reportFile: /var/log/esrally-report.csv
  metrics: true
```

For basic `spec.target.host:port` HTTP targets, the controller creates an Elasticsearch precheck Job first, then one Rally run Job. The run container executes the equivalent of:

```sh
esrally race --pipeline=benchmark-only --target-hosts=<host:port> --track=<track> --report-format=csv --report-file=/var/log/esrally-report.csv
```

Console output is written to `/var/log/esrally.log`. The CSV report is the stable source for status summaries and Prometheus metrics.

## Concepts

Rally calls one benchmark execution a race. A race combines the benchmark pipeline, target hosts, track, challenge, optional task filters, track parameters, client options, telemetry, and reporting.

A track is the workload definition. It can come from Rally's default track repository, a configured track repository, or a local path inside the container through `trackPath`. A challenge is a scenario within a track. `includeTasks` can restrict a challenge to selected tasks.

`testMode` is useful for smoke tests because it makes Rally run a tiny workload. Do not use test-mode results for benchmark comparisons.

## Spec Fields

| Field | Description |
|-------|-------------|
| `track` | Rally track name. Defaults to `geonames`. Ignored when `trackPath` is set. |
| `trackRepository` | Optional Rally track repository name. |
| `trackPath` | Local track file or directory path inside the Rally container. |
| `challenge` | Optional Rally challenge. |
| `includeTasks` | Optional list of task names to run. |
| `trackParams` | String map passed as Rally track parameters. |
| `targetHosts` | Optional list of Rally target hosts. Defaults to `spec.target.host:spec.target.port`. |
| `clientOptions` | Raw Rally `--client-options` value for auth, TLS, compression, API keys, and timeouts. |
| `onError` | `abort`, `continue`, or `continue-on-network`. Defaults to `abort`. |
| `offline` | Adds Rally `--offline`. Track data must already exist. |
| `testMode` | Adds Rally `--test-mode`. |
| `telemetry` | Rally telemetry devices such as `node-stats` or `disk-usage-stats`. |
| `telemetryParams` | Raw Rally telemetry params. |
| `reportFormat` | `csv` or `markdown`. Defaults to `csv`; metrics require CSV. |
| `reportFile` | Report path. Keep it under `/var/log` when metrics are enabled. |
| `metrics` | Enables the exporter sidecar. Defaults to true. |
| `rallyHomePVCClaimName` | Existing PVC mounted at `/rally/.rally`; otherwise an `emptyDir` is used. |

## Auth And TLS

If `clientOptions` is empty and `spec.target.user` or `spec.target.password` is set, kubebench synthesizes Rally basic auth client options. For TLS or advanced auth, set `clientOptions` explicitly:

```yaml
spec:
  target:
    driver: elasticsearch
    host: elasticsearch.default.svc
    port: 9200
  clientOptions: "use_ssl:true,verify_certs:false,basic_auth_user:'elastic',basic_auth_password:'secret'"
```

The precheck contract is intentionally narrow. When `clientOptions` is empty and `targetHosts` is not set, kubebench runs `tools elasticsearch ping` against `spec.target.host:spec.target.port` over HTTP with optional basic auth from `spec.target.user/password`.

When `clientOptions` or `targetHosts` is set, kubebench skips the precheck and lets Rally validate the target during the run Job. Those fields can represent TLS, API keys, certificate paths, URL prefixes, proxies, and multi-host routing, and kubebench does not parse or log the raw Rally client options.

## Tracks And Tasks

```yaml
spec:
  target:
    driver: elasticsearch
    host: elasticsearch.default.svc
    port: 9200
  track: geonames
  challenge: append-no-conflicts
  includeTasks:
    - index-append
  trackParams:
    number_of_shards: "3"
    number_of_replicas: "1"
```

For custom tracks packaged in your Rally image or mounted into `/rally/.rally`, use:

```yaml
spec:
  trackPath: /rally/.rally/tracks/my-track
```

## Offline And Storage

Rally may download tracks and large corpora. By default kubebench mounts an `emptyDir` at `/rally/.rally`, so downloads disappear when the Job ends. For repeated or offline runs, create a PVC and reference it:

```yaml
spec:
  rallyHomePVCClaimName: rally-home
  offline: true
```

`offline: true` only prevents network updates/downloads. It does not provide missing track data; preload the PVC or use an image that already contains the required data.

## Metrics

When `metrics: true`, kubebench adds an exporter sidecar and exposes numeric CSV report rows as:

```text
kubebench_esrally_metric_value{benchmark,name,metric,task,unit}
```

Empty, unavailable, and non-numeric CSV values are ignored. Metric labels come from Rally's summary report fields and do not include raw logs or client options.

## Troubleshooting

If a basic HTTP target is unreachable, inspect the precheck Job logs first. It calls `tools elasticsearch ping` against `/_cluster/health`.

If auth, TLS, `targetHosts`, or other Rally client behavior fails, check Rally's run Job logs and `clientOptions` quoting. YAML quoting matters for values containing commas, quotes, or colons.

If the report file is missing, keep `reportFormat: csv`, keep `reportFile` under `/var/log`, and verify Rally completed successfully.

If `offline` fails with missing corpora or tracks, preload `/rally/.rally` through a PVC or custom image.

If error rate is nonzero, use `onError: continue` only when measuring errors is intentional. The default `abort` fails fast for unhealthy targets.

Unsupported telemetry devices can fail in `benchmark-only` Docker mode. Runtime telemetry devices such as `node-stats`, `segment-stats`, `ingest-pipeline-stats`, and `disk-usage-stats` are the safest choices.
