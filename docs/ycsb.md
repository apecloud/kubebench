# YCSB

The Yahoo! Cloud Serving Benchmark ([YCSB](https://github.com/pingcap/go-ycsb)) is an open-source specification and program suite for evaluating retrieval and maintenance capabilities of computer programs. It is often used to compare the relative performance of NoSQL database management systems.

## Running YCSB

your resource file look like this:

```yaml
apiVersion: benchmark.apecloud.io/v1alpha1
kind: Ycsb
metadata:
  labels:
    app.kubernetes.io/name: ycsb
    app.kubernetes.io/instance: ycsb-sample
    app.kubernetes.io/part-of: kubebench
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: kubebench
  name: ycsb-mysql
spec:
  target:
    driver: mysql
    host: "mycluster-mysql.default.svc.cluster.local"
    port: 3306
    user: root
    password: "ncqsgzb7"
    database: "mydb"
```

Once done creating/editing the resource file, you can run it by:

```sh
# kubectl apply -f config/samples/config/samples/benchmark_v1alpha1_ycsb.yaml # if edited the original one
# kubectl apply -f <path_to_file> # if created a new cr file
```

Deploying the `cr.yaml` would reuslt in:
```sh
# kubectl get ycsbs.benchmark.apecloud.io              
NAME          STATUS    COMPLETIONS   AGE
tpch-sample   Running   0/1           26s

# kubectl get pod
NAME         STATUS     COMPLETIONS   AGE
ycsb-mysql   Complete   2/2           87s
```

You can look at a result by using `kubectl log`, it should look like:
```sh
Using request distribution 'uniform' a keyrange of [0 9999]
***************** properties *****************
"mysql.user"="root"
"threadcount"="1"
"mysql.password"="f5npnslk"
"mysql.db"="mydb"
"operationcount"="10000"
"dotransactions"="true"
"command"="run"
"mysql.port"="3306"
"recordcount"="10000"
"measurement.interval"="1"
"table"="usertable"
"mysql.host"="mycluster-mysql.default.svc.cluster.local"
**********************************************
READ   - Takes(s): 1.0, Count: 4025, OPS: 4044.0, Avg(us): 175, Min(us): 109, Max(us): 5423, 50th(us): 166, 90th(us): 208, 95th(us): 232, 99th(us): 302, 99.9th(us): 606, 99.99th(us): 5423
TOTAL  - Takes(s): 1.0, Count: 4248, OPS: 4262.2, Avg(us): 224, Min(us): 109, Max(us): 5423, 50th(us): 168, 90th(us): 232, 95th(us): 800, 99th(us): 1231, 99.9th(us): 2087, 99.99th(us): 5423
UPDATE - Takes(s): 1.0, Count: 223, OPS: 224.8, Avg(us): 1114, Min(us): 613, Max(us): 5015, 50th(us): 1054, 90th(us): 1341, 95th(us): 1477, 99th(us): 2551, 99.9th(us): 5015, 99.99th(us): 5015
READ   - Takes(s): 2.0, Count: 8198, OPS: 4110.3, Avg(us): 172, Min(us): 109, Max(us): 5423, 50th(us): 165, 90th(us): 206, 95th(us): 228, 99th(us): 294, 99.9th(us): 684, 99.99th(us): 1784
TOTAL  - Takes(s): 2.0, Count: 8631, OPS: 4326.8, Avg(us): 220, Min(us): 109, Max(us): 5423, 50th(us): 166, 90th(us): 226, 95th(us): 718, 99th(us): 1241, 99.9th(us): 1845, 99.99th(us): 5015
UPDATE - Takes(s): 2.0, Count: 433, OPS: 217.6, Avg(us): 1127, Min(us): 613, Max(us): 5015, 50th(us): 1074, 90th(us): 1371, 95th(us): 1531, 99th(us): 2551, 99.9th(us): 5015, 99.99th(us): 5015
**********************************************
Run finished, takes 2.320459918s
READ   - Takes(s): 2.3, Count: 9491, OPS: 4102.9, Avg(us): 172, Min(us): 109, Max(us): 5423, 50th(us): 164, 90th(us): 206, 95th(us): 228, 99th(us): 292, 99.9th(us): 684, 99.99th(us): 3319
TOTAL  - Takes(s): 2.3, Count: 10000, OPS: 4322.6, Avg(us): 220, Min(us): 109, Max(us): 5423, 50th(us): 166, 90th(us): 227, 95th(us): 758, 99th(us): 1236, 99.9th(us): 1845, 99.99th(us): 5015
UPDATE - Takes(s): 2.3, Count: 509, OPS: 220.5, Avg(us): 1119, Min(us): 613, Max(us): 5015, 50th(us): 1071, 90th(us): 1356, 95th(us): 1513, 99th(us): 2333, 99.9th(us): 4635, 99.99th(us): 5015
```

You can also look at all result by using `kubectl describe`, it should look like
```yaml
Name:         ycsb-mysql
Namespace:    default
Labels:       app.kubernetes.io/created-by=kubebench
              app.kubernetes.io/instance=ycsb-sample
              app.kubernetes.io/managed-by=kustomize
              app.kubernetes.io/name=ycsb
              app.kubernetes.io/part-of=kubebench
Annotations:  <none>
API Version:  benchmark.apecloud.io/v1alpha1
Kind:         Ycsb
Metadata:
  Creation Timestamp:  2023-08-12T14:46:27Z
  Generation:          1
  Resource Version:    6391
  UID:                 291f8756-a2bb-4553-99f7-895d96d828f9
Spec:
  Insert Proportion:             0
  Operation Count:               10000
  Read Modify Write Proportion:  0
  Read Proportion:               0
  Record Count:                  10000
  Scan Proportion:               0
  Step:                          all
  Target:
    Database:  mydb
    Driver:    mysql
    Host:      mycluster-mysql.default.svc.cluster.local
    Password:  f5npnslk
    Port:      3306
    User:      root
  Threads:
    1
  Update Proportion:  0
Status:
  Completions:  2/2
  Conditions:
    Last Transition Time:  2023-08-12T14:46:59Z
    Message:               Run finished, takes 16.598536716s
                           INSERT - Takes(s): 16.6, Count: 10000, OPS: 602.8, Avg(us): 1498, Min(us): 731, Max(us): 75327, 50th(us): 1356, 90th(us): 1886, 95th(us): 2215, 99th(us): 4069, 99.9th(us): 12375, 99.99th(us): 48383
                           TOTAL  - Takes(s): 16.6, Count: 10000, OPS: 602.8, Avg(us): 1498, Min(us): 731, Max(us): 75327, 50th(us): 1356, 90th(us): 1886, 95th(us): 2215, 99th(us): 4069, 99.9th(us): 12375, 99.99th(us): 48383
    Reason:   RecordLog
    Status:   True
    Type:     ycsb-mysql-prepare-s8jdb-Running
  Phase:      Complete
  Succeeded:  2
  Total:      2
Events:       <none>
```
