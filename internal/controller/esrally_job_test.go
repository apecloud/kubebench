package controller

import (
	"encoding/json"
	"strings"
	"testing"

	benchmarkv1alpha1 "github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/pkg/constants"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestNewEsrallyJobsDefaultGeneratedDataWorkflow(t *testing.T) {
	cr := newEsrallyTestCR()
	cr.Spec.Target.User = "elastic"
	cr.Spec.Target.Password = "secret"

	jobs := NewEsrallyJobs(cr)
	wantNames := []string{"rally-precheck", "rally-cleanup", "rally-prepare", "rally-run"}
	if got := jobNames(jobs); strings.Join(got, ",") != strings.Join(wantNames, ",") {
		t.Fatalf("expected jobs %v, got %v", wantNames, got)
	}
	if got := jobs[0].Spec.Template.Spec.Containers[0].Args; !containsAll(got, []string{"elasticsearch", "ping", "--host", "es.default.svc", "--port", "9200", "--user", "elastic", "--password", "secret"}) {
		t.Fatalf("unexpected precheck args: %#v", got)
	}

	runJob := jobs[3]
	if runJob.Labels[constants.KubeBenchTypeLabel] != constants.EsrallyType {
		t.Fatalf("missing esrally label: %#v", runJob.Labels)
	}
	if len(runJob.Spec.Template.Spec.Containers) != 2 {
		t.Fatalf("expected workload and metrics containers, got %d", len(runJob.Spec.Template.Spec.Containers))
	}
	if got := metricsContainerArg(runJob, "-file"); got != esrallyReportFile {
		t.Fatalf("expected default exporter report file, got %s", got)
	}

	script := runJob.Spec.Template.Spec.Containers[0].Args[0]
	for _, want := range []string{"--pipeline=benchmark-only", "--target-hosts", "--track-path", "--offline", "--challenge", "--track-params", "--client-options", "--report-file"} {
		if !strings.Contains(script, want) {
			t.Fatalf("script missing %s:\n%s", want, script)
		}
	}
	for _, forbidden := range []string{"--track \"$TRACK\"", "--track-repository"} {
		if strings.Contains(script, forbidden) {
			t.Fatalf("script still supports remote Rally track option %s:\n%s", forbidden, script)
		}
	}
	if got := esrallyClientOptions(cr); !strings.Contains(got, "basic_auth_user:'elastic'") || !strings.Contains(got, "basic_auth_password:'secret'") {
		t.Fatalf("unexpected synthesized client options: %s", got)
	}
}

func TestNewEsrallyRunJobsDefaultsMetricsToCSVSharedReport(t *testing.T) {
	cr := newEsrallyTestCR()

	job := NewEsrallyRunJobs(cr)[0]
	if len(job.Spec.Template.Spec.Containers) != 2 {
		t.Fatalf("expected metrics container by default, got %d containers", len(job.Spec.Template.Spec.Containers))
	}
	if got := metricsContainerArg(job, "-file"); got != esrallyReportFile {
		t.Fatalf("expected default report file, got %s", got)
	}
	if got := envValue(job, "REPORT_FORMAT"); got != esrallyReportFormat {
		t.Fatalf("expected default report format env, got %s", got)
	}
	if got := envValue(job, "TRACK_PATH"); got != esrallyGeneratedTrackPath {
		t.Fatalf("expected internal track path env, got %s", got)
	}
	if got := envValue(job, "ON_ERROR"); got != esrallyDefaultOnError {
		t.Fatalf("expected default onError env, got %s", got)
	}
}

func TestNewEsrallyJobsHonorsStepForGeneratedData(t *testing.T) {
	tests := []struct {
		name string
		step string
		want []string
	}{
		{
			name: "default all",
			want: []string{"rally-precheck", "rally-cleanup", "rally-prepare", "rally-run"},
		},
		{
			name: "cleanup",
			step: constants.CleanupStep,
			want: []string{"rally-precheck", "rally-cleanup"},
		},
		{
			name: "prepare",
			step: constants.PrepareStep,
			want: []string{"rally-precheck", "rally-prepare"},
		},
		{
			name: "run",
			step: constants.RunStep,
			want: []string{"rally-precheck", "rally-run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := newEsrallyTestCR()
			cr.Spec.Step = tt.step

			jobs := NewEsrallyJobs(cr)
			if got := jobNames(jobs); strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("expected jobs %v, got %v", tt.want, got)
			}
		})
	}
}

