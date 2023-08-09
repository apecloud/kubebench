package controller

import (
	"fmt"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/internal/utils"
	"github.com/apecloud/kubebench/pkg/constants"
)

func NewPgbenchJobs(cr *v1alpha1.Pgbench) []*batchv1.Job {
	jobs := make([]*batchv1.Job, 0)

	step := cr.Spec.Step
	if step == "cleanup" || step == "all" {
		jobs = append(jobs, NewPgbenchCleanupJobs(cr)...)
	}
	if step == "prepare" || step == "all" {
		jobs = append(jobs, NewPgbenchPrepareJobs(cr)...)
	}
	if step == "run" || step == "all" {
		jobs = append(jobs, NewPgbenchRunJobs(cr)...)
	}

	// set tolerations for all jobs
	for _, job := range jobs {
		job.Spec.Template.Spec.Tolerations = cr.Spec.Tolerations
	}

	return jobs
}

func NewPgbenchCleanupJobs(cr *v1alpha1.Pgbench) []*batchv1.Job {
	cmd := "pgbench"
	cmd = fmt.Sprintf("%s -i -I d", cmd)

	job := utils.JobTemplate(fmt.Sprintf("%s-cleanup", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetPgbenchImage(),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c", cmd},
			Env: []corev1.EnvVar{
				{
					Name:  "PGHOST",
					Value: cr.Spec.Target.Host,
				},
				{
					Name:  "PGPORT",
					Value: fmt.Sprintf("%d", cr.Spec.Target.Port),
				},
				{
					Name:  "PGUSER",
					Value: cr.Spec.Target.User,
				},
				{
					Name:  "PGPASSWORD",
					Value: cr.Spec.Target.Password,
				},
				{
					Name:  "PGDATABASE",
					Value: cr.Spec.Target.Database,
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "log",
					MountPath: "/var/log",
				},
			},
		},
	)

	return []*batchv1.Job{job}
}

func NewPgbenchPrepareJobs(cr *v1alpha1.Pgbench) []*batchv1.Job {
	cmd := "pgbench"
	cmd = fmt.Sprintf("%s -i -s%d %s", cmd, cr.Spec.Scale, strings.Join(cr.Spec.ExtraArgs, " "))

	job := utils.JobTemplate(fmt.Sprintf("%s-prepare", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetPgbenchImage(),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c", cmd},
			Env: []corev1.EnvVar{
				{
					Name:  "PGHOST",
					Value: cr.Spec.Target.Host,
				},
				{
					Name:  "PGPORT",
					Value: fmt.Sprintf("%d", cr.Spec.Target.Port),
				},
				{
					Name:  "PGUSER",
					Value: cr.Spec.Target.User,
				},
				{
					Name:  "PGPASSWORD",
					Value: cr.Spec.Target.Password,
				},
				{
					Name:  "PGDATABASE",
					Value: cr.Spec.Target.Database,
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "log",
					MountPath: "/var/log",
				},
			},
		},
	)

	return []*batchv1.Job{job}
}

func NewPgbenchRunJobs(cr *v1alpha1.Pgbench) []*batchv1.Job {
	cmd := "pgbench"
	cmd = fmt.Sprintf("%s -P 1", cmd)
	cmd = fmt.Sprintf("%s -j %d", cmd, cr.Spec.Threads)

	// priority: transactions > time
	switch {
	case cr.Spec.Transactions > 0:
		cmd = fmt.Sprintf("%s -t %d", cmd, cr.Spec.Transactions)
	case cr.Spec.Duration > 0:
		cmd = fmt.Sprintf("%s -T %d", cmd, cr.Spec.Duration)
	}

	if cr.Spec.Connect {
		cmd = fmt.Sprintf("%s -C", cmd)
	}

	if cr.Spec.SelectOnly {
		cmd = fmt.Sprintf("%s -S", cmd)
	}

	// TODO add func to parse extra args
	cmd = fmt.Sprintf("%s %s", cmd, strings.Join(cr.Spec.ExtraArgs, " "))

	jobs := make([]*batchv1.Job, 0)
	for i, client := range cr.Spec.Clients {
		curCmd := fmt.Sprintf("%s -c %d", cmd, client)
		jobName := fmt.Sprintf("%s-run-%d", cr.Name, i)
		curJob := utils.JobTemplate(jobName, cr.Namespace)

		curJob.Spec.Template.Spec.Containers = append(
			curJob.Spec.Template.Spec.Containers,
			corev1.Container{
				Name:            constants.ContainerName,
				Image:           constants.GetPgbenchImage(),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"/bin/sh", "-c", curCmd},
				Env: []corev1.EnvVar{
					{
						Name:  "PGHOST",
						Value: cr.Spec.Target.Host,
					},
					{
						Name:  "PGPORT",
						Value: fmt.Sprintf("%d", cr.Spec.Target.Port),
					},
					{
						Name:  "PGUSER",
						Value: cr.Spec.Target.User,
					},
					{
						Name:  "PGPASSWORD",
						Value: cr.Spec.Target.Password,
					},
					{
						Name:  "PGDATABASE",
						Value: cr.Spec.Target.Database,
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "log",
						MountPath: "/var/log",
					},
				},
			})

		curJob.Spec.Template.Spec.Containers = append(
			curJob.Spec.Template.Spec.Containers,
			corev1.Container{
				Name:            "metrics",
				Image:           constants.PrometheusExporterImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 9187,
						Name:          "http-metrics",
						Protocol:      corev1.ProtocolTCP,
					},
				},
				Command: []string{"/exporter"},
				Args:    []string{"-type", "pgbench", "-file", "/var/log/pgbench.log", "-bench", cr.Name, "-job", jobName},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "log",
						MountPath: "/var/log",
					},
				},
			})

		jobs = append(jobs, curJob)
	}

	return jobs
}
