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
	defaultEsrallyTrack        = "geonames"
	defaultEsrallyOnError      = "abort"
	defaultEsrallyReportFormat = "csv"
	defaultEsrallyReportFile   = "/var/log/esrally-report.csv"
	esrallyLogFile             = "/var/log/esrally.log"
	esrallyExitFile            = "/var/log/esrally.exit"
	esrallyHomeMountPath       = "/rally/.rally"
)

func NewEsrallyJobs(cr *v1alpha1.Esrally) []*batchv1.Job {
	jobs := make([]*batchv1.Job, 0)

	if job := utils.NewPreCheckJob(cr.Name, cr.Namespace, constants.ElasticsearchDriver, &cr.Spec.Target); job != nil {
		jobs = append(jobs, job)
	}

	jobs = append(jobs, NewEsrallyRunJobs(cr)...)

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

func NewEsrallyRunJobs(cr *v1alpha1.Esrally) []*batchv1.Job {
	jobName := fmt.Sprintf("%s-run", cr.Name)
	job := utils.JobTemplate(jobName, cr.Namespace)
	addEsrallyHomeVolume(job, cr.Spec.RallyHomePVCClaimName)

	env := []corev1.EnvVar{
		{Name: "TARGET_HOSTS", Value: esrallyTargetHosts(cr)},
		{Name: "TRACK", Value: esrallyTrack(cr)},
		{Name: "TRACK_REPOSITORY", Value: cr.Spec.TrackRepository},
		{Name: "TRACK_PATH", Value: cr.Spec.TrackPath},
		{Name: "CHALLENGE", Value: cr.Spec.Challenge},
		{Name: "INCLUDE_TASKS", Value: strings.Join(cr.Spec.IncludeTasks, ",")},
		{Name: "TRACK_PARAMS", Value: esrallyTrackParams(cr.Spec.TrackParams)},
		{Name: "CLIENT_OPTIONS", Value: esrallyClientOptions(cr)},
		{Name: "ON_ERROR", Value: esrallyOnError(cr)},
		{Name: "TELEMETRY", Value: strings.Join(cr.Spec.Telemetry, ",")},
		{Name: "TELEMETRY_PARAMS", Value: cr.Spec.TelemetryParams},
		{Name: "REPORT_FORMAT", Value: esrallyReportFormat(cr)},
		{Name: "REPORT_FILE", Value: esrallyReportFile(cr)},
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

	if esrallyMetricsEnabled(cr) {
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
					"-file", esrallyReportFile(cr),
					"-bench", cr.Name,
					"-job", jobName,
					"-done-file", esrallyExitFile,
				},
				VolumeMounts: []corev1.VolumeMount{
					{Name: "log", MountPath: "/var/log"},
				},
			},
		)
	}

	return []*batchv1.Job{job}
}

func addEsrallyHomeVolume(job *batchv1.Job, claimName string) {
	volume := corev1.Volume{Name: "rally-home"}
	if claimName != "" {
		volume.VolumeSource.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{
			ClaimName: claimName,
		}
	} else {
		volume.VolumeSource.EmptyDir = &corev1.EmptyDirVolumeSource{}
	}
	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, volume)
}

func esrallyRunScript(cr *v1alpha1.Esrally) string {
	flags := []string{
		`set -eu`,
		`set -- race --pipeline=benchmark-only --target-hosts "$TARGET_HOSTS" --on-error "$ON_ERROR" --report-format "$REPORT_FORMAT" --report-file "$REPORT_FILE"`,
		`if [ -n "$TRACK_PATH" ]; then set -- "$@" --track-path "$TRACK_PATH"; else set -- "$@" --track "$TRACK"; fi`,
		`if [ -n "$TRACK_REPOSITORY" ]; then set -- "$@" --track-repository "$TRACK_REPOSITORY"; fi`,
		`if [ -n "$CHALLENGE" ]; then set -- "$@" --challenge "$CHALLENGE"; fi`,
		`if [ -n "$INCLUDE_TASKS" ]; then set -- "$@" --include-tasks "$INCLUDE_TASKS"; fi`,
		`if [ -n "$TRACK_PARAMS" ]; then set -- "$@" --track-params "$TRACK_PARAMS"; fi`,
		`if [ -n "$CLIENT_OPTIONS" ]; then set -- "$@" --client-options "$CLIENT_OPTIONS"; fi`,
		`if [ -n "$TELEMETRY" ]; then set -- "$@" --telemetry "$TELEMETRY"; fi`,
		`if [ -n "$TELEMETRY_PARAMS" ]; then set -- "$@" --telemetry-params "$TELEMETRY_PARAMS"; fi`,
	}
	if cr.Spec.Offline {
		flags = append(flags, `set -- "$@" --offline`)
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

func esrallyTrack(cr *v1alpha1.Esrally) string {
	if cr.Spec.Track != "" {
		return cr.Spec.Track
	}
	return defaultEsrallyTrack
}

func esrallyOnError(cr *v1alpha1.Esrally) string {
	if cr.Spec.OnError != "" {
		return cr.Spec.OnError
	}
	return defaultEsrallyOnError
}

func esrallyReportFormat(cr *v1alpha1.Esrally) string {
	if cr.Spec.ReportFormat != "" {
		return cr.Spec.ReportFormat
	}
	return defaultEsrallyReportFormat
}

func esrallyReportFile(cr *v1alpha1.Esrally) string {
	if cr.Spec.ReportFile != "" {
		return cr.Spec.ReportFile
	}
	return defaultEsrallyReportFile
}

func esrallyMetricsEnabled(cr *v1alpha1.Esrally) bool {
	return cr.Spec.Metrics || cr.Spec.Metrics == false && cr.Spec.ReportFormat == ""
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