func TestNewEsrallyPrepareJobsGeneratedDataEnvAndScript(t *testing.T) {
	cr := newEsrallyTestCR()
	cr.Spec.DataProfile = constants.EsrallyDataProfileMetrics
	cr.Spec.DocumentCount = 1234
	cr.Spec.Target.Database = "metrics-index"
	cr.Spec.Target.User = "elastic"
	cr.Spec.Target.Password = "secret"

	job := NewEsrallyPrepareJobs(cr)[0]
	if got := envValue(job, "TARGET_URL"); got != "http://es.default.svc:9200" {
		t.Fatalf("expected target URL env, got %s", got)
	}
	if got := envValue(job, "INDEX_NAME"); got != "metrics-index" {
		t.Fatalf("expected generated index env, got %s", got)
	}
	if got := envValue(job, "DATA_PROFILE"); got != constants.EsrallyDataProfileMetrics {
		t.Fatalf("expected metrics data profile env, got %s", got)
	}
	if got := envValue(job, "DOCUMENT_COUNT"); got != "1234" {
		t.Fatalf("expected document count env, got %s", got)
	}
	if got := envValue(job, "TARGET_VERSION"); got != "" {
		t.Fatalf("expected empty target version env by default, got %s", got)
	}

	script := job.Spec.Template.Spec.Containers[0].Args[0]
	for _, want := range []string{"python3 <<'PY'", "/_bulk", "cpu_pct", "memory_mb", "targetVersion 6 or newer", "bulk_index_action", `action["_type"] = "_doc"`, "Generated ESRally dataset is ready"} {
		if !strings.Contains(script, want) {
			t.Fatalf("prepare script missing %s:\n%s", want, script)
		}
	}
}

func TestNewEsrallyPrepareJobsCarriesTargetVersion(t *testing.T) {
	cr := newEsrallyTestCR()
	cr.Spec.TargetVersion = "8.12.2"

	job := NewEsrallyPrepareJobs(cr)[0]
	if got := envValue(job, "TARGET_VERSION"); got != "8.12.2" {
		t.Fatalf("expected target version env, got %s", got)
	}
}

func TestNewEsrallyCleanupJobsDeletesGeneratedIndex(t *testing.T) {
	cr := newEsrallyTestCR()
	cr.Spec.Target.Database = "logs-index"

	job := NewEsrallyCleanupJobs(cr)[0]
	if got := envValue(job, "INDEX_NAME"); got != "logs-index" {
		t.Fatalf("expected cleanup index env, got %s", got)
	}

	script := job.Spec.Template.Spec.Containers[0].Args[0]
	for _, want := range []string{"-X DELETE", "${TARGET_URL}/${INDEX_NAME}", "200|202|404"} {
		if !strings.Contains(script, want) {
			t.Fatalf("cleanup script missing %s:\n%s", want, script)
		}
	}
	if !strings.Contains(script, "targetVersion 6 or newer") {
		t.Fatalf("expected cleanup to validate targetVersion before deleting:\n%s", script)
	}
	if strings.Index(script, "targetVersion 6 or newer") > strings.Index(script, "Deleting generated ESRally index") {
		t.Fatalf("expected targetVersion guard before index deletion:\n%s", script)
	}
}

func TestNewEsrallyPrepareJobsSupportsElasticsearch6TypedBulkMetadata(t *testing.T) {
	cr := newEsrallyTestCR()
	cr.Spec.TargetVersion = "6.8.23"

	job := NewEsrallyPrepareJobs(cr)[0]
	if got := envValue(job, "TARGET_VERSION"); got != "6.8.23" {
		t.Fatalf("expected target version env, got %s", got)
	}

	script := job.Spec.Template.Spec.Containers[0].Args[0]
	for _, want := range []string{`target_major_version == 6`, `action["_type"] = "_doc"`, "bulk_index_action()"} {
		if !strings.Contains(script, want) {
			t.Fatalf("prepare script missing ES6 typed bulk support %s:\n%s", want, script)
		}
	}
}

func TestNewEsrallyRunJobsUsesInternalGeneratedTrack(t *testing.T) {
	cr := newEsrallyTestCR()

	job := NewEsrallyRunJobs(cr)[0]
	if got := envValue(job, "TRACK_PATH"); got != esrallyGeneratedTrackPath {
		t.Fatalf("expected internal track path env, got %s", got)
	}
	if got := envValue(job, "CHALLENGE"); got != esrallyGeneratedChallenge {
		t.Fatalf("expected internal challenge env, got %s", got)
	}
	script := job.Spec.Template.Spec.Containers[0].Args[0]
	if strings.Contains(script, "spec."+"trackPath") {
		t.Fatalf("run script leaked removed public track path API field:\n%s", script)
	}
}

func TestEsrallyGeneratedDataDefaults(t *testing.T) {
	cr := newEsrallyTestCR()

	if got := esrallyDataProfile(cr); got != constants.EsrallyDataProfileLogs {
		t.Fatalf("expected default data profile logs, got %s", got)
	}
	if got := esrallyDocumentCount(cr); got != esrallyDefaultDocs {
		t.Fatalf("expected default document count %d, got %d", esrallyDefaultDocs, got)
	}
	if got := esrallyIndexName(cr); got != esrallyDefaultIndex {
		t.Fatalf("expected default index %s, got %s", esrallyDefaultIndex, got)
	}
}

