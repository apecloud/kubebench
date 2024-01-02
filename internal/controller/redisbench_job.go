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

func NewRedisBenchJobs(cr *v1alpha1.Redisbench) []*batchv1.Job {
	jobs := make([]*batchv1.Job, 0)

	step := cr.Spec.Step

	if step == "run" || step == "all" {
		jobs = append(jobs, NewRedisBenchRunJobs(cr)...)
	}

	return jobs
}

func NewRedisBenchRunJobs(cr *v1alpha1.Redisbench) []*batchv1.Job {
	cmd := "redis-benchmark"
	cmd = fmt.Sprintf("%s -h %s", cmd, cr.Spec.Target.Host)
	cmd = fmt.Sprintf("%s -p %d", cmd, cr.Spec.Target.Port)
	cmd = fmt.Sprintf("%s -a %s", cmd, cr.Spec.Target.Password)
	cmd = fmt.Sprintf("%s -n %d", cmd, cr.Spec.Requests)
	cmd = fmt.Sprintf("%s -d %d", cmd, cr.Spec.DataSize)
	cmd = fmt.Sprintf("%s -P %d", cmd, cr.Spec.Pipeline)
	cmd = fmt.Sprintf("%s %s", cmd, strings.Join(cr.Spec.ExtraArgs, " "))

	if cr.Spec.KeySpace != nil {
		cmd = fmt.Sprintf("%s -r %d", cmd, *cr.Spec.KeySpace)
	}
	if cr.Spec.Tests != "" {
		cmd = fmt.Sprintf("%s -t %s", cmd, cr.Spec.Tests)
	}
	if cr.Spec.Quiet {
		cmd = fmt.Sprintf("%s -q", cmd)
	}

	jobs := make([]*batchv1.Job, 0)
	for client := range cr.Spec.Clients {
		curCmd := fmt.Sprintf("%s -c %d", cmd, client)
		jobName := fmt.Sprintf("%s-%d", cr.Name, client)
		curJob := utils.JobTemplate(jobName, cr.Namespace)

		curJob.Spec.Template.Spec.Containers = append(
			curJob.Spec.Template.Spec.Containers,
			corev1.Container{
				Name:            constants.ContainerName,
				Image:           constants.GetBenchmarkImage(constants.KubebenchEnvRedisBench),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"/bin/sh", "-c", curCmd},
			},
		)

		jobs = append(jobs, curJob)
	}

	return jobs
}
