# TPCH

The [TPC-H](https://github.com/apecloud/customsuites#tpch) is a decision support benchmark. It consists of a suite of business oriented ad-hoc queries and concurrent data modifications. The queries and the data populating the database have been chosen to have broad industry-wide relevance. This benchmark illustrates decision support systems that examine large volumes of data, execute queries with a high degree of complexity, and give answers to critical business questions.

## Running TPCH

your resource file look like this:

```yaml
apiVersion: benchmark.apecloud.io/v1alpha1
kind: Tpch
metadata:
  labels:
    app.kubernetes.io/name: tpch
    app.kubernetes.io/instance: tpch-sample
    app.kubernetes.io/part-of: kubebench
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: kubebench
  name: tpch-sample
spec:
  target:
    host: "mycluster-mysql.default.svc.cluster.local"
    port: 3306
    user: root
    password: "ncqsgzb7"
    database: "mydb"
```

Once done creating/editing the resource file, you can run it by:

```sh
# kubectl apply -f config/samples/config/samples/benchmark_v1alpha1_tpch.yaml # if edited the original one
# kubectl apply -f <path_to_file> # if created a new cr file
```

Deploying the `cr.yaml` would reuslt in:
```sh
# kubectl get tpches.benchmark.apecloud.io             
NAME          STATUS    COMPLETIONS   AGE
tpch-sample   Running   0/1           26s

# kubectl get pod
NAME                    READY   STATUS    RESTARTS      AGE
tpch-sample-all-ztq2q   1/1     Running   0             51s
```

You can look at a result by using `kubectl log`, it should look like:
```sh
ALTER TABLE LINEITEM ADD FOREIGN KEY LINEITEM_FK2 (L_PARTKEY,L_SUPPKEY) references PARTSUPP(PS_PARTKEY, PS_SUPPKEY);
create KEY16 run cost 405.34 seconds
run tpch query
run query 1
query 1 rows: 4
query 1 run cost 8.513 seconds
run query 2
query 2 rows: 100
query 2 run cost 0.266 seconds
run query 3
query 3 rows: 10
query 3 run cost 30.219 seconds
run query 4
query 4 rows: 5
query 4 run cost 2.899 seconds
run query 5
query 5 rows: 5
query 5 run cost 6.512 seconds
run query 6
query 6 rows: 1
query 6 run cost 2.513 seconds
run query 7
query 7 rows: 4
query 7 run cost 7.690 seconds
run query 8
query 8 rows: 2
query 8 run cost 5.820 seconds
run query 9
query 9 rows: 175
query 9 run cost 42.201 seconds
run query 10
query 10 rows: 20
query 10 run cost 5.940 seconds
run query 11
query 11 rows: 1048
query 11 run cost 2.460 seconds
run query 12
query 12 rows: 2
query 12 run cost 3.685 seconds
run query 13
query 13 rows: 42
query 13 run cost 79.760 seconds
run query 14
query 14 rows: 1
query 14 run cost 406.421 seconds
run query 15
query 15 rows: 1
query 15 run cost 3.010 seconds
run query 16
query 16 rows: 18314
query 16 run cost 1.446 seconds
run query 17
query 17 rows: 1
query 17 run cost 1.205 seconds
run query 18
query 18 rows: 57
query 18 run cost 3.472 seconds
run query 19
query 19 rows: 1
query 19 run cost 1.077 seconds
run query 20
query 20 rows: 186
query 20 run cost 4.272 seconds
run query 21
query 21 rows: 100
query 21 run cost 10.965 seconds
run query 22
query 22 rows: 7
query 22 run cost 0.389 seconds
```

You can also look at all result by using `kubectl describe`, it should look like
```yaml
Name:         tpch-sample
Namespace:    default
Labels:       app.kubernetes.io/created-by=kubebench
              app.kubernetes.io/instance=tpch-sample
              app.kubernetes.io/managed-by=kustomize
              app.kubernetes.io/name=tpch
              app.kubernetes.io/part-of=kubebench
Annotations:  <none>
API Version:  benchmark.apecloud.io/v1alpha1
Kind:         Tpch
Metadata:
  Creation Timestamp:  2023-08-12T14:16:37Z
  Generation:          1
  Resource Version:    5396
  UID:                 e1abcefa-da1e-4f71-8017-73de43c1d3f2
Spec:
  Sizes:
    1
  Step:  all
  Target:
    Database:  mydb
    Host:      mycluster-mysql.default.svc.cluster.local
    Password:  f5npnslk
    Port:      3306
    User:      root
Status:
  Completions:  1/1
  Conditions:
    Last Transition Time:  2023-08-12T14:39:28Z
    Message:               run tpch query
                           run query 1
                           query 1 rows: 4
                           query 1 run cost 8.513 seconds
                           run query 2
                           query 2 rows: 100
                           query 2 run cost 0.266 seconds
                           run query 3
                           query 3 rows: 10
                           query 3 run cost 30.219 seconds
                           run query 4
                           query 4 rows: 5
                           query 4 run cost 2.899 seconds
                           run query 5
                           query 5 rows: 5
                           query 5 run cost 6.512 seconds
                           run query 6
                           query 6 rows: 1
                           query 6 run cost 2.513 seconds
                           run query 7
                           query 7 rows: 4
                           query 7 run cost 7.690 seconds
                           run query 8
                           query 8 rows: 2
                           query 8 run cost 5.820 seconds
                           run query 9
                           query 9 rows: 175
                           query 9 run cost 42.201 seconds
                           run query 10
                           query 10 rows: 20
                           query 10 run cost 5.940 seconds
                           run query 11
                           query 11 rows: 1048
                           query 11 run cost 2.460 seconds
                           run query 12
                           query 12 rows: 2
                           query 12 run cost 3.685 seconds
                           run query 13
                           query 13 rows: 42
                           query 13 run cost 79.760 seconds
                           run query 14
                           query 14 rows: 1
                           query 14 run cost 406.421 seconds
                           run query 15
                           query 15 rows: 1
                           query 15 run cost 3.010 seconds
                           run query 16
                           query 16 rows: 18314
                           query 16 run cost 1.446 seconds
                           run query 17
                           query 17 rows: 1
                           query 17 run cost 1.205 seconds
                           run query 18
                           query 18 rows: 57
                           query 18 run cost 3.472 seconds
                           run query 19
                           query 19 rows: 1
                           query 19 run cost 1.077 seconds
                           run query 20
                           query 20 rows: 186
                           query 20 run cost 4.272 seconds
                           run query 21
                           query 21 rows: 100
                           query 21 run cost 10.965 seconds
                           run query 22
                           query 22 rows: 7
                           query 22 run cost 0.389 seconds
    Reason:   RecordLog
    Status:   True
    Type:     tpch-sample-all-9d66h-Succeeded
  Phase:      Complete
  Succeeded:  1
  Total:      1
Events:       <none>
```