func TestNewEsrallyJobsAlwaysPrechecksBasicTarget(t *testing.T) {
	cr := newEsrallyTestCR()
	cr.Spec.Step = constants.RunStep

	jobs := NewEsrallyJobs(cr)
	if len(jobs) != 2 {
		t.Fatalf("expected precheck and run jobs, got %d", len(jobs))
	}
	if jobs[0].Name != "rally-precheck" || jobs[1].Name != "rally-run" {
		t.Fatalf("expected precheck then run jobs, got %v", jobNames(jobs))
	}
	if got := esrallyTargetHosts(cr); got != "es.default.svc:9200" {
		t.Fatalf("unexpected target hosts: %s", got)
	}
}

func TestNewEsrallyRunJobsUsesTargetHostAndPort(t *testing.T) {
	cr := newEsrallyTestCR()

	job := NewEsrallyRunJobs(cr)[0]
	if esrallyTargetHosts(cr) != "es.default.svc:9200" {
		t.Fatalf("unexpected target hosts: %s", esrallyTargetHosts(cr))
	}
	if job.Spec.Template.Spec.Volumes[1].EmptyDir == nil {
		t.Fatalf("expected rally-home emptyDir volume: %#v", job.Spec.Template.Spec.Volumes)
	}
	script := job.Spec.Template.Spec.Containers[0].Args[0]
	for _, want := range []string{"--target-hosts", "--track-path", "--offline"} {
		if !strings.Contains(script, want) {
			t.Fatalf("script missing %s:\n%s", want, script)
		}
	}
	for _, forbidden := range []string{"--test-mode", "--telemetry-params"} {
		if strings.Contains(script, forbidden) {
			t.Fatalf("script still exposes removed Rally option %s:\n%s", forbidden, script)
		}
	}
}

func TestNewEsrallyRunJobsAddsTargetVersionTrackParam(t *testing.T) {
	cr := newEsrallyTestCR()
	cr.Spec.TargetVersion = " 8.12.2 "

	job := NewEsrallyRunJobs(cr)[0]
	if got := envValue(job, "TARGET_VERSION"); got != "8.12.2" {
		t.Fatalf("expected trimmed target version env, got %q", got)
	}

	params := trackParamsFromEnv(t, job)
	if got := params["target_index"]; got != "kubebench" {
		t.Fatalf("expected existing track param to remain, got %q", got)
	}
	if got := params["target_version"]; got != "8.12.2" {
		t.Fatalf("expected target_version track param, got %q in %#v", got, params)
	}

	script := job.Spec.Template.Spec.Containers[0].Args[0]
	if strings.Contains(script, "--distribution-version") {
		t.Fatalf("benchmark-only run script must not pass --distribution-version:\n%s", script)
	}
}

func TestNewEsrallyRunJobsProductionOptions(t *testing.T) {
	tests := []struct {
		name            string
		mutate          func(*benchmarkv1alpha1.Esrally)
		wantContainers  int
		wantEnv         map[string]string
		wantScriptParts []string
		wantMetricFile  string
	}{
		{
			name:           "defaults keep csv metrics contract",
			wantContainers: 2,
			wantEnv: map[string]string{
				"TRACK_PATH":    esrallyGeneratedTrackPath,
				"CHALLENGE":     esrallyGeneratedChallenge,
				"ON_ERROR":      esrallyDefaultOnError,
				"REPORT_FORMAT": esrallyReportFormat,
				"REPORT_FILE":   esrallyReportFile,
			},
			wantScriptParts: []string{"--pipeline=benchmark-only", "--track-path", "--offline", "--report-format", "--report-file"},
			wantMetricFile:  esrallyReportFile,
		},
		{
			name: "telemetry and extra args are wired through",
			mutate: func(cr *benchmarkv1alpha1.Esrally) {
				cr.Spec.Telemetry = []benchmarkv1alpha1.EsrallyTelemetry{
					benchmarkv1alpha1.EsrallyTelemetryNodeStats,
					benchmarkv1alpha1.EsrallyTelemetryDiskUsageStats,
				}
				cr.Spec.ExtraArgs = []string{"--kill-running-processes", "--enable-driver-profiling"}
			},
			wantContainers: 2,
			wantEnv: map[string]string{
				"TELEMETRY":  "node-stats,disk-usage-stats",
				"EXTRA_ARGS": "--kill-running-processes --enable-driver-profiling",
			},
			wantScriptParts: []string{"--telemetry", "$EXTRA_ARGS"},
			wantMetricFile:  esrallyReportFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := newEsrallyTestCR()
			if tt.mutate != nil {
				tt.mutate(cr)
			}

			job := NewEsrallyRunJobs(cr)[0]
			if len(job.Spec.Template.Spec.Containers) != tt.wantContainers {
				t.Fatalf("expected %d containers, got %d", tt.wantContainers, len(job.Spec.Template.Spec.Containers))
			}
			for name, want := range tt.wantEnv {
				if got := envValue(job, name); got != want {
					t.Fatalf("expected env %s=%q, got %q", name, want, got)
				}
			}
			script := job.Spec.Template.Spec.Containers[0].Args[0]
			for _, want := range tt.wantScriptParts {
				if !strings.Contains(script, want) {
					t.Fatalf("script missing %s:\n%s", want, script)
				}
			}
			if tt.wantMetricFile != "" {
				if got := metricsContainerArg(job, "-file"); got != tt.wantMetricFile {
					t.Fatalf("expected exporter file %s, got %s", tt.wantMetricFile, got)
				}
				if got := metricsContainerArg(job, "-done-file"); got != esrallyExitFile {
					t.Fatalf("expected exporter done file %s, got %s", esrallyExitFile, got)
				}
			}
		})
	}
}

