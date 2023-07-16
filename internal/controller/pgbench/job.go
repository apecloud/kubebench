package pgbench

import (
	"fmt"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/pkg/constants"
)

func NewJob(cr *v1alpha1.Pgbench, jobName string) *batchv1.Job {
	backoffLimit := int32(0) // no retry

	cmd := "pgbench"
	if cr.Status.Ready {
		cmd = fmt.Sprintf("%s -c %d", cmd, cr.Spec.Clients[cr.Status.Succeeded])

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
	} else {
		// TODO add func to parse extra args
		cmd = fmt.Sprintf("%s -i -s%d %s", cmd, cr.Spec.Scale, strings.Join(cr.Spec.ExtraArgs, " "))
	}

	cmd = fmt.Sprintf("%s 2>&1 | tee /var/log/pgbench.log", cmd)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: cr.Namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/managed-by": "kubebench",
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "log",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	if cr.Status.Ready {
		job.Spec.Template.Spec.InitContainers = append(
			job.Spec.Template.Spec.InitContainers,
			corev1.Container{
				Name:            constants.ContainerName,
				Image:           constants.PgbenchImage,
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
			})

		job.Spec.Template.Spec.Containers = append(
			job.Spec.Template.Spec.Containers,
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
				Args:    []string{"-type", "pgbench", "-file", "/var/log/pgbench.log"},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "log",
						MountPath: "/var/log",
					},
				},
			})
	} else {
		job.Spec.Template.Spec.Containers = append(
			job.Spec.Template.Spec.Containers,
			corev1.Container{
				Name:            constants.ContainerName,
				Image:           constants.PgbenchImage,
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
	}

	return job
}
