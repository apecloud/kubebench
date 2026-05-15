package controller

import (
	"encoding/json"
	"fmt"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/internal/utils"
	"github.com/apecloud/kubebench/pkg/constants"
)

const (
	esrallyLogFile        = "/var/log/esrally.log"
	esrallyExitFile       = "/var/log/esrally.exit"
	esrallyHomeMountPath  = "/rally/.rally"
	esrallyDefaultOnError = "abort"
	esrallyReportFormat   = "csv"
	esrallyReportFile     = "/var/log/esrally-report.csv"
	esrallyDefaultIndex   = "kubebench"
	esrallyDefaultDocs    = 10000
)

func NewEsrallyJobs(cr *v1alpha1.Esrally) []*batchv1.Job {
	workJobs := newEsrallyWorkJobs(cr)
	jobs := make([]*batchv1.Job, 0, len(workJobs)+1)

	if len(workJobs) > 0 {
		if job := newEsrallyPreCheckJob(cr); job != nil {
			jobs = append(jobs, job)
		}
	}
	jobs = append(jobs, workJobs...)

	utils.AddTolerationToJobs(jobs, cr.Spec.Tolerations)
	utils.AddLabelsToJobs(jobs, cr.Labels)
	utils.AddLabelsToJobs(jobs, map[string]string{
		constants.KubeBenchNameLabel: cr.Name,
		constants.KubeBenchTypeLabel: constants.EsrallyType,
	})
	utils.AddResourceLimitsToJobs(jobs, cr.Spec.ResourceLimits)
	utils.AddResourceRequestsToJobs(jobs, cr.Spec.ResourceRequests)

	return jobs
}

func newEsrallyWorkJobs(cr *v1alpha1.Esrally) []*batchv1.Job {
	jobs := make([]*batchv1.Job, 0)
	step := esrallyStep(cr)

	if step == constants.CleanupStep || step == constants.AllStep {
		jobs = append(jobs, NewEsrallyCleanupJobs(cr)...)
	}
	if step == constants.PrepareStep || step == constants.AllStep {
		jobs = append(jobs, NewEsrallyPrepareJobs(cr)...)
	}
	if step == constants.RunStep || step == constants.AllStep {
		jobs = append(jobs, NewEsrallyRunJobs(cr)...)
	}

	return jobs
}

func newEsrallyPreCheckJob(cr *v1alpha1.Esrally) *batchv1.Job {
	if cr.Spec.ClientOptions != "" || len(cr.Spec.TargetHosts) > 0 {
		return nil
	}
	return utils.NewPreCheckJob(cr.Name, cr.Namespace, constants.ElasticsearchDriver, &cr.Spec.Target)
}

func NewEsrallyCleanupJobs(cr *v1alpha1.Esrally) []*batchv1.Job {
	job := utils.JobTemplate(fmt.Sprintf("%s-cleanup", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchEnvEsrally),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c"},
			Args:            []string{esrallyCleanupScript()},
			Env:             esrallyGeneratedDataEnv(cr),
			VolumeMounts: []corev1.VolumeMount{
				{Name: "log", MountPath: "/var/log"},
			},
		},
	)
	return []*batchv1.Job{job}
}

func NewEsrallyPrepareJobs(cr *v1alpha1.Esrally) []*batchv1.Job {
	job := utils.JobTemplate(fmt.Sprintf("%s-prepare", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchEnvEsrally),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c"},
			Args:            []string{esrallyPrepareScript()},
			Env:             esrallyGeneratedDataEnv(cr),
			VolumeMounts: []corev1.VolumeMount{
				{Name: "log", MountPath: "/var/log"},
			},
		},
	)
	return []*batchv1.Job{job}
}

