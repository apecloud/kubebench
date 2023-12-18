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

func NewTpccJobs(cr *v1alpha1.Tpcc) []*batchv1.Job {

	jobs := make([]*batchv1.Job, 0)

	step := cr.Spec.Step
	if step == "cleanup" || step == "all" {
		jobs = append(jobs, NewTpccCleanupJobs(cr)...)
	}
	if step == "prepare" || step == "all" {
		jobs = append(jobs, NewTpccPrepareJobs(cr)...)
	}
	if step == "run" || step == "all" {
		jobs = append(jobs, NewTpccRunJobs(cr)...)
	}

	// set tolerations for all jobs
	utils.AddTolerationToJobs(jobs, cr.Spec.Tolerations)

	// add cr labels to all jobs
	utils.AddLabelsToJobs(jobs, cr.Labels)
	utils.AddLabelsToJobs(jobs, map[string]string{
		constants.KubeBenchNameLabel: cr.Name,
		constants.KubeBenchTypeLabel: constants.TpccType,
	})

	// add cpu and memory to all jobs
	utils.AddCpuAndMemoryToJobs(jobs, cr.Spec.Cpu, cr.Spec.Memory)

	return jobs
}

func NewTpccCleanupJobs(cr *v1alpha1.Tpcc) []*batchv1.Job {
	cmd := "python3 main.py"
	cmd = fmt.Sprintf("%s --mode %s", cmd, "cleanup")
	cmd = fmt.Sprintf("%s --db %s", cmd, cr.Spec.Target.Driver)
	cmd = fmt.Sprintf("%s --user %s", cmd, cr.Spec.Target.User)
	cmd = fmt.Sprintf("%s --password %s", cmd, cr.Spec.Target.Password)
	cmd = fmt.Sprintf("%s %s", cmd, NewTpccWorkLoadParams(cr))
	cmd = fmt.Sprintf("%s %s", cmd, strings.Join(cr.Spec.ExtraArgs, " "))

	job := utils.JobTemplate(fmt.Sprintf("%s-cleanup", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchEnvTpcc),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c", cmd},
		},
	)

	return []*batchv1.Job{job}
}

func NewTpccPrepareJobs(cr *v1alpha1.Tpcc) []*batchv1.Job {
	cmd := "python3 main.py"
	cmd = fmt.Sprintf("%s --mode %s", cmd, "prepare")
	cmd = fmt.Sprintf("%s --db %s", cmd, cr.Spec.Target.Driver)
	cmd = fmt.Sprintf("%s --user %s", cmd, cr.Spec.Target.User)
	cmd = fmt.Sprintf("%s --password %s", cmd, cr.Spec.Target.Password)
	cmd = fmt.Sprintf("%s %s", cmd, NewTpccWorkLoadParams(cr))
	cmd = fmt.Sprintf("%s --warehouses %d", cmd, cr.Spec.WareHouses)
	cmd = fmt.Sprintf("%s %s", cmd, strings.Join(cr.Spec.ExtraArgs, " "))

	job := utils.JobTemplate(fmt.Sprintf("%s-prepare", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchEnvTpcc),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c", cmd},
		},
	)

	return []*batchv1.Job{job}
}

func NewTpccRunJobs(cr *v1alpha1.Tpcc) []*batchv1.Job {
	cmd := "python3 main.py"
	cmd = fmt.Sprintf("%s --mode %s", cmd, "run")
	cmd = fmt.Sprintf("%s --db %s", cmd, cr.Spec.Target.Driver)
	cmd = fmt.Sprintf("%s --user %s", cmd, cr.Spec.Target.User)
	cmd = fmt.Sprintf("%s --password %s", cmd, cr.Spec.Target.Password)
	cmd = fmt.Sprintf("%s %s", cmd, NewTpccWorkLoadParams(cr))
	cmd = fmt.Sprintf("%s --warehouses %d", cmd, cr.Spec.WareHouses)
	cmd = fmt.Sprintf("%s --limitTxnsPerMin %d", cmd, cr.Spec.LimitTxPerMin)

	if cr.Spec.Transactions != 0 {
		cmd = fmt.Sprintf("%s --runTxnsPerTerminal %d", cmd, cr.Spec.Transactions)
	} else {
		cmd = fmt.Sprintf("%s --runMins %d", cmd, cr.Spec.Duration)
	}

	cmd = fmt.Sprintf("%s --newOrderWeight %d", cmd, cr.Spec.NewOrder)
	cmd = fmt.Sprintf("%s --paymentWeight %d", cmd, cr.Spec.Payment)
	cmd = fmt.Sprintf("%s --orderStatusWeight %d", cmd, cr.Spec.OrderStatus)
	cmd = fmt.Sprintf("%s --deliveryWeight %d", cmd, cr.Spec.Delivery)
	cmd = fmt.Sprintf("%s --stockLevelWeight %d", cmd, cr.Spec.StockLevel)
	cmd = fmt.Sprintf("%s %s", cmd, strings.Join(cr.Spec.ExtraArgs, " "))

	jobs := make([]*batchv1.Job, 0)
	for i, thread := range cr.Spec.Threads {
		curCmd := fmt.Sprintf("%s --threads %d", cmd, thread)
		jobName := fmt.Sprintf("%s-run-%d", cr.Name, i)
		curJob := utils.JobTemplate(jobName, cr.Namespace)
		curJob.Spec.Template.Spec.Containers = append(
			curJob.Spec.Template.Spec.Containers,
			corev1.Container{
				Name:            constants.ContainerName,
				Image:           constants.GetBenchmarkImage(constants.KubebenchEnvTpcc),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"/bin/sh", "-c", curCmd},
			},
		)
		jobs = append(jobs, curJob)
	}

	return jobs
}

func NewTpccWorkLoadParams(cr *v1alpha1.Tpcc) string {
	switch cr.Spec.Target.Driver {
	case "mysql":
		return NewTpccMysqlParams(cr)
	case "postgres":
		return NewTpccPostgresParams(cr)
	default:
		return ""
	}
}

func NewTpccMysqlParams(cr *v1alpha1.Tpcc) string {
	result := fmt.Sprintf("--driver %s", "com.mysql.cj.jdbc.Driver")
	result = fmt.Sprintf("%s --conn \"jdbc:mysql://%s:%d/%s?useSSL=false&allowPublicKeyRetrieval=true\"", result, cr.Spec.Target.Host, cr.Spec.Target.Port, cr.Spec.Target.Database)
	return result
}

func NewTpccPostgresParams(cr *v1alpha1.Tpcc) string {
	result := fmt.Sprintf("--driver %s", "org.postgresql.Driver")
	result = fmt.Sprintf("%s --conn jdbc:postgresql://%s:%d/%s", result, cr.Spec.Target.Host, cr.Spec.Target.Port, cr.Spec.Target.Database)
	return result
}
