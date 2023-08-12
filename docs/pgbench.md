# Pgbench

[pgbench](https://docs.postgresql.fr/12/pgbench.html) is a benchmarking tool specifically designed for PostgreSQL, which is an open-source relational database management system. 

## Running Pgbench

your resource file look like this:

```yaml
apiVersion: benchmark.apecloud.io/v1alpha1
kind: Pgbench
metadata:
  labels:
    app.kubernetes.io/name: pgbench
    app.kubernetes.io/instance: pgbench-sample
    app.kubernetes.io/part-of: kubebench
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: kubebench
  name: pgbench-sample
spec:
  scale: 10
  clients:
    - 2
    - 4
  threads: 2
    #transactions: 30000
  duration: 60
  target:
    host: "test-pg-postgresql.default.svc.cluster.local"
    port: 5432
    user: "postgres"
    password: "xxx"
    database: "postgres"
```

Once done creating/editing the resource file, you can run it by:

```sh
# kubectl apply -f config/samples/config/samples/benchmark_v1alpha1_pgbench.yaml # if edited the original one
# kubectl apply -f <path_to_file> # if created a new cr file
```

Deploying the `cr.yaml` would reuslt in:
```sh
# kubectl get pgbenches.benchmark.apecloud.io
NAME             STATUS    COMPLETIONS   AGE
pgbench-sample   Running   0/4           19s

# kubectl get pod
NAME                           READY   STATUS      RESTARTS      AGE
pgbench-sample-cleanup-mdqcv   0/1     Completed   0             10s
pgbench-sample-prepare-n9zc5   0/1     Completed   0             6s
pgbench-sample-run-0-zv6vs     2/2     Running     0             1s
```

You can look at a result by using `kubectl log`, it should look like:
```sh
Defaulted container "kubebench" out of: kubebench, metrics
pgbench (14.8 (Ubuntu 14.8-1.pgdg22.04+1))
starting vacuum...end.
progress: 1.0 s, 1747.9 tps, lat 1.130 ms stddev 3.471
progress: 2.0 s, 1734.0 tps, lat 1.153 ms stddev 3.380
progress: 3.0 s, 1643.0 tps, lat 1.217 ms stddev 3.932
progress: 4.0 s, 1757.0 tps, lat 1.138 ms stddev 3.511
progress: 5.0 s, 1784.9 tps, lat 1.120 ms stddev 3.466
progress: 6.0 s, 1804.0 tps, lat 1.108 ms stddev 3.361
progress: 7.0 s, 1794.0 tps, lat 1.115 ms stddev 3.543
progress: 8.0 s, 1847.1 tps, lat 1.083 ms stddev 3.446
progress: 9.0 s, 1725.0 tps, lat 1.159 ms stddev 3.818
progress: 10.0 s, 1521.9 tps, lat 1.314 ms stddev 3.603
progress: 11.0 s, 1822.1 tps, lat 1.097 ms stddev 3.453
progress: 12.0 s, 1778.0 tps, lat 1.125 ms stddev 3.162
progress: 13.0 s, 1807.1 tps, lat 1.107 ms stddev 3.563
progress: 14.0 s, 1744.8 tps, lat 1.146 ms stddev 3.504
progress: 15.0 s, 1816.1 tps, lat 1.101 ms stddev 3.474
progress: 16.0 s, 1800.1 tps, lat 1.111 ms stddev 3.301
progress: 17.0 s, 1784.0 tps, lat 1.121 ms stddev 3.294
progress: 18.0 s, 1660.0 tps, lat 1.205 ms stddev 3.975
progress: 19.0 s, 1803.1 tps, lat 1.109 ms stddev 3.394
progress: 20.0 s, 1717.9 tps, lat 1.164 ms stddev 3.210
progress: 21.0 s, 1812.0 tps, lat 1.104 ms stddev 3.460
progress: 22.0 s, 1778.0 tps, lat 1.124 ms stddev 3.410
progress: 23.0 s, 1758.0 tps, lat 1.137 ms stddev 3.595
progress: 24.0 s, 1819.0 tps, lat 1.099 ms stddev 3.430
progress: 25.0 s, 1789.9 tps, lat 1.117 ms stddev 3.559
progress: 26.0 s, 1808.1 tps, lat 1.106 ms stddev 3.358
progress: 27.0 s, 1859.0 tps, lat 1.076 ms stddev 3.358
progress: 28.0 s, 1858.1 tps, lat 1.077 ms stddev 3.427
progress: 29.0 s, 1867.9 tps, lat 1.070 ms stddev 3.409
transaction type: <builtin: TPC-B (sort of)>
scaling factor: 10
query mode: simple
number of clients: 2
number of threads: 2
duration: 30 s
number of transactions actually processed: 53234
latency average = 1.126 ms
latency stddev = 3.476 ms
initial connection time = 13.927 ms
tps = 1775.255590 (without initial connection time)
```

You can also look at all result by using `kubectl describe`, it should look like
```yaml
Name:         pgbench-sample
Namespace:    default
Labels:       app.kubernetes.io/created-by=kubebench
              app.kubernetes.io/instance=pgbench-sample
              app.kubernetes.io/managed-by=kustomize
              app.kubernetes.io/name=pgbench
              app.kubernetes.io/part-of=kubebench
Annotations:  <none>
API Version:  benchmark.apecloud.io/v1alpha1
Kind:         Pgbench
Metadata:
  Creation Timestamp:  2023-08-11T13:44:05Z
  Generation:          1
  Resource Version:    49418
  UID:                 9d0e751e-eb9f-4d4e-b3da-6e3043ea2920
Spec:
  Clients:
    2
    4
  Connect:   false
  Duration:  30
  Scale:     10
  Step:      all
  Target:
    Database:    postgres
    Host:        test-pg-postgresql.default.svc.cluster.local
    Password:    kznh5s5v
    Port:        5432
    User:        postgres
  Threads:       2
  Transactions:  0
Status:
  Completions:  4/4
  Conditions:
    Last Transition Time:  2023-08-11T13:44:18Z
    Message:               
    Reason:                RecordLog
    Status:                True
    Type:                  pgbench-sample-cleanup-jc8nb-Succeeded
    Last Transition Time:  2023-08-11T13:44:24Z
    Message:               
    Reason:                RecordLog
    Status:                True
    Type:                  pgbench-sample-prepare-2l5x9-Succeeded
    Last Transition Time:  2023-08-11T13:45:28Z
    Message:               transaction type: <builtin: TPC-B (sort of)>
                           scaling factor: 10
                           query mode: simple
                           number of clients: 2
                           number of threads: 2
                           duration: 30 s
                           number of transactions actually processed: 53870
                           latency average = 1.113 ms
                           latency stddev = 3.495 ms
                           initial connection time = 10.590 ms
                           tps = 1796.251586 (without initial connection time)
    Reason:                RecordLog
    Status:                True
    Type:                  pgbench-sample-run-0-pfwpq-Running
    Last Transition Time:  2023-08-11T13:46:32Z
    Message:               transaction type: <builtin: TPC-B (sort of)>
                           scaling factor: 10
                           query mode: simple
                           number of clients: 4
                           number of threads: 2
                           duration: 30 s
                           number of transactions actually processed: 53899
                           latency average = 2.216 ms
                           latency stddev = 8.860 ms
                           initial connection time = 67.960 ms
                           tps = 1800.615936 (without initial connection time)
    Reason:   RecordLog
    Status:   True
    Type:     pgbench-sample-run-1-5xb2n-Running
  Phase:      Complete
  Succeeded:  4
  Total:      4
Events:       <none>

```