func TestNewEsrallyRunJobsSharesReportVolumeWithExporter(t *testing.T) {
	cr := newEsrallyTestCR()
	cr.Spec.Step = constants.RunStep

	job := NewEsrallyJobs(cr)[1]
	workload := job.Spec.Template.Spec.Containers[0]
	metrics := containerByName(job, "metrics")
	if metrics == nil {
		t.Fatal("expected metrics container")
	}
	if !hasVolumeMount(workload.VolumeMounts, "log", "/var/log") {
		t.Fatalf("workload missing shared log volume mount: %#v", workload.VolumeMounts)
	}
	if !hasVolumeMount(metrics.VolumeMounts, "log", "/var/log") {
		t.Fatalf("metrics container missing shared log volume mount: %#v", metrics.VolumeMounts)
	}
	if job.Labels[constants.KubeBenchNameLabel] != "rally" || job.Spec.Template.Labels[constants.KubeBenchNameLabel] != "rally" {
		t.Fatalf("expected kubebench name labels on job and pod template: job=%#v template=%#v", job.Labels, job.Spec.Template.Labels)
	}
	if job.Labels[constants.KubeBenchTypeLabel] != constants.EsrallyType || job.Spec.Template.Labels[constants.KubeBenchTypeLabel] != constants.EsrallyType {
		t.Fatalf("expected kubebench type labels on job and pod template: job=%#v template=%#v", job.Labels, job.Spec.Template.Labels)
	}
}

func containsAll(values []string, wants []string) bool {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		seen[value] = true
	}
	for _, want := range wants {
		if !seen[want] {
			return false
		}
	}
	return true
}

func envValue(job *batchv1.Job, name string) string {
	for _, env := range job.Spec.Template.Spec.Containers[0].Env {
		if env.Name == name {
			return env.Value
		}
	}
	return ""
}

func containerByName(job *batchv1.Job, name string) *corev1.Container {
	for i := range job.Spec.Template.Spec.Containers {
		if job.Spec.Template.Spec.Containers[i].Name == name {
			return &job.Spec.Template.Spec.Containers[i]
		}
	}
	return nil
}

func hasVolumeMount(mounts []corev1.VolumeMount, name, mountPath string) bool {
	for _, mount := range mounts {
		if mount.Name == name && mount.MountPath == mountPath {
			return true
		}
	}
	return false
}

func metricsContainerArg(job *batchv1.Job, name string) string {
	for _, container := range job.Spec.Template.Spec.Containers {
		if container.Name != "metrics" {
			continue
		}
		for i, arg := range container.Args {
			if arg == name && i+1 < len(container.Args) {
				return container.Args[i+1]
			}
		}
	}
	return ""
}

func trackParamsFromEnv(t *testing.T, job *batchv1.Job) map[string]string {
	t.Helper()
	return parseTrackParams(t, envValue(job, "TRACK_PARAMS"))
}

func parseTrackParams(t *testing.T, value string) map[string]string {
	t.Helper()
	params := map[string]string{}
	if err := json.Unmarshal([]byte(value), &params); err != nil {
		t.Fatalf("failed to parse track params %q: %v", value, err)
	}
	return params
}

func newEsrallyTestCR() *benchmarkv1alpha1.Esrally {
	cr := &benchmarkv1alpha1.Esrally{}
	cr.Name = "rally"
	cr.Namespace = "default"
	cr.Spec.Target.Driver = constants.ElasticsearchDriver
	cr.Spec.Target.Host = "es.default.svc"
	cr.Spec.Target.Port = 9200
	return cr
}

func jobNames(jobs []*batchv1.Job) []string {
	names := make([]string, 0, len(jobs))
	for _, job := range jobs {
		names = append(names, job.Name)
	}
	return names
}
