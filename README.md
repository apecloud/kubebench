# kubebench
A Kubernetes operator for running benchmark tests on databases to evaluate their performance.

## Installation (Helm)

Installing the kubebench via Helm can be done with the following commands. This requires your machine to have Helm installed. Install Helm

```sh
# add repo
helm repo add kubeblocks https://apecloud.github.io/helm-charts

# install kubebench
helm install kubebench kubeblocks/kubebench --version 0.0.1
```

to delete this release, you can do so with the followint command:
```sh
helm uninstall kubebench
```

## Benchmarks

| Benchmark Name | Use                  | Status    |
| -------------- | -------------------- | --------- |
| [Pgbench](docs/pgbench.md)        | Postgres Performance | Supported |
| [Sysbench](docs/sysbench.md)       | Database Performance | Supported |
| [TPCC](docs/tpcc.md)           | OLTP Performance     | Supported |
| [TPCH](docs/tpch.md)           | DS Performance       | Supported |
| [YCSB](docs/ycsb.md)           | Database Performance | Supported |
| ClickBench     | Database Performance | Planned   |

## License
kubebench is under the Apache License v2.0. See the [LICENSE](LICENSE) file for details.
