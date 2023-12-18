package controller

import (
	"fmt"
	"strconv"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/internal/utils"
	"github.com/apecloud/kubebench/pkg/constants"
)

func NewYcsbJobs(cr *v1alpha1.Ycsb) []*batchv1.Job {
	jobs := make([]*batchv1.Job, 0)

	step := cr.Spec.Step
	// TODO: add cleanup
	if step == "prepare" || step == "all" {
		jobs = append(jobs, NewYcsbPrepareJobs(cr)...)
	}
	if step == "run" || step == "all" {
		jobs = append(jobs, NewYcsbRunJobs(cr)...)
	}

	// set tolerations for all jobs
	utils.AddTolerationToJobs(jobs, cr.Spec.Tolerations)

	// add cr labels to all jobs
	utils.AddLabelsToJobs(jobs, cr.Labels)
	utils.AddLabelsToJobs(jobs, map[string]string{
		constants.KubeBenchNameLabel: cr.Name,
		constants.KubeBenchTypeLabel: constants.YcsbType,
	})

	// add cpu and memory to all jobs
	utils.AddCpuAndMemoryToJobs(jobs, cr.Spec.Cpu, cr.Spec.Memory)

	return jobs
}

// TODO
func NewYscbCleanupJobs(cr *v1alpha1.Ycsb) []*batchv1.Job {
	return nil
}

func NewYcsbPrepareJobs(cr *v1alpha1.Ycsb) []*batchv1.Job {
	cmd := "/go-ycsb"
	cmd = fmt.Sprintf("%s load %s --interval 1", cmd, cr.Spec.Target.Driver)
	cmd = fmt.Sprintf("%s %s", cmd, NewYcsbWorkloadParams(cr))
	cmd = fmt.Sprintf("%s -p recordcount=%d", cmd, cr.Spec.RecordCount)
	cmd = fmt.Sprintf("%s -p operationcount=%d", cmd, cr.Spec.OperationCount)
	cmd = fmt.Sprintf("%s -p threadcount=%d", cmd, cr.Spec.Threads[0])
	cmd = fmt.Sprintf("%s %s", cmd, strings.Join(cr.Spec.ExtraArgs, " "))
	cmd = fmt.Sprintf("%s 2>&1 | tee /var/log/ycsb.log", cmd)

	job := utils.JobTemplate(fmt.Sprintf("%s-prepare", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchEnvYcsb),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c", cmd},
		},
	)

	return []*batchv1.Job{job}
}

func NewYcsbRunJobs(cr *v1alpha1.Ycsb) []*batchv1.Job {
	cmd := "/go-ycsb"
	cmd = fmt.Sprintf("%s run %s --interval 1", cmd, cr.Spec.Target.Driver)
	cmd = fmt.Sprintf("%s %s", cmd, NewYcsbWorkloadParams(cr))
	cmd = fmt.Sprintf("%s -p recordcount=%d", cmd, cr.Spec.RecordCount)
	cmd = fmt.Sprintf("%s -p operationcount=%d", cmd, cr.Spec.OperationCount)
	cmd = fmt.Sprintf("%s %s", cmd, strings.Join(cr.Spec.ExtraArgs, " "))

	totalProportion := cr.Spec.ReadProportion + cr.Spec.UpdateProportion + cr.Spec.InsertProportion + cr.Spec.ReadModifyWriteProportion + cr.Spec.ScanProportion
	if cr.Spec.ReadProportion > 0 {
		readProportion, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(cr.Spec.ReadProportion)/float64(totalProportion)), 64)
		cmd = fmt.Sprintf("%s -p readproportion=%f", cmd, readProportion)
	}
	if cr.Spec.UpdateProportion > 0 {
		updateProportion, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(cr.Spec.UpdateProportion)/float64(totalProportion)), 64)
		cmd = fmt.Sprintf("%s -p updateproportion=%f", cmd, updateProportion)
	}
	if cr.Spec.InsertProportion > 0 {
		insertProportion, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(cr.Spec.InsertProportion)/float64(totalProportion)), 64)
		cmd = fmt.Sprintf("%s -p insertproportion=%f", cmd, insertProportion)
	}
	if cr.Spec.ReadModifyWriteProportion > 0 {
		readModifyWriteProportion, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(cr.Spec.ReadModifyWriteProportion)/float64(totalProportion)), 64)
		cmd = fmt.Sprintf("%s -p readmodifywriteproportion=%f", cmd, readModifyWriteProportion)
	}
	if cr.Spec.ScanProportion > 0 {
		scanProportion, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(cr.Spec.ScanProportion)/float64(totalProportion)), 64)
		cmd = fmt.Sprintf("%s -p scanproportion=%f", cmd, scanProportion)
	}

	jobs := make([]*batchv1.Job, 0)
	for i, thread := range cr.Spec.Threads {
		jobName := fmt.Sprintf("%s-run-%d", cr.Name, i)
		curJob := utils.JobTemplate(jobName, cr.Namespace)
		curCmd := fmt.Sprintf("%s -p threadcount=%d", cmd, thread)
		curJob.Spec.Template.Spec.Containers = append(
			curJob.Spec.Template.Spec.Containers,
			corev1.Container{
				Name:            constants.ContainerName,
				Image:           constants.GetBenchmarkImage(constants.KubebenchEnvYcsb),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"/bin/sh", "-c", curCmd},
			},
		)
		jobs = append(jobs, curJob)
	}

	return jobs
}

