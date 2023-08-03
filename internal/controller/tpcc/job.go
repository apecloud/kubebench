package tpcc

import (
	"fmt"
	"strings"

	"github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/internal/utils"
	"github.com/apecloud/kubebench/pkg/constants"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func NewJobs(cr *v1alpha1.Tpcc) []*batchv1.Job {

	jobs := make([]*batchv1.Job, 0)

	mode := cr.Spec.Mode
	if mode == "cleanup" || mode == "all" {
		jobs = append(jobs, NewCleanupJobs(cr)...)
	}
	if mode == "prepare" || mode == "all" {
		jobs = append(jobs, NewPrepareJobs(cr)...)
	}
	if mode == "run" || mode == "all" {
		jobs = append(jobs, NewRunJobs(cr)...)
	}

	return jobs
}

func NewCleanupJobs(cr *v1alpha1.Tpcc) []*batchv1.Job {
	cmd := "python3 main.py"
	cmd = fmt.Sprintf("%s --mode %s", cmd, "cleanup")
	cmd = fmt.Sprintf("%s --db %s", cmd, cr.Spec.Target.Driver)
	cmd = fmt.Sprintf("%s --user %s", cmd, cr.Spec.Target.User)
	cmd = fmt.Sprintf("%s --password %s", cmd, cr.Spec.Target.Password)
	cmd = fmt.Sprintf("%s %s", cmd, NewWorkLoadParams(cr))
	cmd = fmt.Sprintf("%s %s", cmd, strings.Join(cr.Spec.ExtraArgs, " "))

	job := utils.JobTemplate(fmt.Sprintf("%s-cleanup", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.TpccImage,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c", cmd},
		},
	)

	return []*batchv1.Job{job}
}

func NewPrepareJobs(cr *v1alpha1.Tpcc) []*batchv1.Job {
	cmd := "python3 main.py"
	cmd = fmt.Sprintf("%s --mode %s", cmd, "prepare")
	cmd = fmt.Sprintf("%s --db %s", cmd, cr.Spec.Target.Driver)
	cmd = fmt.Sprintf("%s --user %s", cmd, cr.Spec.Target.User)
	cmd = fmt.Sprintf("%s --password %s", cmd, cr.Spec.Target.Password)
	cmd = fmt.Sprintf("%s %s", cmd, NewWorkLoadParams(cr))
	cmd = fmt.Sprintf("%s --warehouses %d", cmd, cr.Spec.WareHouses)
	cmd = fmt.Sprintf("%s %s", cmd, strings.Join(cr.Spec.ExtraArgs, " "))

	job := utils.JobTemplate(fmt.Sprintf("%s-prepare", cr.Name), cr.Namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.TpccImage,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/bin/sh", "-c", cmd},
		},
	)

	return []*batchv1.Job{job}
}

func NewRunJobs(cr *v1alpha1.Tpcc) []*batchv1.Job {
	cmd := "python3 main.py"
	cmd = fmt.Sprintf("%s --mode %s", cmd, "run")
	cmd = fmt.Sprintf("%s --db %s", cmd, cr.Spec.Target.Driver)
	cmd = fmt.Sprintf("%s --user %s", cmd, cr.Spec.Target.User)
	cmd = fmt.Sprintf("%s --password %s", cmd, cr.Spec.Target.Password)
	cmd = fmt.Sprintf("%s %s", cmd, NewWorkLoadParams(cr))
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
				Image:           constants.TpccImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"/bin/sh", "-c", curCmd},
			},
		)
		jobs = append(jobs, curJob)
	}

	return jobs
}

func NewWorkLoadParams(cr *v1alpha1.Tpcc) string {
	switch cr.Spec.Target.Driver {
	case "mysql":
		return NewMysqlParams(cr)
	case "postgres":
		return NewPostgresParams(cr)
	default:
		return ""
	}
}

func NewMysqlParams(cr *v1alpha1.Tpcc) string {
	result := fmt.Sprintf("--driver %s", "com.mysql.cj.jdbc.Driver")
	result = fmt.Sprintf("%s --conn jdbc:mysql://%s:%d/%s?useSSL=false", result, cr.Spec.Target.Host, cr.Spec.Target.Port, cr.Spec.Target.Database)
	return result
}

func NewPostgresParams(cr *v1alpha1.Tpcc) string {
	result := fmt.Sprintf("--driver %s", "org.postgresql.Driver")
	result = fmt.Sprintf("%s --conn jdbc:postgresql://%s:%d/%s", result, cr.Spec.Target.Host, cr.Spec.Target.Port, cr.Spec.Target.Database)
	return result
}
