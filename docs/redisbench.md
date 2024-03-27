# Redis Benchmark

[redis-benchmark](https://redis.io/docs/management/optimization/benchmarks/) is the utility to check the performance of Redis by running n commands simultaneously.

## Running Redis Benchmark

your resource file look like this:

```yaml
apiVersion: benchmark.apecloud.io/v1alpha1
kind: Redisbench
metadata:
  labels:
    app.kubernetes.io/name: redisbench
    app.kubernetes.io/instance: redisbench-sample
    app.kubernetes.io/part-of: kubebench
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: kubebench
  name: redisbench-sample
spec:
  clients:
    - 2
    - 4
  requests: 100000
  dataSize: 3
  pipeline: 16
  target:
    host: "test-redis-redis.default.svc.cluster.local"
    port: 6379
    user: "xxx"
    password: "xxx"
```

Once done creating/editing the resource file, you can run it by:

```sh
# kubectl apply -f config/samples/config/samples/benchmark_v1alpha1_redisbench.yaml # if edited the original one
# kubectl apply -f <path_to_file> # if created a new cr file
```

Deploying the `cr.yaml` would reuslt in:
```sh
# kubectl get redisbenches.benchmark.apecloud.io
NAME                     STATUS    COMPLETIONS   AGE
redis-benchmark-3m0qu0   Running   1/2           60s

# kubectl get pod
NAME                                    READY   STATUS      RESTARTS        AGE
redis-benchmark-3m0qu0-precheck-x9pdk   0/1     Completed   0               58s
redis-benchmark-3m0qu0-0-run-gc8dk      1/1     Running     0               54s
```

You can look at a result by using `kubectl log`, it should look like:
```sh
PING_INLINE: 221190.00 requests per second, p50=0.063 msec                    
PING_MBULK: 233426.70 requests per second, p50=0.063 msec                    
SET: 165562.92 requests per second, p50=0.087 msec                    
GET: 224719.11 requests per second, p50=0.063 msec                    
INCR: 171703.30 requests per second, p50=0.079 msec                    
LPUSH: 173400.38 requests per second, p50=0.079 msec                    
RPUSH: 186046.52 requests per second, p50=0.079 msec                    
LPOP: 191021.97 requests per second, p50=0.079 msec                    
RPOP: 197199.77 requests per second, p50=0.071 msec                    
SADD: 219538.97 requests per second, p50=0.063 msec                    
HSET: 151791.14 requests per second, p50=0.095 msec                    
SPOP: 218818.38 requests per second, p50=0.063 msec                    
ZADD: 199084.22 requests per second, p50=0.071 msec                    
ZPOPMIN: 205338.81 requests per second, p50=0.063 msec                    
LPUSH (needed to benchmark LRANGE): 178284.89 requests per second, p50=0.079 msec                    
LRANGE_100 (first 100 elements): 59747.87 requests per second, p50=0.143 msec                   
LRANGE_300 (first 300 elements): 22757.27 requests per second, p50=0.295 msec                   
LRANGE_500 (first 500 elements): 14898.02 requests per second, p50=0.439 msec                   
LRANGE_600 (first 600 elements): 12245.45 requests per second, p50=0.519 msec                   
MSET (10 keys): 70111.48 requests per second, p50=0.215 msec                
```