func NewYcsbWorkloadParams(cr *v1alpha1.Ycsb) string {
	switch cr.Spec.Target.Driver {
	case "mysql":
		return NewYcsbMysqlParams(cr)
	case "redis":
		return NewYcsbRedisParams(cr)
	case "postgresql":
		return NewYcsbPostgresParams(cr)
	case "mongodb":
		return NewYcsbMongodbParams(cr)
	default:
		return ""
	}
}

func NewYcsbMysqlParams(cr *v1alpha1.Ycsb) string {
	result := fmt.Sprintf("-p mysql.host=%s", cr.Spec.Target.Host)
	result = fmt.Sprintf("%s -p mysql.port=%d", result, cr.Spec.Target.Port)
	if cr.Spec.Target.User != "" {
		result = fmt.Sprintf("%s -p mysql.user=%s", result, cr.Spec.Target.User)
	}
	if cr.Spec.Target.Password != "" {
		result = fmt.Sprintf("%s -p mysql.password=%s", result, cr.Spec.Target.Password)
	}
	if cr.Spec.Target.Database != "" {
		result = fmt.Sprintf("%s -p mysql.db=%s", result, cr.Spec.Target.Database)
	}

	return result
}

func NewYcsbRedisParams(cr *v1alpha1.Ycsb) string {
	result := fmt.Sprintf("-p redis.addr=%s", fmt.Sprintf("%s:%d", cr.Spec.Target.Host, cr.Spec.Target.Port))
	if cr.Spec.Target.User != "" {
		result = fmt.Sprintf("%s -p redis.username=%s", result, cr.Spec.Target.User)
	}
	if cr.Spec.Target.Password != "" {
		result = fmt.Sprintf("%s -p redis.password=%s", result, cr.Spec.Target.Password)
	}
	if cr.Spec.Target.Database != "" {
		result = fmt.Sprintf("%s -p redis.db=%s", result, cr.Spec.Target.Database)
	}
	return result
}

func NewYcsbPostgresParams(cr *v1alpha1.Ycsb) string {
	result := fmt.Sprintf("-p pg.host=%s", cr.Spec.Target.Host)
	result = fmt.Sprintf("%s -p pg.port=%d", result, cr.Spec.Target.Port)
	if cr.Spec.Target.User != "" {
		result = fmt.Sprintf("%s -p pg.user=%s", result, cr.Spec.Target.User)
	}
	if cr.Spec.Target.Password != "" {
		result = fmt.Sprintf("%s -p pg.password=%s", result, cr.Spec.Target.Password)
	}
	if cr.Spec.Target.Database != "" {
		result = fmt.Sprintf("%s -p pg.db=%s", result, cr.Spec.Target.Database)
	}
	return result
}

func NewYcsbMongodbParams(cr *v1alpha1.Ycsb) string {
	// TODO: parse extra args make user can set mongodb.uri
	mongdbUri := "mongodb://%s:%s@%s:%d/admin"
	result := fmt.Sprintf("-p mongodb.url=%s", fmt.Sprintf(mongdbUri, cr.Spec.Target.User, cr.Spec.Target.Password, cr.Spec.Target.Host, cr.Spec.Target.Port))
	return result
}
