# TPCDS Benchmark

[TPCDS](https://github.com/apecloud-inc/tpcds) is a decision support benchmark that models several aspects of a decision support system, including queries and data maintenance. The benchmark provides a set of 99 queries and the data model to support them. The queries are representative of queries that are commonly run against decision support systems. The benchmark is designed to be run against a database system that is capable of handling large volumes of data.

## Running the Benchmark
your resource file looks like this:
```yaml
apiVersion: benchmark.apecloud.io/v1alpha1
kind: Tpcds
metadata:
  labels:
    app.kubernetes.io/name: tpcds
    app.kubernetes.io/instance: tpcds-sample
    app.kubernetes.io/part-of: kubebench
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: kubebench
  name: tpcds-sample
spec:
  size: 1
  target:
    host: "test-pg-postgresql.default.svc.cluster.local"
    port: 5432
    user: postgres
    password: "25ld8qc9"
    database: "test"
    driver: "postgresql"
```

Once done creating/editing the resource file, you can run it by:
```shell
# kubectl apply -f config/samples/config/samples/benchmark_v1alpha1_tpcds.yaml
# kubectl apply -f <path_to_file> # if created a new cr file
```

Deploying the `cr.yaml` would result in:
```shell
# kubectl get tpcds.benchmark.apecloud.io

NAME           STATUS    COMPLETIONS   AGE
tpcds-m1z5yt   Running   3/4           3m53s

# kubectl get pod
NAME                          READY   STATUS      RESTARTS   AGE
tpcds-m1z5yt-precheck-7zjwf   0/1     Completed   0          4m49s
tpcds-m1z5yt-cleanup-zb4bm    0/1     Completed   0          4m38s
tpcds-m1z5yt-prepare-jlss4    0/1     Completed   0          4m16s
tpcds-m1z5yt-run-p8fkt        1/1     Running     0          2m1s
test-pg-postgresql-0          5/5     Running     0          11m
```

You can look at a result of the benchmark by `kubectl logs <pod_name> -c tpcds-run`:
```shell
# kubectl logs tpcds-m1z5yt-run-p8fkt -c tpcds-run
-------------- TPCDS_Run --------------
-------------- Gen_Query_PG --------------
qgen2 Query Generator (Version 3.2.0)
Copyright Transaction Processing Performance Council (TPC) 2001 - 2021
Warning: This scale factor is valid for QUALIFICATION ONLY
Parsed 99 templates
run query 1
query 1 rows: 100
query 1 run cost 1366.673 seconds
run query 2
query 2 rows: 2513
query 2 run cost 3.745 seconds
run query 3
query 3 rows: 31
query 3 run cost 0.291 seconds
run query 4
query 4 rows: 4
query 4 run cost 3042.811 seconds
run query 5
query 5 rows: 100
query 5 run cost 4.529 seconds
run query 6
query 6 rows: 44
query 6 run cost 82.256 seconds
run query 7
query 7 rows: 100
query 7 run cost 4.098 seconds
run query 8
query 8 rows: 0
query 8 run cost 0.327 seconds
run query 9
query 9 rows: 1
query 9 run cost 6.822 seconds
run query 10
query 10 rows: 1
query 10 run cost 15.203 seconds
run query 11
query 11 rows: 95
query 11 run cost 1603.787 seconds
run query 12
query 12 rows: 100
query 12 run cost 1.016 seconds
run query 13
query 13 rows: 1
query 13 run cost 1.507 seconds
run query 14
query 14 rows: 100
query 14 run cost 73.970 seconds
run query 15
query 15 rows: 100
query 15 run cost 1.463 seconds
run query 16
query 16 rows: 1
query 16 run cost 1.603 seconds
run query 17
query 17 rows: 1
query 17 run cost 6.614 seconds
run query 18
query 18 rows: 100
query 18 run cost 2.354 seconds
run query 19
query 19 rows: 100
query 19 run cost 2.217 seconds
run query 20
query 20 rows: 100
query 20 run cost 2.020 seconds
run query 21
query 21 rows: 100
query 21 run cost 2.783 seconds
run query 22
query 22 rows: 100
query 22 run cost 16.036 seconds
run query 23
query 23 rows: 0
query 23 run cost 23.254 seconds
run query 24
query 24 rows: 0
query 24 run cost 0.405 seconds
run query 25
query 25 rows: 0
query 25 run cost 8.983 seconds
run query 26
query 26 rows: 100
query 26 run cost 2.184 seconds
run query 27
query 27 rows: 100
query 27 run cost 2.541 seconds
run query 28
query 28 rows: 1
query 28 run cost 5.487 seconds
run query 29
query 29 rows: 1
query 29 run cost 4.911 seconds
run query 30
query 30 rows: 63
query 30 run cost 7.409 seconds
run query 31
query 31 rows: 43
query 31 run cost 7.984 seconds
run query 32
query 32 rows: 1
query 32 run cost 1.061 seconds
run query 33
query 33 rows: 100
query 33 run cost 2.323 seconds
run query 34
query 34 rows: 374
query 34 run cost 1.315 seconds
run query 35
query 35 rows: 100
query 35 run cost 4.924 seconds
run query 36
query 36 rows: 100
query 36 run cost 3.456 seconds
run query 37
query 37 rows: 0
query 37 run cost 0.082 seconds
run query 38
query 38 rows: 1
query 38 run cost 7.109 seconds
run query 39
query 39 rows: 6
query 39 run cost 13.806 seconds
run query 40
query 40 rows: 100
query 40 run cost 2.735 seconds
run query 41
query 41 rows: 1
query 41 run cost 1.108 seconds
run query 42
query 42 rows: 10
query 42 run cost 2.035 seconds
run query 43
query 43 rows: 6
query 43 run cost 2.105 seconds
run query 44
query 44 rows: 10
query 44 run cost 3.316 seconds
run query 45
query 45 rows: 14
query 45 run cost 0.909 seconds
run query 46
query 46 rows: 100
query 46 run cost 1.701 seconds
run query 47
query 47 rows: 100
query 47 run cost 4.933 seconds
run query 48
query 48 rows: 1
query 48 run cost 3.070 seconds
run query 49
query 49 rows: 34
query 49 run cost 0.915 seconds
run query 50
query 50 rows: 6
query 50 run cost 0.465 seconds
run query 51
query 51 rows: 100
query 51 run cost 3.710 seconds
run query 52
query 52 rows: 100
query 52 run cost 0.921 seconds
run query 53
query 53 rows: 100
query 53 run cost 0.753 seconds
run query 54
query 54 rows: 0
query 54 run cost 2.440 seconds
run query 55
query 55 rows: 73
query 55 run cost 1.185 seconds
run query 56
query 56 rows: 100
query 56 run cost 5.781 seconds
run query 57
query 57 rows: 100
query 57 run cost 3.812 seconds
run query 58
query 58 rows: 0
query 58 run cost 6.190 seconds
run query 59
query 59 rows: 100
query 59 run cost 4.311 seconds
run query 60
query 60 rows: 100
query 60 run cost 8.283 seconds
run query 61
query 61 rows: 1
query 61 run cost 0.103 seconds
run query 62
query 62 rows: 100
query 62 run cost 0.590 seconds
run query 63
query 63 rows: 100
query 63 run cost 3.539 seconds
run query 64
query 64 rows: 8
query 64 run cost 2.179 seconds
run query 65
query 65 rows: 100
query 65 run cost 5.196 seconds
run query 66
query 66 rows: 5
query 66 run cost 1.917 seconds
run query 67
query 67 rows: 100
query 67 run cost 5.978 seconds
run query 68
query 68 rows: 100
query 68 run cost 2.813 seconds
run query 69
query 69 rows: 100
query 69 run cost 2.603 seconds
run query 70
query 70 rows: 3
query 70 run cost 3.870 seconds
run query 71
query 71 rows: 1129
query 71 run cost 3.089 seconds
run query 72
query 72 rows: 100
query 72 run cost 3.081 seconds
run query 73
query 73 rows: 3
query 73 run cost 1.556 seconds
run query 74
query 74 rows: 100
query 74 run cost 748.336 seconds
run query 75
query 75 rows: 100
query 75 run cost 10.709 seconds
run query 76
query 76 rows: 100
query 76 run cost 2.505 seconds
run query 77
query 77 rows: 44
query 77 run cost 2.933 seconds
run query 78
query 78 rows: 100
query 78 run cost 8.433 seconds
run query 79
query 79 rows: 100
query 79 run cost 2.946 seconds
run query 80
query 80 rows: 100
query 80 run cost 4.031 seconds
run query 81
query 81 rows: 100
query 81 run cost 32.415 seconds
run query 82
query 82 rows: 2
query 82 run cost 0.175 seconds
run query 83
query 83 rows: 22
query 83 run cost 0.777 seconds
run query 84
query 84 rows: 18
query 84 run cost 0.217 seconds
run query 85
query 85 rows: 2
query 85 run cost 0.612 seconds
run query 86
query 86 rows: 100
query 86 run cost 0.674 seconds
run query 87
query 87 rows: 1
query 87 run cost 5.797 seconds
run query 88
query 88 rows: 1
query 88 run cost 9.585 seconds
run query 89
query 89 rows: 100
query 89 run cost 1.552 seconds
run query 90
query 90 rows: 1
query 90 run cost 0.950 seconds
run query 91
query 91 rows: 1
query 91 run cost 0.330 seconds
run query 92
query 92 rows: 1
query 92 run cost 0.193 seconds
run query 93
query 93 rows: 0
query 93 run cost 0.101 seconds
run query 94
query 94 rows: 1
query 94 run cost 0.344 seconds
run query 95
query 95 rows: 1
query 95 run cost 65.834 seconds
run query 96
query 96 rows: 1
query 96 run cost 1.890 seconds
run query 97
query 97 rows: 1
query 97 run cost 2.313 seconds
run query 98
query 98 rows: 2531
query 98 run cost 3.434 seconds
run query 99
query 99 rows: 90
query 99 run cost 1.259 seconds
...
```
