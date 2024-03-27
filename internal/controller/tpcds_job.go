package controller

import (
	"fmt"
	"github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/internal/utils"
	"github.com/apecloud/kubebench/pkg/constants"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func NewTpcdsJobs(cr v1alpha1.Tpcds) []*batchv1.Job {
	jobs := make([]*batchv1.Job, 0)

	// add pre-check job
	if utils.NewPreCheckJob(cr.Name, cr.Namespace, cr.Spec.Target.Driver, &cr.Spec.Target) != nil {
		jobs = append(jobs, utils.NewPreCheckJob(cr.Name, cr.Namespace, cr.Spec.Target.Driver, &cr.Spec.Target))
	}

	step := cr.Spec.Step
	if step == constants.CleanupStep || step == constants.AllStep {
		jobs = append(jobs, NewTpcdsCleanupJobs(cr)...)
	}
	if step == constants.PrepareStep || step == constants.AllStep {
		jobs = append(jobs, NewTpcdsPrepareJobs(cr)...)
	}
	if step == constants.RunStep || step == constants.AllStep {
		jobs = append(jobs, NewTpcdsRunJobs(cr)...)
	}

	// set tolerations for all jobs
	utils.AddTolerationToJobs(jobs, cr.Spec.Tolerations)

	// add cr labels to all jobs
	utils.AddLabelsToJobs(jobs, cr.Labels)
	utils.AddLabelsToJobs(jobs, map[string]string{
		constants.KubeBenchNameLabel: cr.Name,
		constants.KubeBenchTypeLabel: constants.TpcdsType,
	})

	// add resource requirements for all jobs
	utils.AddResourceLimitsToJobs(jobs, cr.Spec.ResourceLimits)
	utils.AddResourceRequestsToJobs(jobs, cr.Spec.ResourceRequests)

	return jobs
}

func NewTpcdsCleanupJobs(cr v1alpha1.Tpcds) []*batchv1.Job {
	cmd := "python3 -u main.py"
	cmd = fmt.Sprintf("%s --driver %s", cmd, cr.Spec.Target.Driver)
	cmd = fmt.Sprintf("%s --host %s", cmd, cr.Spec.Target.Host)
	cmd = fmt.Sprintf("%s --port %d", cmd, cr.Spec.Target.Port)
	cmd = fmt.Sprintf("%s --user %s", cmd, cr.Spec.Target.User)
	cmd = fmt.Sprintf("%s --password %s", cmd, cr.Spec.Target.Password)
	cmd = fmt.Sprintf("%s --database %s", cmd, cr.Spec.Target.Database)
	cmd = fmt.Sprintf("%s --step %s", cmd, "cleanup")

	job := utils.JobTemplate(fmt.Sprintf("%s-cleanup", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchEnvTpcds),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c", cmd},
		},
	)

	// add init container to create database for cleanup
	if initContainer := TpcdsInitContainers(&cr); initContainer != nil {
		job.Spec.Template.Spec.InitContainers = append(job.Spec.Template.Spec.InitContainers, *initContainer)
	}

	return []*batchv1.Job{job}
}

func NewTpcdsPrepareJobs(cr v1alpha1.Tpcds) []*batchv1.Job {
	cmd := "python3 -u main.py"
	cmd = fmt.Sprintf("%s --scale %d", cmd, cr.Spec.Size)
	cmd = fmt.Sprintf("%s --driver %s", cmd, cr.Spec.Target.Driver)
	cmd = fmt.Sprintf("%s --host %s", cmd, cr.Spec.Target.Host)
	cmd = fmt.Sprintf("%s --port %d", cmd, cr.Spec.Target.Port)
	cmd = fmt.Sprintf("%s --user %s", cmd, cr.Spec.Target.User)
	cmd = fmt.Sprintf("%s --password %s", cmd, cr.Spec.Target.Password)
	cmd = fmt.Sprintf("%s --database %s", cmd, cr.Spec.Target.Database)
	cmd = fmt.Sprintf("%s --step %s", cmd, "prepare")
	if cr.Spec.UseKey {
		cmd = fmt.Sprintf("%s --use-key", cmd)
	}

	job := utils.JobTemplate(fmt.Sprintf("%s-prepare", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchEnvTpcds),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c", cmd},
		},
	)

	// add init container to create database for prepare job
	if initContainer := TpcdsInitContainers(&cr); initContainer != nil {
		job.Spec.Template.Spec.InitContainers = append(job.Spec.Template.Spec.InitContainers, *initContainer)
	}

	return []*batchv1.Job{job}
}

func NewTpcdsRunJobs(cr v1alpha1.Tpcds) []*batchv1.Job {
	cmd := "python3 -u main.py"
	cmd = fmt.Sprintf("%s --scale %d", cmd, cr.Spec.Size)
	cmd = fmt.Sprintf("%s --driver %s", cmd, cr.Spec.Target.Driver)
	cmd = fmt.Sprintf("%s --host %s", cmd, cr.Spec.Target.Host)
	cmd = fmt.Sprintf("%s --port %d", cmd, cr.Spec.Target.Port)
	cmd = fmt.Sprintf("%s --user %s", cmd, cr.Spec.Target.User)
	cmd = fmt.Sprintf("%s --password %s", cmd, cr.Spec.Target.Password)
	cmd = fmt.Sprintf("%s --database %s", cmd, cr.Spec.Target.Database)
	cmd = fmt.Sprintf("%s --step %s", cmd, "run")

	job := utils.JobTemplate(fmt.Sprintf("%s-run", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchEnvTpcds),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c", cmd},
		},
	)

	return []*batchv1.Job{job}
}

func TpcdsInitContainers(cr *v1alpha1.Tpcds) *corev1.Container {
	database := cr.Spec.Target.Database

	switch cr.Spec.Target.Driver {
	case constants.MySqlDriver:
		return utils.InitMysqlDatabaseContainer(cr.Spec.Target, database)
	case constants.PostgreSqlDriver:
		return utils.InitPGDatabaseContainer(cr.Spec.Target, database)
	default:
		return nil
	}
}
