package controller

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/internal/utils"
	"github.com/apecloud/kubebench/pkg/constants"
)

const (
	esrallyLogFile            = "/var/log/esrally.log"
	esrallyExitFile           = "/var/log/esrally.exit"
	esrallyHomeMountPath      = "/rally/.rally"
	esrallyGeneratedTrackPath = "/tmp/kubebench-esrally-track"
	esrallyDocumentsFile      = esrallyGeneratedTrackPath + "/documents.json"
	esrallyDefaultOnError     = "abort"
	esrallyReportFormat       = "csv"
	esrallyReportFile         = "/var/log/esrally-report.csv"
	esrallyDefaultIndex       = "kubebench"
	esrallyDefaultDocs        = 10000
	esrallyTargetIndexParam   = "target_index"
	esrallyTargetVersionParam = "target_version"
	esrallyScriptDir          = "/usr/local/share/kubebench/esrally"
	esrallyCleanupScriptPath  = esrallyScriptDir + "/cleanup.py"
	esrallyPrepareScriptPath  = esrallyScriptDir + "/prepare.py"
	esrallyGenerateScriptPath = esrallyScriptDir + "/generate_track.py"
	esrallyRunScriptPath      = esrallyScriptDir + "/run.sh"
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
			Command:         []string{"python3"},
			Args:            []string{esrallyCleanupScriptPath},
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
			Command:         []string{"python3"},
			Args:            []string{esrallyPrepareScriptPath},
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
		{Name: "TARGET_VERSION", Value: esrallyTargetVersion(cr)},
		{Name: "INDEX_NAME", Value: esrallyIndexName(cr)},
		{Name: "DATA_PROFILE", Value: esrallyDataProfile(cr)},
		{Name: "DOCUMENT_COUNT", Value: fmt.Sprintf("%d", esrallyDocumentCount(cr))},
		{Name: "WORKLOAD", Value: esrallyWorkload(cr)},
		{Name: "GENERATED_TRACK_PATH", Value: esrallyGeneratedTrackPath},
		{Name: "DOCUMENTS_FILE", Value: esrallyDocumentsFile},
		{Name: "TRACK_PARAMS", Value: esrallyTrackParams(cr)},
		{Name: "CLIENT_OPTIONS", Value: esrallyClientOptions(cr)},
		{Name: "ON_ERROR", Value: esrallyOnError(cr)},
		{Name: "TELEMETRY", Value: strings.Join(esrallyTelemetry(cr), ",")},
		{Name: "REPORT_FORMAT", Value: esrallyReportFormat},
		{Name: "REPORT_FILE", Value: esrallyReportFile},
		{Name: "ESRALLY_LOG_FILE", Value: esrallyLogFile},
		{Name: "ESRALLY_EXIT_FILE", Value: esrallyExitFile},
		{Name: "GENERATE_TRACK_SCRIPT", Value: esrallyGenerateScriptPath},
		{Name: "EXTRA_ARGS", Value: strings.Join(cr.Spec.ExtraArgs, " ")},
	}

	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchEnvEsrally),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c"},
			Args:            []string{fmt.Sprintf("/bin/sh %s", esrallyRunScriptPath)},
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
		{Name: "WORKLOAD", Value: esrallyWorkload(cr)},
		{Name: "TARGET_VERSION", Value: esrallyTargetVersion(cr)},
		{Name: "ES_USERNAME", Value: cr.Spec.Target.User},
		{Name: "ES_PASSWORD", Value: cr.Spec.Target.Password},
		{Name: "ES_INSECURE_SKIP_VERIFY", Value: strconv.FormatBool(cr.Spec.Target.TLS)},
		{Name: "ESRALLY_LOG_FILE", Value: esrallyLogFile},
	}
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

func esrallyWorkload(cr *v1alpha1.Esrally) string {
	if cr.Spec.Workload != "" {
		return cr.Spec.Workload
	}
	return constants.EsrallyWorkloadAll
}

func esrallyIndexName(cr *v1alpha1.Esrally) string {
	if cr.Spec.Target.Database != "" {
		return cr.Spec.Target.Database
	}
	return esrallyDefaultIndex
}

func esrallyTargetURL(cr *v1alpha1.Esrally) string {
	return fmt.Sprintf("%s://%s:%d", esrallyTargetScheme(cr), cr.Spec.Target.Host, cr.Spec.Target.Port)
}

func esrallyTargetHosts(cr *v1alpha1.Esrally) string {
	return fmt.Sprintf("%s:%d", cr.Spec.Target.Host, cr.Spec.Target.Port)
}

func esrallyTargetVersion(cr *v1alpha1.Esrally) string {
	return strings.TrimSpace(cr.Spec.TargetVersion)
}

func esrallyTargetScheme(cr *v1alpha1.Esrally) string {
	if cr.Spec.Target.TLS {
		return "https"
	}
	return "http"
}

func esrallyTrackParams(cr *v1alpha1.Esrally) string {
	params := map[string]string{
		esrallyTargetIndexParam: esrallyIndexName(cr),
	}
	targetVersion := esrallyTargetVersion(cr)
	if targetVersion != "" {
		params[esrallyTargetVersionParam] = targetVersion
	}
	data, err := json.Marshal(params)
	if err != nil {
		return ""
	}
	return string(data)
}

func esrallyTelemetry(cr *v1alpha1.Esrally) []string {
	telemetry := make([]string, 0, len(cr.Spec.Telemetry))
	for _, device := range cr.Spec.Telemetry {
		telemetry = append(telemetry, string(device))
	}
	return telemetry
}

func esrallyClientOptions(cr *v1alpha1.Esrally) string {
	options := make([]string, 0, 4)
	if cr.Spec.Target.TLS {
		options = append(options, "use_ssl:true", "verify_certs:false")
	}
	if cr.Spec.Target.User != "" && cr.Spec.Target.Password != "" {
		options = append(options,
			fmt.Sprintf("basic_auth_user:'%s'", cr.Spec.Target.User),
			fmt.Sprintf("basic_auth_password:'%s'", cr.Spec.Target.Password),
		)
	}
	if len(options) == 0 {
		return ""
	}
	return strings.Join(options, ",")
}

func esrallyOnError(cr *v1alpha1.Esrally) string {
	if cr.Spec.OnError != "" {
		return cr.Spec.OnError
	}
	return esrallyDefaultOnError
}
