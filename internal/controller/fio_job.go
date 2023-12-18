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

func NewFioJobs(cr *v1alpha1.Fio) []*batchv1.Job {
	jobs := NewFioRunJobs(cr)

	// set tolerations for all jobs
	utils.AddTolerationToJobs(jobs, cr.Spec.Tolerations)

	// add cr labels to all jobs
	utils.AddLabelsToJobs(jobs, cr.Labels)
	utils.AddLabelsToJobs(jobs, map[string]string{
		constants.KubeBenchNameLabel: cr.Name,
		constants.KubeBenchTypeLabel: constants.FioType,
	})

	// add cpu and memory to all jobs
	utils.AddCpuAndMemoryToJobs(jobs, cr.Spec.Cpu, cr.Spec.Memory)

	return jobs
}

func NewFioRunJobs(cr *v1alpha1.Fio) []*batchv1.Job {
	cmd := "fio -group_reporting"
	cmd = fmt.Sprintf("%s -size %s", cmd, cr.Spec.Size)
	cmd = fmt.Sprintf("%s -bs %s", cmd, cr.Spec.Bs)
	cmd = fmt.Sprintf("%s -iodepth %d", cmd, cr.Spec.Iodepth)
	cmd = fmt.Sprintf("%s -ioengine %s", cmd, cr.Spec.IoEngine)
	cmd = fmt.Sprintf("%s %s", cmd, strings.Join(cr.Spec.ExtraArgs, " "))
	if cr.Spec.RunTime > 0 {
		cmd = fmt.Sprintf("%s -runtime %d", cmd, cr.Spec.RunTime)
	}
	if cr.Spec.Direct {
		cmd = fmt.Sprintf("%s -direct 1", cmd)
	}

	jobs := make([]*batchv1.Job, 0)
	for i := 0; i < len(cr.Spec.Numjobs)*len(cr.Spec.Rws); i++ {
		curCmd := fmt.Sprintf("%s -numjobs %d", cmd, cr.Spec.Numjobs[i/len(cr.Spec.Rws)])
		curCmd = fmt.Sprintf("%s -rw %s", curCmd, cr.Spec.Rws[i%len(cr.Spec.Rws)])
		curCmd = fmt.Sprintf("%s -name %s-%d", curCmd, cr.Name, i)
		jobName := fmt.Sprintf("%s-%d", cr.Name, i)
		curJob := utils.JobTemplate(jobName, cr.Namespace)

		curJob.Spec.Template.Spec.Containers = append(
			curJob.Spec.Template.Spec.Containers,
			corev1.Container{
				Name:            constants.ContainerName,
				Image:           constants.GetBenchmarkImage(constants.KubebenchEnvFio),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"/bin/sh", "-c", curCmd},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "log",
						MountPath: "/data",
					},
				},
			},
		)

		jobs = append(jobs, curJob)
	}

	return jobs
}
