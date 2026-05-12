package controller

import (
	"strings"
	"testing"

	benchmarkv1alpha1 "github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/pkg/constants"
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
	cr.Spec.Metrics = true

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
