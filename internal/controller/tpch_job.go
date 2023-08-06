package controller

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/internal/utils"
	"github.com/apecloud/kubebench/pkg/constants"
)

func NewTpchJobs(cr *v1alpha1.Tpch) []*batchv1.Job {
	jobs := make([]*batchv1.Job, 0)

	step := cr.Spec.Step

	if step == "all" {
		return NewTpchAllJobs(cr)
	}
	if step == "run" {
		return NewTpchRunJobs(cr)
	}

	return jobs
}

func NewTpchCleanupJobs(cr *v1alpha1.Tpch) []*batchv1.Job {
	// TODO: implement this
	return nil
}

func NewTpchPrepareJobs(cr *v1alpha1.Tpch) []*batchv1.Job {
	return nil
}

func NewTpchRunJobs(cr *v1alpha1.Tpch) []*batchv1.Job {
	value := fmt.Sprintf("host:%s", cr.Spec.Target.Host)
	value = fmt.Sprintf("%s,port:%d", value, cr.Spec.Target.Port)
	value = fmt.Sprintf("%s,user:%s", value, cr.Spec.Target.User)
	value = fmt.Sprintf("%s,password:%s", value, cr.Spec.Target.Password)
	value = fmt.Sprintf("%s,db:%s", value, cr.Spec.Target.Database)
	value = fmt.Sprintf("%s,local:%s", value, "True")

	job := utils.JobTemplate(fmt.Sprintf("%s-run", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.TpchImage,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c"},
			Args:            []string{"python3 -u infratest.py -t \"$TYPE\" -f \"${FLAG}\" -c \"${CONFIGS}\" -j \"${JSONS}\" | tee /var/log/sysbench.log"},
			Env: []corev1.EnvVar{
				{
					Name:  "TYPE",
					Value: "3",
				},
				{
					Name:  "FLAG",
					Value: "1",
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
	)

	return []*batchv1.Job{job}
}

func NewTpchAllJobs(cr *v1alpha1.Tpch) []*batchv1.Job {
	value := fmt.Sprintf("host:%s", cr.Spec.Target.Host)
	value = fmt.Sprintf("%s,port:%d", value, cr.Spec.Target.Port)
	value = fmt.Sprintf("%s,user:%s", value, cr.Spec.Target.User)
	value = fmt.Sprintf("%s,password:%s", value, cr.Spec.Target.Password)
	value = fmt.Sprintf("%s,db:%s", value, cr.Spec.Target.Database)
	value = fmt.Sprintf("%s,local:%s", value, "True")

	job := utils.JobTemplate(fmt.Sprintf("%s-all", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.TpchImage,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c"},
			Args:            []string{"python3 -u infratest.py -t \"$TYPE\" -f \"${FLAG}\" -c \"${CONFIGS}\" -j \"${JSONS}\" | tee /var/log/sysbench.log"},
			Env: []corev1.EnvVar{
				{
					Name:  "TYPE",
					Value: "3",
				},
				{
					Name:  "FLAG",
					Value: "5",
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
	)

	return []*batchv1.Job{job}
}
