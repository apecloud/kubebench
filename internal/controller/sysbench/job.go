package sysbench

import (
	"fmt"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/pkg/constants"
)

func NewJob(cr *v1alpha1.Sysbench, jobName string) *batchv1.Job {
	backoffLimit := int32(0) // no retry

	value := fmt.Sprintf("mode:%s", "all")
	value = fmt.Sprintf("%s,driver:%s", value, cr.Spec.Target.Driver)
	value = fmt.Sprintf("%s,host:%s", value, cr.Spec.Target.Host)
	value = fmt.Sprintf("%s,port:%d", value, cr.Spec.Target.Port)
	value = fmt.Sprintf("%s,user:%s", value, cr.Spec.Target.User)
	value = fmt.Sprintf("%s,password:%s", value, cr.Spec.Target.Password)
	value = fmt.Sprintf("%s,db:%s", value, cr.Spec.Target.Database)
	value = fmt.Sprintf("%s,tables:%d", value, cr.Spec.Tables)
	value = fmt.Sprintf("%s,size:%d", value, cr.Spec.Size)
	value = fmt.Sprintf("%s,times:%d", value, cr.Spec.Duration)
	value = fmt.Sprintf("%s,threads:%d", value, cr.Spec.Threads[cr.Status.Succeeded/len(cr.Spec.Threads)])
	value = fmt.Sprintf("%s,type:%s", value, cr.Spec.Types[cr.Status.Succeeded%len(cr.Spec.Types)])

	// TODO add func to parse extra args
	value = fmt.Sprintf("%s,others:%s", value, strings.Join(cr.Spec.ExtraArgs, " "))

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
					InitContainers: []corev1.Container{
						{
							Name:            constants.ContainerName,
							Image:           constants.SysbenchImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"/bin/sh", "-c"},
							Args:            []string{"python3 -u infratest.py -t \"$TYPE\" -f \"${FLAG}\" -c \"${CONFIGS}\" -j \"${JSONS}\" | tee /var/log/sysbench.log"},
							Env: []corev1.EnvVar{
								{
									Name:  "TYPE",
									Value: "2",
								},
								{
									Name:  "FLAG",
									Value: "0",
								},
								{
									Name:  "CONFIGS",
									Value: value,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "log",
									MountPath: "/var/log",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
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
							Args:    []string{"-type", "sysbench", "-file", "/var/log/sysbench.log", "-bench", cr.Name, "-job", jobName},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "log",
									MountPath: "/var/log",
								},
							},
						},
					},
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

	return job
}