func NewEsrallyRunJobs(cr *v1alpha1.Esrally) []*batchv1.Job {
	jobName := fmt.Sprintf("%s-run", cr.Name)
	job := utils.JobTemplate(jobName, cr.Namespace)
	addEsrallyHomeVolume(job)

	env := []corev1.EnvVar{
		{Name: "TARGET_HOSTS", Value: esrallyTargetHosts(cr)},
		{Name: "TRACK_PATH", Value: cr.Spec.TrackPath},
		{Name: "CHALLENGE", Value: cr.Spec.Challenge},
		{Name: "INCLUDE_TASKS", Value: strings.Join(cr.Spec.IncludeTasks, ",")},
		{Name: "TRACK_PARAMS", Value: esrallyTrackParams(cr.Spec.TrackParams)},
		{Name: "CLIENT_OPTIONS", Value: esrallyClientOptions(cr)},
		{Name: "ON_ERROR", Value: esrallyOnError(cr)},
		{Name: "TELEMETRY", Value: strings.Join(cr.Spec.Telemetry, ",")},
		{Name: "TELEMETRY_PARAMS", Value: cr.Spec.TelemetryParams},
		{Name: "REPORT_FORMAT", Value: esrallyReportFormat},
		{Name: "REPORT_FILE", Value: esrallyReportFile},
		{Name: "EXTRA_ARGS", Value: strings.Join(cr.Spec.ExtraArgs, " ")},
	}

	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchEnvEsrally),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c"},
			Args:            []string{esrallyRunScript(cr)},
			Env:             env,
			VolumeMounts: []corev1.VolumeMount{
				{Name: "log", MountPath: "/var/log"},
				{Name: "rally-home", MountPath: esrallyHomeMountPath},
			},
		},
	)

	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            "metrics",
			Image:           constants.GetBenchmarkImage(constants.KubebenchExporter),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Ports: []corev1.ContainerPort{
				{
					ContainerPort: 9187,
					Name:          "http-metrics",
					Protocol:      corev1.ProtocolTCP,
				},
			},
			Command: []string{"/exporter"},
			Args: []string{
				"-type", constants.EsrallyType,
				"-file", esrallyReportFile,
				"-bench", cr.Name,
				"-job", jobName,
				"-done-file", esrallyExitFile,
			},
			VolumeMounts: []corev1.VolumeMount{
				{Name: "log", MountPath: "/var/log"},
			},
		},
	)

	return []*batchv1.Job{job}
}

func addEsrallyHomeVolume(job *batchv1.Job) {
	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "rally-home",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})
}

func esrallyGeneratedDataEnv(cr *v1alpha1.Esrally) []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: "TARGET_URL", Value: esrallyTargetURL(cr)},
		{Name: "INDEX_NAME", Value: esrallyIndexName(cr)},
		{Name: "DATA_PROFILE", Value: esrallyDataProfile(cr)},
		{Name: "DOCUMENT_COUNT", Value: fmt.Sprintf("%d", esrallyDocumentCount(cr))},
		{Name: "ES_USERNAME", Value: cr.Spec.Target.User},
		{Name: "ES_PASSWORD", Value: cr.Spec.Target.Password},
		{Name: "CLIENT_OPTIONS", Value: cr.Spec.ClientOptions},
		{Name: "TARGET_HOSTS_OVERRIDE", Value: strings.Join(cr.Spec.TargetHosts, ",")},
	}
}

func esrallyCleanupScript() string {
	return strings.Join([]string{
		`set -eu`,
		esrallyGeneratedDataUnsupportedConfigCheck(),
		`echo "Deleting generated ESRally index ${INDEX_NAME} from ${TARGET_URL}" | tee -a "` + esrallyLogFile + `"`,
		`: > /tmp/esrally-cleanup.out`,
		`if [ -n "$ES_USERNAME" ] || [ -n "$ES_PASSWORD" ]; then`,
		`  status=$(curl -sS -o /tmp/esrally-cleanup.out -w "%{http_code}" -u "${ES_USERNAME}:${ES_PASSWORD}" -X DELETE "${TARGET_URL}/${INDEX_NAME}" || true)`,
		`else`,
		`  status=$(curl -sS -o /tmp/esrally-cleanup.out -w "%{http_code}" -X DELETE "${TARGET_URL}/${INDEX_NAME}" || true)`,
		`fi`,
		`cat /tmp/esrally-cleanup.out | tee -a "` + esrallyLogFile + `"`,
		`case "$status" in`,
		`  200|202|404) echo "Cleanup finished with HTTP ${status}" | tee -a "` + esrallyLogFile + `" ;;`,
		`  *) echo "Failed to delete index ${INDEX_NAME}: HTTP ${status}" | tee -a "` + esrallyLogFile + `"; exit 1 ;;`,
		`esac`,
	}, "\n")
}

func esrallyPrepareScript() string {
	return strings.Join([]string{
		`set -eu`,
		esrallyGeneratedDataUnsupportedConfigCheck(),
		`set +e`,
		`python3 <<'PY' > /tmp/esrally-prepare.out 2>&1`,
		esrallyPreparePythonScript(),
		`PY`,
		`status=$?`,
		`set -e`,
		`cat /tmp/esrally-prepare.out | tee -a "` + esrallyLogFile + `"`,
		`exit "$status"`,
	}, "\n")
}

