# Sysbench

[sysbench](https://github.com/apecloud/customsuites#sysbench) is a scriptable multi-threaded benchmark tool based on LuaJIT. It is most frequently used for database benchmarks, but can also be used to create arbitrarily complex workloads that do not involve a database server.

## Running Sysbench

your resource file look like this:

```yaml
apiVersion: benchmark.apecloud.io/v1alpha1
kind: Sysbench
metadata:
  labels:
    app.kubernetes.io/name: sysbench
    app.kubernetes.io/instance: sysbench-sample
    app.kubernetes.io/part-of: kubebench
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: kubebench
  name: sysbench-sample
spec:
  spec:
  tables: 10
  size: 20000
  threads:
    - 2
    - 4
  types:
    - "oltp_read_write_pct"
    - "oltp_read_write"
  duration: 60
  extraArgs: 
    - "--read-percent=80" 
    - "--write-percent=20"
  target:
    driver: "pgsql"
    host: "test-pg-postgresql.default.svc.cluster.local"
    port: 5432
    user: "postgres"
    password: "dlsmdcb6"
    database: "postgres"
```

Once done creating/editing the resource file, you can run it by:

```sh
# kubectl apply -f config/samples/config/samples/benchmark_v1alpha1_sysbench.yaml # if edited the original one
# kubectl apply -f <path_to_file> # if created a new cr file
```

Deploying the `cr.yaml` would reuslt in:
```sh
# kubectl get sysbenches.benchmark.apecloud.io
NAME              STATUS    COMPLETIONS   AGE
sysbench-sample   Running   0/6           11s

# kubectl get pod
NAME                            READY   STATUS      RESTARTS      AGE
sysbench-sample-cleanup-5bgsx   0/1     Completed   0             6m2s
sysbench-sample-prepare-r94br   0/1     Completed   0             5m58s
sysbench-sample-run-0-rtp6s     0/2     Completed   0             5m52s
sysbench-sample-run-1-mfm95     0/2     Completed   0             4m48s
sysbench-sample-run-2-h4jgz     0/2     Completed   0             3m44s
```

You can look at a result by using `kubectl log`, it should look like:
```sh
Defaulted container "kubebench" out of: kubebench, metrics
run sysbench
check pgsql database exists
pgsql host test-pg-postgresql.default.svc.cluster.local connect success!
database postgres exists
/usr/bin/sysbench
cd ./sysbench/src/lua;sysbench --db-driver=pgsql --pgsql-host=test-pg-postgresql.default.svc.cluster.local --pgsql-port=5432 --pgsql-user=postgres --pgsql-password=kznh5s5v --pgsql-db=postgres --table_size=20000 --tables=10 --time=30 --threads=2 --events=0 --percentile=99 --read-percent=80 --write-percent=20 --report-interval=1 oltp_read_write_pct run
sysbench 1.0.20 (using system LuaJIT 2.1.0-beta3)

Running the test with following options:
Number of threads: 2
Report intermediate results every 1 second(s)
Initializing random number generator from current time


Initializing worker threads...

Threads started!

[ 1s ] thds: 2 tps: 108.90 qps: 11075.78 (r/w/o: 8871.82/2203.97/0.00) lat (ms,99%): 50.11 err/s: 0.00 reconn/s: 0.00
[ 2s ] thds: 2 tps: 122.06 qps: 12195.73 (r/w/o: 9764.58/2431.14/0.00) lat (ms,99%): 49.21 err/s: 0.00 reconn/s: 0.00
[ 3s ] thds: 2 tps: 110.00 qps: 10931.00 (r/w/o: 8745.00/2186.00/0.00) lat (ms,99%): 73.13 err/s: 0.00 reconn/s: 0.00
[ 4s ] thds: 2 tps: 105.97 qps: 10654.42 (r/w/o: 8532.93/2121.49/0.00) lat (ms,99%): 66.84 err/s: 0.00 reconn/s: 0.00
[ 5s ] thds: 2 tps: 112.02 qps: 11138.24 (r/w/o: 8901.79/2236.45/0.00) lat (ms,99%): 53.85 err/s: 0.00 reconn/s: 0.00
[ 6s ] thds: 2 tps: 111.13 qps: 11087.89 (r/w/o: 8865.35/2222.54/0.00) lat (ms,99%): 51.94 err/s: 0.00 reconn/s: 0.00
[ 7s ] thds: 2 tps: 95.75 qps: 9603.81 (r/w/o: 7674.79/1929.03/0.00) lat (ms,99%): 80.03 err/s: 0.00 reconn/s: 0.00
[ 8s ] thds: 2 tps: 109.00 qps: 10959.45 (r/w/o: 8786.36/2173.09/0.00) lat (ms,99%): 58.92 err/s: 0.00 reconn/s: 0.00
[ 9s ] thds: 2 tps: 123.01 qps: 12217.51 (r/w/o: 9764.40/2453.10/0.00) lat (ms,99%): 50.11 err/s: 0.00 reconn/s: 0.00
[ 10s ] thds: 2 tps: 119.01 qps: 11938.89 (r/w/o: 9553.71/2385.18/0.00) lat (ms,99%): 51.02 err/s: 0.00 reconn/s: 0.00
[ 11s ] thds: 2 tps: 118.00 qps: 11801.61 (r/w/o: 9440.69/2360.92/0.00) lat (ms,99%): 49.21 err/s: 0.00 reconn/s: 0.00
[ 12s ] thds: 2 tps: 91.96 qps: 9117.12 (r/w/o: 7283.90/1833.22/0.00) lat (ms,99%): 73.13 err/s: 0.00 reconn/s: 0.00
[ 13s ] thds: 2 tps: 71.03 qps: 7148.53 (r/w/o: 5728.03/1420.50/0.00) lat (ms,99%): 70.55 err/s: 0.00 reconn/s: 0.00
[ 14s ] thds: 2 tps: 97.00 qps: 9706.75 (r/w/o: 7754.80/1951.95/0.00) lat (ms,99%): 69.29 err/s: 0.00 reconn/s: 0.00
[ 15s ] thds: 2 tps: 119.01 qps: 11990.98 (r/w/o: 9598.78/2392.20/0.00) lat (ms,99%): 52.89 err/s: 0.00 reconn/s: 0.00
[ 16s ] thds: 2 tps: 122.00 qps: 12105.21 (r/w/o: 9689.17/2416.04/0.00) lat (ms,99%): 49.21 err/s: 0.00 reconn/s: 0.00
[ 17s ] thds: 2 tps: 118.99 qps: 11933.30 (r/w/o: 9542.44/2390.86/0.00) lat (ms,99%): 49.21 err/s: 0.00 reconn/s: 0.00
[ 18s ] thds: 2 tps: 106.00 qps: 10527.77 (r/w/o: 8418.81/2108.95/0.00) lat (ms,99%): 74.46 err/s: 0.00 reconn/s: 0.00
[ 19s ] thds: 2 tps: 120.01 qps: 12082.60 (r/w/o: 9667.48/2415.12/0.00) lat (ms,99%): 52.89 err/s: 0.00 reconn/s: 0.00
[ 20s ] thds: 2 tps: 125.00 qps: 12488.84 (r/w/o: 10000.87/2487.97/0.00) lat (ms,99%): 49.21 err/s: 0.00 reconn/s: 0.00
[ 21s ] thds: 2 tps: 121.00 qps: 12074.83 (r/w/o: 9654.86/2419.97/0.00) lat (ms,99%): 51.02 err/s: 0.00 reconn/s: 0.00
[ 22s ] thds: 2 tps: 121.01 qps: 12122.86 (r/w/o: 9699.69/2423.17/0.00) lat (ms,99%): 48.34 err/s: 0.00 reconn/s: 0.00
[ 23s ] thds: 2 tps: 122.00 qps: 12162.98 (r/w/o: 9728.98/2434.00/0.00) lat (ms,99%): 49.21 err/s: 0.00 reconn/s: 0.00
[ 24s ] thds: 2 tps: 113.00 qps: 11339.02 (r/w/o: 9066.02/2273.00/0.00) lat (ms,99%): 66.84 err/s: 0.00 reconn/s: 0.00
[ 25s ] thds: 2 tps: 119.99 qps: 12021.45 (r/w/o: 9634.56/2386.89/0.00) lat (ms,99%): 49.21 err/s: 0.00 reconn/s: 0.00
[ 26s ] thds: 2 tps: 118.01 qps: 11734.60 (r/w/o: 9374.48/2360.12/0.00) lat (ms,99%): 49.21 err/s: 0.00 reconn/s: 0.00
[ 27s ] thds: 2 tps: 101.00 qps: 10133.79 (r/w/o: 8113.83/2019.96/0.00) lat (ms,99%): 69.29 err/s: 0.00 reconn/s: 0.00
[ 28s ] thds: 2 tps: 82.99 qps: 8198.12 (r/w/o: 6538.30/1659.82/0.00) lat (ms,99%): 73.13 err/s: 0.00 reconn/s: 0.00
[ 29s ] thds: 2 tps: 82.49 qps: 8278.18 (r/w/o: 6628.31/1649.87/0.00) lat (ms,99%): 78.60 err/s: 0.00 reconn/s: 0.00
[ 30s ] thds: 2 tps: 106.66 qps: 10803.28 (r/w/o: 8651.88/2151.40/0.00) lat (ms,99%): 48.34 err/s: 0.00 reconn/s: 0.00
SQL statistics:
    queries performed:
        read:                            263680
        write:                           65920
        other:                           0
        total:                           329600
    transactions:                        3296   (109.84 per sec.)
    queries:                             329600 (10984.07 per sec.)
    ignored errors:                      0      (0.00 per sec.)
    reconnects:                          0      (0.00 per sec.)

General statistics:
    total time:                          30.0067s
    total number of events:              3296

Latency (ms):
         min:                                    7.33
         avg:                                   18.20
         max:                                   81.44
         99th percentile:                       69.29
         sum:                                59997.10

Threads fairness:
    events (avg/stddev):           1648.0000/2.00
    execution time (avg/stddev):   29.9985/0.00

```

You can also look at all result by using `kubectl describe`, it should look like
```yaml
Name:         sysbench-test
Namespace:    default
Labels:       app.kubernetes.io/created-by=kubebench
              app.kubernetes.io/instance=sysbench-sample
              app.kubernetes.io/managed-by=kustomize
              app.kubernetes.io/name=sysbench
              app.kubernetes.io/part-of=kubebench
Annotations:  <none>
API Version:  benchmark.apecloud.io/v1alpha1
Kind:         Sysbench
Metadata:
  Creation Timestamp:  2023-08-12T13:10:34Z
  Generation:          1
  Resource Version:    127995
  UID:                 cae0307a-be3c-4204-afde-3a08fabadf32
Spec:
  Duration:  30
  Extra Args:
    --read-percent=80
    --write-percent=20
  Size:    20000
  Step:    all
  Tables:  10
  Target:
    Database:  postgres
    Driver:    pgsql
    Host:      test-pg-postgresql.default.svc.cluster.local
    Password:  kznh5s5v
    Port:      5432
    User:      postgres
  Threads:
    2
    4
  Types:
    oltp_read_write_pct
    oltp_read_write
Status:
  Completions:  3/6
  Conditions:
    Last Transition Time:  2023-08-12T13:10:38Z
    Message:               
    Reason:                RecordLog
    Status:                True
    Type:                  sysbench-test-cleanup-j2fv2-Pending
    Last Transition Time:  2023-08-12T13:11:48Z
    Message:               SQL statistics:
                               queries performed:
                                   read:                            263680
                                   write:                           65920
                                   other:                           0
                                   total:                           329600
                               transactions:                        3296   (109.84 per sec.)
                               queries:                             329600 (10984.07 per sec.)
                               ignored errors:                      0      (0.00 per sec.)
                               reconnects:                          0      (0.00 per sec.)
                           General statistics:
                               total time:                          30.0067s
                               total number of events:              3296
                           Latency (ms):
                                    min:                                    7.33
                                    avg:                                   18.20
                                    max:                                   81.44
                                    99th percentile:                       69.29
                                    sum:                                59997.10
                           Threads fairness:
                               events (avg/stddev):           1648.0000/2.00
                               execution time (avg/stddev):   29.9985/0.00
    Reason:   RecordLog
    Status:   True
    Type:     sysbench-test-run-0-rtp6s-Running
  Phase:      Running
  Succeeded:  3
  Total:      6
Events:       <none>
```
