package controller

import (
	"strings"
	"testing"

	benchmarkv1alpha1 "github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/pkg/constants"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestNewEsrallyJobs(t *testing.T) {
	cr := &benchmarkv1alpha1.Esrally{}
	cr.Name = "rally"
	cr.Namespace = "default"
	cr.Spec.Target.Driver = constants.ElasticsearchDriver
	cr.Spec.Target.Host = "es.default.svc"
	cr.Spec.Target.Port = 9200
	cr.Spec.Target.User = "elastic"
	cr.Spec.Target.Password = "secret"
	cr.Spec.Track = "geonames"
	cr.Spec.Challenge = "append-no-conflicts"
	cr.Spec.IncludeTasks = []string{"index-append"}
	cr.Spec.TrackParams = map[string]string{"number_of_shards": "3"}

	jobs := NewEsrallyJobs(cr)
	if len(jobs) != 2 {
		t.Fatalf("expected precheck and run jobs, got %d", len(jobs))
	}
	if jobs[0].Name != "rally-precheck" {
		t.Fatalf("expected precheck first, got %s", jobs[0].Name)
	}
	if got := jobs[0].Spec.Template.Spec.Containers[0].Args; !containsAll(got, []string{"elasticsearch", "ping", "--host", "es.default.svc", "--port", "9200", "--user", "elastic", "--password", "secret"}) {
		t.Fatalf("unexpected precheck args: %#v", got)
	}
	if jobs[1].Name != "rally-run" {
		t.Fatalf("expected run job second, got %s", jobs[1].Name)
	}
	if jobs[1].Labels[constants.KubeBenchTypeLabel] != constants.EsrallyType {
		t.Fatalf("missing esrally label: %#v", jobs[1].Labels)
	}
	if len(jobs[1].Spec.Template.Spec.Containers) != 2 {
		t.Fatalf("expected workload and metrics containers, got %d", len(jobs[1].Spec.Template.Spec.Containers))
	}
	if got := metricsContainerArg(jobs[1], "-file"); got != esrallyReportFile {
		t.Fatalf("expected default exporter report file, got %s", got)
	}

	script := jobs[1].Spec.Template.Spec.Containers[0].Args[0]
	for _, want := range []string{"--pipeline=benchmark-only", "--target-hosts", "--track", "--challenge", "--include-tasks", "--track-params", "--client-options", "--report-file"} {
		if !strings.Contains(script, want) {
			t.Fatalf("script missing %s:\n%s", want, script)
		}
	}
	if got := esrallyClientOptions(cr); !strings.Contains(got, "basic_auth_user:'elastic'") || !strings.Contains(got, "basic_auth_password:'secret'") {
		t.Fatalf("unexpected synthesized client options: %s", got)
	}
}

func TestNewEsrallyRunJobsDefaultsMetricsToCSVSharedReport(t *testing.T) {
	cr := &benchmarkv1alpha1.Esrally{}
	cr.Name = "rally"
	cr.Namespace = "default"
	cr.Spec.Target.Host = "es.default.svc"
	cr.Spec.Target.Port = 9200

	jobs := NewEsrallyRunJobs(cr)
	job := jobs[0]
	if len(job.Spec.Template.Spec.Containers) != 2 {
		t.Fatalf("expected metrics container by default, got %d containers", len(job.Spec.Template.Spec.Containers))
	}
	if got := metricsContainerArg(job, "-file"); got != esrallyReportFile {
		t.Fatalf("expected default report file, got %s", got)
	}
	if got := envValue(job, "REPORT_FORMAT"); got != esrallyReportFormat {
		t.Fatalf("expected default report format env, got %s", got)
	}
	if got := envValue(job, "TRACK"); got != esrallyDefaultTrack {
		t.Fatalf("expected default track env, got %s", got)
	}
	if got := envValue(job, "ON_ERROR"); got != esrallyDefaultOnError {
		t.Fatalf("expected default onError env, got %s", got)
	}
}

func TestNewEsrallyJobsSkipsPrecheckForAdvancedRallyClientOptions(t *testing.T) {
	cr := &benchmarkv1alpha1.Esrally{}
	cr.Name = "rally"
	cr.Namespace = "default"
	cr.Spec.Target.Host = "es.default.svc"
	cr.Spec.Target.Port = 9200
	cr.Spec.ClientOptions = "use_ssl:true,verify_certs:false,api_key:'secret'"

	jobs := NewEsrallyJobs(cr)
	if len(jobs) != 1 {
		t.Fatalf("expected only run job, got %d", len(jobs))
	}
	if jobs[0].Name != "rally-run" {
		t.Fatalf("expected run job, got %s", jobs[0].Name)
	}
	if got := esrallyClientOptions(cr); got != cr.Spec.ClientOptions {
		t.Fatalf("expected explicit client options to pass through")
	}
}

func TestNewEsrallyJobsSkipsPrecheckForTargetHosts(t *testing.T) {
	cr := &benchmarkv1alpha1.Esrally{}
	cr.Name = "rally"
	cr.Namespace = "default"
	cr.Spec.Target.Host = "ignored"
	cr.Spec.Target.Port = 9200
	cr.Spec.TargetHosts = []string{"es-0:9200", "es-1:9200/prefix"}

	jobs := NewEsrallyJobs(cr)
	if len(jobs) != 1 {
		t.Fatalf("expected only run job, got %d", len(jobs))
	}
	if jobs[0].Name != "rally-run" {
		t.Fatalf("expected run job, got %s", jobs[0].Name)
	}
	if got := esrallyTargetHosts(cr); got != "es-0:9200,es-1:9200/prefix" {
		t.Fatalf("unexpected target hosts: %s", got)
	}
}

func TestNewEsrallyRunJobsWithTrackPathAndTargetHosts(t *testing.T) {
	cr := &benchmarkv1alpha1.Esrally{}
	cr.Name = "rally"
	cr.Namespace = "default"
	cr.Spec.Target.Host = "ignored"
	cr.Spec.Target.Port = 9200
	cr.Spec.TargetHosts = []string{"es-0:9200", "es-1:9200"}
	cr.Spec.TrackPath = "/rally/.rally/tracks/custom"
	cr.Spec.Offline = true
	cr.Spec.TestMode = true

	jobs := NewEsrallyRunJobs(cr)
	job := jobs[0]
	if esrallyTargetHosts(cr) != "es-0:9200,es-1:9200" {
		t.Fatalf("unexpected target hosts: %s", esrallyTargetHosts(cr))
	}
	if job.Spec.Template.Spec.Volumes[1].EmptyDir == nil {
		t.Fatalf("expected rally-home emptyDir volume: %#v", job.Spec.Template.Spec.Volumes)
	}
	script := job.Spec.Template.Spec.Containers[0].Args[0]
	for _, want := range []string{"--track-path", "--offline", "--test-mode"} {
		if !strings.Contains(script, want) {
			t.Fatalf("script missing %s:\n%s", want, script)
		}
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
				"TRACK":         esrallyDefaultTrack,
				"ON_ERROR":      esrallyDefaultOnError,
				"REPORT_FORMAT": esrallyReportFormat,
				"REPORT_FILE":   esrallyReportFile,
			},
			wantScriptParts: []string{"--pipeline=benchmark-only", "--report-format", "--report-file"},
			wantMetricFile:  esrallyReportFile,
		},
		{
			name: "telemetry telemetry params and extra args are wired through",
			mutate: func(cr *benchmarkv1alpha1.Esrally) {
				cr.Spec.Telemetry = []string{"node-stats", "disk-usage-stats"}
				cr.Spec.TelemetryParams = "node-stats-sample-interval:1"
				cr.Spec.ExtraArgs = []string{"--kill-running-processes", "--enable-driver-profiling"}
			},
			wantContainers: 2,
			wantEnv: map[string]string{
				"TELEMETRY":        "node-stats,disk-usage-stats",
				"TELEMETRY_PARAMS": "node-stats-sample-interval:1",
				"EXTRA_ARGS":       "--kill-running-processes --enable-driver-profiling",
			},
			wantScriptParts: []string{"--telemetry", "--telemetry-params", "$EXTRA_ARGS"},
			wantMetricFile:  esrallyReportFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &benchmarkv1alpha1.Esrally{}
			cr.Name = "rally"
			cr.Namespace = "default"
			cr.Spec.Target.Host = "es.default.svc"
			cr.Spec.Target.Port = 9200
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
	cr := &benchmarkv1alpha1.Esrally{}
	cr.Name = "rally"
	cr.Namespace = "default"
	cr.Spec.Target.Host = "es.default.svc"
	cr.Spec.Target.Port = 9200
	cr.Spec.ClientOptions = "use_ssl:false"

	job := NewEsrallyJobs(cr)[0]
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