func esrallyGeneratedDataUnsupportedConfigCheck() string {
	return strings.Join([]string{
		`if [ -n "$CLIENT_OPTIONS" ] || [ -n "$TARGET_HOSTS_OVERRIDE" ]; then`,
		`  echo "generated ESRally data mode supports only spec.target.host, spec.target.port, spec.target.user, and spec.target.password for cleanup/prepare" | tee -a "` + esrallyLogFile + `"`,
		`  exit 1`,
		`fi`,
	}, "\n")
}

func esrallyPreparePythonScript() string {
	return `import base64
import datetime
import json
import os
import random
import sys
import urllib.error
import urllib.request

target_url = os.environ["TARGET_URL"].rstrip("/")
index_name = os.environ["INDEX_NAME"]
profile = os.environ["DATA_PROFILE"]
document_count = int(os.environ["DOCUMENT_COUNT"])
username = os.environ.get("ES_USERNAME", "")
password = os.environ.get("ES_PASSWORD", "")
batch_size = 500
random.seed(7)

def request(method, path, body=None):
    headers = {}
    data = None
    if body is not None:
        data = body.encode("utf-8")
        headers["Content-Type"] = "application/x-ndjson" if path.endswith("_bulk") else "application/json"
    req = urllib.request.Request(target_url + path, data=data, headers=headers, method=method)
    if username or password:
        token = base64.b64encode(f"{username}:{password}".encode("utf-8")).decode("ascii")
        req.add_header("Authorization", "Basic " + token)
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            return resp.status, resp.read().decode("utf-8")
    except urllib.error.HTTPError as err:
        return err.code, err.read().decode("utf-8")

def log_doc(i):
    status = random.choice([200, 200, 200, 201, 204, 301, 400, 404, 500])
    service = random.choice(["api", "checkout", "search", "billing"])
    path = random.choice(["/api/search", "/api/orders", "/api/cart", "/health", "/login"])
    return {
        "@timestamp": (datetime.datetime.utcnow() - datetime.timedelta(seconds=document_count - i)).isoformat() + "Z",
        "service": service,
        "host": f"host-{i % 12}",
        "method": random.choice(["GET", "POST", "PUT", "DELETE"]),
        "path": path,
        "status": status,
        "latency_ms": round(random.uniform(2, 900), 3),
        "bytes": random.randint(128, 65536),
        "message": f"{service} handled {path} with status {status}",
    }

def metrics_doc(i):
    return {
        "@timestamp": (datetime.datetime.utcnow() - datetime.timedelta(seconds=document_count - i)).isoformat() + "Z",
        "host": f"node-{i % 10}",
        "pod": f"pod-{i % 30}",
        "container": random.choice(["app", "sidecar", "worker"]),
        "namespace": random.choice(["default", "search", "payments"]),
        "cpu_pct": round(random.uniform(0, 100), 3),
        "memory_mb": round(random.uniform(128, 32768), 3),
        "disk_read_bytes": random.randint(0, 10485760),
        "disk_write_bytes": random.randint(0, 10485760),
        "network_rx_bytes": random.randint(0, 10485760),
        "network_tx_bytes": random.randint(0, 10485760),
    }

if profile == "logs":
    make_doc = log_doc
elif profile == "metrics":
    make_doc = metrics_doc
else:
    print(f"unsupported dataProfile {profile!r}; expected logs or metrics", file=sys.stderr)
    sys.exit(1)

print(f"Generating {document_count} {profile} documents into index {index_name}")
sent = 0
while sent < document_count:
    upper = min(sent + batch_size, document_count)
    lines = []
    for i in range(sent, upper):
        lines.append(json.dumps({"index": {"_index": index_name}}, separators=(",", ":")))
        lines.append(json.dumps(make_doc(i), separators=(",", ":")))
    status, body = request("POST", "/_bulk", "\n".join(lines) + "\n")
    if status >= 300:
        print(f"bulk request failed with HTTP {status}: {body[:1000]}", file=sys.stderr)
        sys.exit(1)
    result = json.loads(body)
    if result.get("errors"):
        first_error = next((item for item in result.get("items", []) if item.get("index", {}).get("error")), None)
        print(f"bulk request returned item errors: {first_error}", file=sys.stderr)
        sys.exit(1)
    sent = upper
    print(f"Indexed {sent}/{document_count} documents")

status, body = request("POST", f"/{index_name}/_refresh", "{}")
if status >= 300:
    print(f"refresh failed with HTTP {status}: {body[:1000]}", file=sys.stderr)
    sys.exit(1)
print("Generated ESRally dataset is ready")
`
}

