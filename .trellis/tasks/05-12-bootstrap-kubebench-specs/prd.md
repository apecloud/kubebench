# Bootstrap Kubebench Trellis specs

## Goal

Replace the generic Trellis backend templates with project-specific coding guidelines for the Kubebench Kubernetes operator.

## Context

Kubebench is a Go/Kubebuilder operator under module `github.com/apecloud/kubebench`. It defines eight namespaced benchmark CRDs in `api/v1alpha1`: `Sysbench`, `Pgbench`, `Ycsb`, `Tpcc`, `Tpch`, `Fio`, `RedisBench`, and `Tpcds`. Each CRD has a spec/status type, `+kubebuilder` validation/printcolumn markers, and a registration call in `init`.

Runtime reconciliation lives in `internal/controller`. Each benchmark usually has two files:

- `<benchmark>_controller.go`: fetch CR, skip terminal phases, build the ordered Job list, create or observe one Job at a time, record pod logs into status conditions, patch status with `client.MergeFrom(old)`, and requeue.
- `<benchmark>_job.go`: build one or more Kubernetes `batchv1.Job` objects using helpers in `internal/utils` and image names from `pkg/constants`.

Shared helpers live in:

- `internal/controllerutil`: small reconcile result helpers and the project-wide one-second requeue interval.
- `internal/utils`: Job existence/status checks, pod log collection, log-to-condition recording, base Job templates, labels, tolerations, resource requests/limits, and database pre-check Jobs.
- `pkg/constants`: benchmark type, driver, step, label, container name, and image environment constants.

Metrics export is separate from the manager. `cmd/exporter/main.go` starts a Gin `/metrics` endpoint, registers Prometheus collectors, tails benchmark log files, updates gauges/counters, waits 30 seconds for Prometheus scrape, and exits. The exporter currently has real parser coverage for sysbench and pgbench in `internal/exporter/*_test.go` with fixtures under `internal/exporter/testdata`.

`cmd/tools/main.go` exposes Cobra subcommands for MySQL, PostgreSQL, MongoDB, and Redis pre-check / database helper operations. The controller Job helpers invoke this binary through `utils.NewPreCheckJob`.

Generated and deployment surfaces:

- `api/v1alpha1/zz_generated.deepcopy.go`, `config/crd/bases`, and generated RBAC are produced by `make generate` / `make manifests`.
- Kustomize config lives under `config/`.
- Helm packaging lives under `deploy/helm/`, including CRDs, dashboards, templates, and values.
- `PROJECT` is generated Kubebuilder scaffold metadata and lists all CRD resources.

## Tools Available

The skill expects GitNexus and ABCoder. Current environment notes:

- `npx gitnexus status` failed because `npx` is not on `PATH`.
- `/Users/buyanbujuan/go/bin/abcoder list-repos` is not supported by this installed ABCoder version.
- `/Users/buyanbujuan/go/bin/abcoder parse go . --no-need-test --repo-id kubebench -o /private/tmp/kubebench_abcoder_ast.json` failed because `go` is not on `PATH`.

Use local source reads, `rg`, and focused file inspection as the fallback for this bootstrap.

## Files to Fill

- `.trellis/spec/backend/index.md`: entry point and checklist for Kubebench backend work.
- `.trellis/spec/backend/directory-structure.md`: package responsibilities and generated-file boundaries.
- `.trellis/spec/backend/crd-api-guidelines.md`: CRD type, status, marker, and generation conventions.
- `.trellis/spec/backend/controller-reconcile-guidelines.md`: Reconciler flow, status patching, log conditions, and requeue behavior.
- `.trellis/spec/backend/job-construction-guidelines.md`: benchmark Job builder patterns, labels, images, resources, pre-checks.
- `.trellis/spec/backend/exporter-metrics-guidelines.md`: Prometheus exporter, parser, tailing, and test fixture conventions.
- `.trellis/spec/backend/cli-tools-guidelines.md`: Cobra database helper CLI conventions.
- `.trellis/spec/backend/config-deployment-guidelines.md`: Kustomize, Helm, generated CRD/RBAC, and build command boundaries.
- `.trellis/spec/backend/error-handling.md`: project-specific Go/controller error propagation.
- `.trellis/spec/backend/logging-guidelines.md`: controller-runtime, klog, and CLI logging conventions.
- `.trellis/spec/backend/quality-guidelines.md`: test and review rules.
- `.trellis/spec/guides/*.md`: cross-cutting thinking guides adapted to Kubebench.

## Important Rules

- Spec files are not fixed. Delete template files that do not apply and create files for real local patterns.
- Do not edit source code while bootstrapping specs.
- Do not edit generated files.
- Use real examples and file paths.
- Do not leave template placeholders.

## Acceptance Criteria

- [ ] Backend spec index reflects actual files.
- [ ] Generic database template is removed or replaced by operator-specific guidance.
- [ ] Specs document CRD, reconciler, Job, exporter, CLI tool, config/deployment, errors, logging, and quality patterns.
- [ ] Thinking guides mention Kubebench data-flow and reuse hotspots.
- [ ] No placeholder text remains under `.trellis/spec`.
- [ ] Verification notes mention unavailable GitNexus/ABCoder/Go tooling.

## Technical Notes

- Language: Go.
- Frameworks: Kubebuilder/controller-runtime, Kubernetes APIs, Ginkgo/Gomega envtest, Prometheus client, Gin, Cobra, Viper.
- Main commands from `Makefile`: `make manifests`, `make generate`, `make fmt`, `make vet`, `make test`, `make run`, `make deploy`.
