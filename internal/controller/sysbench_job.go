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

func NewSysbenchJobs(cr *v1alpha1.Sysbench) []*batchv1.Job {
	jobs := make([]*batchv1.Job, 0)

	// add pre-check job
	jobs = append(jobs, utils.NewPreCheckJob(cr.Name, cr.Namespace, cr.Spec.Target.Driver, &cr.Spec.Target))

	step := cr.Spec.Step
	if step == constants.CleanupStep || step == constants.AllStep {
		jobs = append(jobs, NewSysbenchCleanupJobs(cr)...)
	}
	if step == constants.PrepareStep || step == constants.AllStep {
		jobs = append(jobs, NewSysbenchPrepareJobs(cr)...)
	}
	if step == constants.RunStep || step == constants.AllStep {
		jobs = append(jobs, NewSysbenchRunJobs(cr)...)
	}

	// set tolerations for all jobs
	utils.AddTolerationToJobs(jobs, cr.Spec.Tolerations)

	// add cr labels to all jobs
	utils.AddLabelsToJobs(jobs, cr.Labels)
	utils.AddLabelsToJobs(jobs, map[string]string{
		constants.KubeBenchNameLabel: cr.Name,
		constants.KubeBenchTypeLabel: constants.SysbenchType,
	})

	// add resource requirements for all jobs
	utils.AddResourceLimitsToJobs(jobs, cr.Spec.ResourceLimits)
	utils.AddResourceRequestsToJobs(jobs, cr.Spec.ResourceRequests)

	return jobs
}

func NewSysbenchCleanupJobs(cr *v1alpha1.Sysbench) []*batchv1.Job {
	value := fmt.Sprintf("mode:%s", "cleanup")
	value = fmt.Sprintf("%s,driver:%s", value, getSysbenchDriver(cr.Spec.Target.Driver))
	value = fmt.Sprintf("%s,host:%s", value, cr.Spec.Target.Host)
	value = fmt.Sprintf("%s,port:%d", value, cr.Spec.Target.Port)
	value = fmt.Sprintf("%s,user:%s", value, cr.Spec.Target.User)
	value = fmt.Sprintf("%s,password:%s", value, cr.Spec.Target.Password)
	value = fmt.Sprintf("%s,db:%s", value, cr.Spec.Target.Database)
	value = fmt.Sprintf("%s,tables:%d", value, cr.Spec.Tables)
	value = fmt.Sprintf("%s,size:%d", value, cr.Spec.Size)
	value = fmt.Sprintf("%s,times:%d", value, cr.Spec.Duration)
	value = fmt.Sprintf("%s,threads:%d", value, cr.Spec.Threads[0])
	value = fmt.Sprintf("%s,type:%s", value, cr.Spec.Types[0])

	// TODO add func to parse extra args
	value = fmt.Sprintf("%s,others:%s", value, strings.Join(cr.Spec.ExtraArgs, " "))

	job := utils.JobTemplate(fmt.Sprintf("%s-cleanup", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchEnvSysbench),
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
	)

	return []*batchv1.Job{job}
}

func NewSysbenchPrepareJobs(cr *v1alpha1.Sysbench) []*batchv1.Job {
	value := fmt.Sprintf("mode:%s", "prepare")
	value = fmt.Sprintf("%s,driver:%s", value, getSysbenchDriver(cr.Spec.Target.Driver))
	value = fmt.Sprintf("%s,host:%s", value, cr.Spec.Target.Host)
	value = fmt.Sprintf("%s,port:%d", value, cr.Spec.Target.Port)
	value = fmt.Sprintf("%s,user:%s", value, cr.Spec.Target.User)
	value = fmt.Sprintf("%s,password:%s", value, cr.Spec.Target.Password)
	value = fmt.Sprintf("%s,db:%s", value, cr.Spec.Target.Database)
	value = fmt.Sprintf("%s,tables:%d", value, cr.Spec.Tables)
	value = fmt.Sprintf("%s,size:%d", value, cr.Spec.Size)
	value = fmt.Sprintf("%s,times:%d", value, cr.Spec.Duration)
	value = fmt.Sprintf("%s,threads:%d", value, cr.Spec.Threads[0])
	value = fmt.Sprintf("%s,type:%s", value, cr.Spec.Types[0])

	// TODO add func to parse extra args
	value = fmt.Sprintf("%s,others:%s", value, strings.Join(cr.Spec.ExtraArgs, " "))

	job := utils.JobTemplate(fmt.Sprintf("%s-prepare", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchEnvSysbench),
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
	)

	return []*batchv1.Job{job}
}

func NewSysbenchRunJobs(cr *v1alpha1.Sysbench) []*batchv1.Job {
	value := fmt.Sprintf("mode:%s", "run")
	value = fmt.Sprintf("%s,driver:%s", value, getSysbenchDriver(cr.Spec.Target.Driver))
	value = fmt.Sprintf("%s,host:%s", value, cr.Spec.Target.Host)
	value = fmt.Sprintf("%s,port:%d", value, cr.Spec.Target.Port)
	value = fmt.Sprintf("%s,user:%s", value, cr.Spec.Target.User)
	value = fmt.Sprintf("%s,password:%s", value, cr.Spec.Target.Password)
	value = fmt.Sprintf("%s,db:%s", value, cr.Spec.Target.Database)
	value = fmt.Sprintf("%s,tables:%d", value, cr.Spec.Tables)
	value = fmt.Sprintf("%s,size:%d", value, cr.Spec.Size)
	value = fmt.Sprintf("%s,times:%d", value, cr.Spec.Duration)

	// TODO add func to parse extra args
	value = fmt.Sprintf("%s,others:%s", value, strings.Join(cr.Spec.ExtraArgs, " "))

	jobs := make([]*batchv1.Job, 0)
	for i := 0; i < len(cr.Spec.Threads)*len(cr.Spec.Types); i++ {
		curValue := fmt.Sprintf("%s,threads:%d", value, cr.Spec.Threads[i/len(cr.Spec.Types)])
		curValue = fmt.Sprintf("%s,type:%s", curValue, cr.Spec.Types[i%len(cr.Spec.Types)])
		jobName := fmt.Sprintf("%s-run-%d", cr.Name, i)
		curJob := utils.JobTemplate(jobName, cr.Namespace)

		curJob.Spec.Template.Spec.Containers = append(
			curJob.Spec.Template.Spec.Containers,
			corev1.Container{
				Name:            constants.ContainerName,
				Image:           constants.GetBenchmarkImage(constants.KubebenchEnvSysbench),
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
						Value: curValue,
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

		curJob.Spec.Template.Spec.Containers = append(
			curJob.Spec.Template.Spec.Containers,
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
				Args:    []string{"-type", "sysbench", "-file", "/var/log/sysbench.log", "-bench", cr.Name, "-job", jobName},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "log",
						MountPath: "/var/log",
					},
				},
			},
		)

		jobs = append(jobs, curJob)
	}

	return jobs
}

// getSysbenchDriver returns the database type required by sysbench
func getSysbenchDriver(driver string) string {
	switch driver {
	case constants.MySqlDriver:
		return "mysql"
	case constants.PostgreSqlDriver:
		return "pgsql"
	default:
		return driver
	}
}
