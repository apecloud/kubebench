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

	// add pre-check job
	if utils.NewPreCheckJob(cr.Name, cr.Namespace, cr.Spec.Target.Driver, &cr.Spec.Target) != nil {
		jobs = append(jobs, utils.NewPreCheckJob(cr.Name, cr.Namespace, cr.Spec.Target.Driver, &cr.Spec.Target))
	}

	step := cr.Spec.Step
	if step == constants.CleanupStep || step == constants.AllStep {
		jobs = append(jobs, NewTpccCleanupJobs(cr)...)
	}
	if step == constants.PrepareStep || step == constants.AllStep {
		jobs = append(jobs, NewTpccPrepareJobs(cr)...)
	}
	if step == constants.RunStep || step == constants.AllStep {
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

	// add resource requirements for all jobs
	utils.AddResourceLimitsToJobs(jobs, cr.Spec.ResourceLimits)
	utils.AddResourceRequestsToJobs(jobs, cr.Spec.ResourceRequests)

	return jobs
}

func NewTpccCleanupJobs(cr *v1alpha1.Tpcc) []*batchv1.Job {
	cmd := "python3 main.py"
	cmd = fmt.Sprintf("%s --mode %s", cmd, "cleanup")
	cmd = fmt.Sprintf("%s --db %s", cmd, getTpccDriver(cr.Spec.Target.Driver))
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
	cmd = fmt.Sprintf("%s --db %s", cmd, getTpccDriver(cr.Spec.Target.Driver))
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

	// add init containers to create database for prepare job
	if initContainer := TpccInitContainers(cr); initContainer != nil {
		job.Spec.Template.Spec.InitContainers = append(job.Spec.Template.Spec.InitContainers, *initContainer)
	}

	return []*batchv1.Job{job}
}

func NewTpccRunJobs(cr *v1alpha1.Tpcc) []*batchv1.Job {
	cmd := "python3 main.py"
	cmd = fmt.Sprintf("%s --mode %s", cmd, "run")
	cmd = fmt.Sprintf("%s --db %s", cmd, getTpccDriver(cr.Spec.Target.Driver))
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
	case constants.MySqlDriver:
		return NewTpccMysqlParams(cr)
	case constants.PostgreSqlDriver:
		return NewTpccPostgresParams(cr)
	case constants.OceanBaseOracleTenantDriver:
		return NewOceanBaseOracleTenantParams(cr)
	case constants.DamengDriver:
		return NewDamengParams(cr)
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

func NewOceanBaseOracleTenantParams(cr *v1alpha1.Tpcc) string {
	result := fmt.Sprintf("--driver %s", "com.alipay.oceanbase.obproxy.mysql.jdbc.Driver")
	result = fmt.Sprintf("%s --conn \"jdbc:oceanbase://%s:%d/%s?useUnicode=true&characterEncoding=utf-8\"", result, cr.Spec.Target.Host, cr.Spec.Target.Port, cr.Spec.Target.Database)
	return result
}

func NewDamengParams(cr *v1alpha1.Tpcc) string {
	result := fmt.Sprintf("--driver %s", "dm.jdbc.driver.DmDriver")
	result = fmt.Sprintf("%s --conn jdbc:dm://%s:%d", result, cr.Spec.Target.Host, cr.Spec.Target.Port)
	return result
}

// TpccInitContainers returns the init containers for tpcc
// tpcc will fail if database not exists, so we need to create database first
func TpccInitContainers(cr *v1alpha1.Tpcc) *corev1.Container {
	switch cr.Spec.Target.Driver {
	case constants.MySqlDriver:
		return utils.InitMysqlDatabaseContainer(cr.Spec.Target, cr.Spec.Target.Database)
	case constants.PostgreSqlDriver:
		return utils.InitPGDatabaseContainer(cr.Spec.Target, cr.Spec.Target.Database)
	default:
		return nil
	}
}

// getTpccDriver returns the database type required by tpcc
func getTpccDriver(driver string) string {
	switch driver {
	case constants.MySqlDriver:
		return "mysql"
	case constants.PostgreSqlDriver:
		return "postgres"
	case constants.OceanBaseOracleTenantDriver:
		return "oracle"
	default:
		return driver
	}
}
