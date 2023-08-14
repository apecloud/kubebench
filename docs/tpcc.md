# TPCC

[TPC-C](https://github.com/apecloud/benchmarksql) (Transaction Processing Performance Council - Benchmark C) is a widely used benchmark for evaluating the performance of database management systems, particularly those used for Online Transaction Processing (OLTP) workloads. 

## Running TPCC

your resource file look like this:

```yaml
apiVersion: benchmark.apecloud.io/v1alpha1
kind: Tpcc
metadata:
  labels:
    app.kubernetes.io/name: tpcc
    app.kubernetes.io/instance: tpcc-sample
    app.kubernetes.io/part-of: kubebench
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: kubebench
  name: tpcc-sample
spec:
  threads:
    - 1
    - 2
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
# kubectl apply -f config/samples/config/samples/benchmark_v1alpha1_tpcc.yaml # if edited the original one
# kubectl apply -f <path_to_file> # if created a new cr file
```

Deploying the `cr.yaml` would reuslt in:
```sh
# kubectl get tpccs.benchmark.apecloud.io
NAME          STATUS    COMPLETIONS   AGE
tpcc-sample   Running   1/4           2m30s

# kubectl get pod
NAME                        READY   STATUS      RESTARTS      AGE
tpcc-sample-cleanup-r84qg   0/1     Completed   0             3m2s
tpcc-sample-prepare-2sb2t   0/1     Completed   0             114s
tpcc-sample-run-0-8bn52     1/1     Running     0             7s
```

You can look at a result by using `kubectl log`, it should look like:
```sh
13:22:26,560 [Thread-1] INFO   jTPCC : Term-00, elapsed=59956,latency=9,dblatency=9,ttype=NEW_ORDER,rbk=0,dskipped=0,error=0
13:22:26,570 [Thread-1] INFO   jTPCC : Term-00, elapsed=59966,latency=10,dblatency=10,ttype=NEW_ORDER,rbk=0,dskipped=0,error=0
13:22:26,573 [Thread-1] INFO   jTPCC : Term-00, elapsed=59969,latency=3,dblatency=3,ttype=PAYMENT,rbk=0,dskipped=0,error=0
13:22:26,577 [Thread-1] INFO   jTPCC : Term-00, elapsed=59972,latency=2,dblatency=2,ttype=PAYMENT,rbk=0,dskipped=0,error=0
13:22:26,585 [Thread-1] INFO   jTPCC : Term-00, elapsed=59981,latency=8,dblatency=8,ttype=NEW_ORDER,rbk=0,dskipped=0,error=0
13:22:26,594 [Thread-1] INFO   jTPCC : Term-00, elapsed=59990,latency=9,dblatency=9,ttype=NEW_ORDER,rbk=0,dskipped=0,error=0
13:22:26,600 [Thread-1] INFO   jTPCC : Term-00, elapsed=59996,latency=6,dblatency=6,ttype=NEW_ORDER,rbk=0,dskipped=0,error=0
13:22:26,605 [Thread-1] INFO   jTPCC : Term-00, elapsed=60001,latency=5,dblatency=5,ttype=PAYMENT,rbk=0,dskipped=0,error=0

13:22:26,757 [Thread-1] INFO   jTPCC : Term-00, 
13:22:26,758 [Thread-1] INFO   jTPCC : Term-00, 
13:22:26,783 [Thread-1] INFO   jTPCC : Term-00, Measured tpmC (NewOrders) = 4681.21
13:22:26,784 [Thread-1] INFO   jTPCC : Term-00, Measured tpmTOTAL = 10513.54
13:22:26,784 [Thread-1] INFO   jTPCC : Term-00, Session Start     = 2023-08-12 13:21:26
13:22:26,785 [Thread-1] INFO   jTPCC : Term-00, Session End       = 2023-08-12 13:22:26
13:22:26,788 [Thread-1] INFO   jTPCC : Term-00, Transaction Count = 10539
13:22:26,790 [Thread-1] INFO   jTPCC : executeTime[Payment]=17522
13:22:26,790 [Thread-1] INFO   jTPCC : executeTime[Delivery]=3933
13:22:26,790 [Thread-1] INFO   jTPCC : executeTime[Order-Status]=916
13:22:26,790 [Thread-1] INFO   jTPCC : executeTime[Stock-Level]=3291
13:22:26,790 [Thread-1] INFO   jTPCC : executeTime[New-Order]=34188
```

You can also look at all result by using `kubectl describe`, it should look like
```yaml
Name:         tpcc-sample
Namespace:    default
Labels:       app.kubernetes.io/created-by=kubebench
              app.kubernetes.io/instance=tpcc-sample
              app.kubernetes.io/managed-by=kustomize
              app.kubernetes.io/name=tpcc
              app.kubernetes.io/part-of=kubebench
Annotations:  <none>
API Version:  benchmark.apecloud.io/v1alpha1
Kind:         Tpcc
Metadata:
  Creation Timestamp:  2023-08-12T13:18:26Z
  Generation:          1
  Resource Version:    129615
  UID:                 73020669-548d-4359-b7f9-123d61a31df0
Spec:
  Delivery:          4
  Duration:          1
  Limit Tx Per Min:  0
  New Order:         45
  Order Status:      4
  Payment:           43
  Step:              all
  Stock Level:       4
  Target:
    Database:  mydb
    Driver:    mysql
    Host:      mycluster-mysql.default.svc.cluster.local
    Password:  ncqsgzb7
    Port:      3306
    User:      root
  Threads:
    1
    2
  Ware Houses:  1
Status:
  Completions:  3/4
  Conditions:
    Last Transition Time:  2023-08-12T13:19:34Z
    Message:               
    Reason:                RecordLog
    Status:                True
    Type:                  tpcc-sample-cleanup-r84qg-Pending
    Last Transition Time:  2023-08-12T13:21:21Z
    Message:               
    Reason:                RecordLog
    Status:                True
    Type:                  tpcc-sample-prepare-2sb2t-Running
    Last Transition Time:  2023-08-12T13:22:30Z
    Message:               13:22:26,783 [Thread-1] INFO   jTPCC : Term-00, Measured tpmC (NewOrders) = 4681.21
                           13:22:26,784 [Thread-1] INFO   jTPCC : Term-00, Measured tpmTOTAL = 10513.54
                           13:22:26,784 [Thread-1] INFO   jTPCC : Term-00, Session Start     = 2023-08-12 13:21:26
                           13:22:26,785 [Thread-1] INFO   jTPCC : Term-00, Session End       = 2023-08-12 13:22:26
                           13:22:26,788 [Thread-1] INFO   jTPCC : Term-00, Transaction Count = 10539
                           13:22:26,790 [Thread-1] INFO   jTPCC : executeTime[Payment]=17522
                           13:22:26,790 [Thread-1] INFO   jTPCC : executeTime[Delivery]=3933
                           13:22:26,790 [Thread-1] INFO   jTPCC : executeTime[Order-Status]=916
                           13:22:26,790 [Thread-1] INFO   jTPCC : executeTime[Stock-Level]=3291
                           13:22:26,790 [Thread-1] INFO   jTPCC : executeTime[New-Order]=34188
    Reason:   RecordLog
    Status:   True
    Type:     tpcc-sample-run-0-8bn52-Running
  Phase:      Running
  Succeeded:  3
  Total:      4
Events:       <none>
```
