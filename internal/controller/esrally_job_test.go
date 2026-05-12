package controller

import (
	"strings"
	"testing"

	benchmarkv1alpha1 "github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/pkg/constants"
	batchv1 "k8s.io/api/batch/v1"
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
	cr.Spec.Metrics = boolPtr(true)

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
	if got := metricsContainerArg(jobs[1], "-file"); got != defaultEsrallyReportFile {
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
	if got := metricsContainerArg(job, "-file"); got != defaultEsrallyReportFile {
		t.Fatalf("expected default report file, got %s", got)
	}
	if got := envValue(job, "REPORT_FORMAT"); got != defaultEsrallyReportFormat {
		t.Fatalf("expected default report format env, got %s", got)
	}
	if got := envValue(job, "KUBEBENCH_METRICS_UNAVAILABLE"); got != "" {
		t.Fatalf("expected metrics to be available, got reason %q", got)
	}
}

func TestNewEsrallyRunJobsDisablesMetricsExplicitly(t *testing.T) {
	cr := &benchmarkv1alpha1.Esrally{}
	cr.Name = "rally"
	cr.Namespace = "default"
	cr.Spec.Target.Host = "es.default.svc"
	cr.Spec.Target.Port = 9200
	cr.Spec.Metrics = boolPtr(false)

	jobs := NewEsrallyRunJobs(cr)
	job := jobs[0]
	if len(job.Spec.Template.Spec.Containers) != 1 {
		t.Fatalf("expected only workload container when metrics are disabled, got %d", len(job.Spec.Template.Spec.Containers))
	}
	if got := envValue(job, "KUBEBENCH_METRICS_UNAVAILABLE"); got != "kubebench metrics unavailable: spec.metrics is false" {
		t.Fatalf("expected disabled metrics reason, got %q", got)
	}
}

func TestNewEsrallyRunJobsDisablesExporterForMarkdownReport(t *testing.T) {
	cr := &benchmarkv1alpha1.Esrally{}
	cr.Name = "rally"
	cr.Namespace = "default"
	cr.Spec.Target.Host = "es.default.svc"
	cr.Spec.Target.Port = 9200
	cr.Spec.ReportFormat = "markdown"
	cr.Spec.ReportFile = "/var/log/esrally-report.md"
	cr.Spec.Metrics = boolPtr(false)

	jobs := NewEsrallyRunJobs(cr)
	job := jobs[0]
	if len(job.Spec.Template.Spec.Containers) != 1 {
		t.Fatalf("expected no metrics container for markdown report, got %d containers", len(job.Spec.Template.Spec.Containers))
	}
	script := job.Spec.Template.Spec.Containers[0].Args[0]
	if !strings.Contains(script, "Rally $REPORT_FORMAT report (kubebench metrics unavailable):") {
		t.Fatalf("script should make markdown metrics unavailability explicit:\n%s", script)
	}
	if got := envValue(job, "KUBEBENCH_METRICS_UNAVAILABLE"); got != "kubebench metrics unavailable: spec.metrics is false" {
		t.Fatalf("expected disabled metrics reason, got %q", got)
	}
}

func TestNewEsrallyRunJobsRequiresSharedCSVReportForMetrics(t *testing.T) {
	tests := []struct {
		name       string
		reportFile string
		wantMetric bool
		wantReason string
	}{
		{
			name:       "custom shared CSV report",
			reportFile: "/var/log/custom-rally.csv",
			wantMetric: true,
		},
		{
			name:       "non shared CSV report",
			reportFile: "/tmp/custom-rally.csv",
			wantMetric: false,
			wantReason: "kubebench metrics unavailable: reportFile must be under /var/log for the exporter shared volume",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &benchmarkv1alpha1.Esrally{}
			cr.Name = "rally"
			cr.Namespace = "default"
			cr.Spec.Target.Host = "es.default.svc"
			cr.Spec.Target.Port = 9200
			cr.Spec.ReportFormat = "csv"
			cr.Spec.ReportFile = tt.reportFile
			cr.Spec.Metrics = boolPtr(true)

			jobs := NewEsrallyRunJobs(cr)
			job := jobs[0]
			if gotMetric := len(job.Spec.Template.Spec.Containers) == 2; gotMetric != tt.wantMetric {
				t.Fatalf("expected metrics container=%t, got %t", tt.wantMetric, gotMetric)
			}
			if tt.wantMetric {
				if got := metricsContainerArg(job, "-file"); got != tt.reportFile {
					t.Fatalf("expected exporter file %s, got %s", tt.reportFile, got)
				}
			}
			if got := envValue(job, "KUBEBENCH_METRICS_UNAVAILABLE"); got != tt.wantReason {
				t.Fatalf("expected reason %q, got %q", tt.wantReason, got)
			}
		})
	}
}

func TestNewEsrallyRunJobsDoesNotStartExporterForInvalidMetricsFormat(t *testing.T) {
	cr := &benchmarkv1alpha1.Esrally{}
	cr.Name = "rally"
	cr.Namespace = "default"
	cr.Spec.Target.Host = "es.default.svc"
	cr.Spec.Target.Port = 9200
	cr.Spec.ReportFormat = "markdown"
	cr.Spec.ReportFile = "/var/log/esrally-report.md"
	cr.Spec.Metrics = boolPtr(true)

	jobs := NewEsrallyRunJobs(cr)
	job := jobs[0]
	if len(job.Spec.Template.Spec.Containers) != 1 {
		t.Fatalf("expected no metrics container for non-csv report, got %d containers", len(job.Spec.Template.Spec.Containers))
	}
	if got := envValue(job, "KUBEBENCH_METRICS_UNAVAILABLE"); got != "kubebench metrics unavailable: the exporter only supports reportFormat csv" {
		t.Fatalf("expected unsupported format reason, got %q", got)
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

func TestNewEsrallyRunJobsWithTrackPathPVCAndTargetHosts(t *testing.T) {
	cr := &benchmarkv1alpha1.Esrally{}
	cr.Name = "rally"
	cr.Namespace = "default"
	cr.Spec.Target.Host = "ignored"
	cr.Spec.Target.Port = 9200
	cr.Spec.TargetHosts = []string{"es-0:9200", "es-1:9200"}
	cr.Spec.TrackPath = "/rally/.rally/tracks/custom"
	cr.Spec.RallyHomePVCClaimName = "rally-home"
	cr.Spec.Offline = true
	cr.Spec.TestMode = true

	jobs := NewEsrallyRunJobs(cr)
	job := jobs[0]
	if esrallyTargetHosts(cr) != "es-0:9200,es-1:9200" {
		t.Fatalf("unexpected target hosts: %s", esrallyTargetHosts(cr))
	}
	if job.Spec.Template.Spec.Volumes[1].PersistentVolumeClaim == nil || job.Spec.Template.Spec.Volumes[1].PersistentVolumeClaim.ClaimName != "rally-home" {
		t.Fatalf("expected rally-home PVC volume: %#v", job.Spec.Template.Spec.Volumes)
	}
	script := job.Spec.Template.Spec.Containers[0].Args[0]
	for _, want := range []string{"--track-path", "--offline", "--test-mode"} {
		if !strings.Contains(script, want) {
			t.Fatalf("script missing %s:\n%s", want, script)
		}
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

func boolPtr(value bool) *bool {
	return &value
}

func envValue(job *batchv1.Job, name string) string {
	for _, env := range job.Spec.Template.Spec.Containers[0].Env {
		if env.Name == name {
			return env.Value
		}
	}
	return ""
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