func esrallyRunScript(cr *v1alpha1.Esrally) string {
	flags := []string{
		`set -eu`,
		`fail_before_rally() { echo "$1" | tee "` + esrallyLogFile + `"; echo "1" > "` + esrallyExitFile + `"; exit 1; }`,
		`if [ -z "$TRACK_PATH" ]; then fail_before_rally "ESRally generated-data mode requires spec.trackPath; remote Rally tracks and corpora are not supported"; fi`,
		`set -- race --pipeline=benchmark-only --target-hosts "$TARGET_HOSTS" --track-path "$TRACK_PATH" --offline --on-error "$ON_ERROR" --report-format "$REPORT_FORMAT" --report-file "$REPORT_FILE"`,
		`if [ -n "$CHALLENGE" ]; then set -- "$@" --challenge "$CHALLENGE"; fi`,
		`if [ -n "$INCLUDE_TASKS" ]; then set -- "$@" --include-tasks "$INCLUDE_TASKS"; fi`,
		`if [ -n "$TRACK_PARAMS" ]; then set -- "$@" --track-params "$TRACK_PARAMS"; fi`,
		`if [ -n "$CLIENT_OPTIONS" ]; then set -- "$@" --client-options "$CLIENT_OPTIONS"; fi`,
		`if [ -n "$TELEMETRY" ]; then set -- "$@" --telemetry "$TELEMETRY"; fi`,
		`if [ -n "$TELEMETRY_PARAMS" ]; then set -- "$@" --telemetry-params "$TELEMETRY_PARAMS"; fi`,
	}
	if cr.Spec.TestMode {
		flags = append(flags, `set -- "$@" --test-mode`)
	}
	flags = append(flags,
		`if [ -n "$EXTRA_ARGS" ]; then set -- "$@" $EXTRA_ARGS; fi`,
		`set +e`,
		`esrally "$@" > /tmp/esrally.out 2>&1`,
		`status=$?`,
		`cat /tmp/esrally.out | tee "`+esrallyLogFile+`"`,
		`if [ -f "$REPORT_FILE" ]; then echo "Rally CSV report:" | tee -a "`+esrallyLogFile+`"; cat "$REPORT_FILE" | tee -a "`+esrallyLogFile+`"; fi`,
		`echo "$status" > "`+esrallyExitFile+`"`,
		`exit "$status"`,
	)
	return strings.Join(flags, "\n")
}

func esrallyStep(cr *v1alpha1.Esrally) string {
	if cr.Spec.Step != "" {
		return cr.Spec.Step
	}
	return constants.AllStep
}

func esrallyDataProfile(cr *v1alpha1.Esrally) string {
	if cr.Spec.DataProfile != "" {
		return cr.Spec.DataProfile
	}
	return constants.EsrallyDataProfileLogs
}

func esrallyDocumentCount(cr *v1alpha1.Esrally) int {
	if cr.Spec.DocumentCount > 0 {
		return cr.Spec.DocumentCount
	}
	return esrallyDefaultDocs
}

func esrallyIndexName(cr *v1alpha1.Esrally) string {
	if cr.Spec.Target.Database != "" {
		return cr.Spec.Target.Database
	}
	return esrallyDefaultIndex
}

func esrallyTargetURL(cr *v1alpha1.Esrally) string {
	return fmt.Sprintf("http://%s:%d", cr.Spec.Target.Host, cr.Spec.Target.Port)
}

func esrallyTargetHosts(cr *v1alpha1.Esrally) string {
	if len(cr.Spec.TargetHosts) > 0 {
		return strings.Join(cr.Spec.TargetHosts, ",")
	}
	return fmt.Sprintf("%s:%d", cr.Spec.Target.Host, cr.Spec.Target.Port)
}

func esrallyTrackParams(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	data, err := json.Marshal(params)
	if err != nil {
		return ""
	}
	return string(data)
}

func esrallyClientOptions(cr *v1alpha1.Esrally) string {
	if cr.Spec.ClientOptions != "" {
		return cr.Spec.ClientOptions
	}
	if cr.Spec.Target.User == "" && cr.Spec.Target.Password == "" {
		return ""
	}
	return fmt.Sprintf("basic_auth_user:'%s',basic_auth_password:'%s'", cr.Spec.Target.User, cr.Spec.Target.Password)
}

func esrallyOnError(cr *v1alpha1.Esrally) string {
	if cr.Spec.OnError != "" {
		return cr.Spec.OnError
	}
	return esrallyDefaultOnError
}
