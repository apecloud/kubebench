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

	// add pre-check job
	if utils.NewPreCheckJob(cr.Name, cr.Namespace, cr.Spec.Target.Driver, &cr.Spec.Target) != nil {
		jobs = append(jobs, utils.NewPreCheckJob(cr.Name, cr.Namespace, cr.Spec.Target.Driver, &cr.Spec.Target))
	}

	step := cr.Spec.Step
	if step == constants.CleanupStep || step == constants.AllStep {
		jobs = append(jobs, NewYcsbCleanupJobs(cr)...)
	}
	if step == constants.PrepareStep || step == constants.AllStep {
		jobs = append(jobs, NewYcsbPrepareJobs(cr)...)
	}
	if step == constants.RunStep || step == constants.AllStep {
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

	// add resource requirements for all jobs
	utils.AddResourceLimitsToJobs(jobs, cr.Spec.ResourceLimits)
	utils.AddResourceRequestsToJobs(jobs, cr.Spec.ResourceRequests)

	return jobs
}

func NewYcsbCleanupJobs(cr *v1alpha1.Ycsb) []*batchv1.Job {
	var container *corev1.Container
	job := utils.JobTemplate(fmt.Sprintf("%s-cleanup", cr.Name), cr.Namespace)

	switch cr.Spec.Target.Driver {
	case constants.MySqlDriver:
		container = utils.CleanMysqlDatabaseContainer(cr.Spec.Target, cr.Spec.Target.Database)
	case constants.PostgreSqlDriver:
		container = utils.CleanPGDatabaseContainer(cr.Spec.Target, cr.Spec.Target.Database)
	case constants.MongoDbDriver:
		// 'ycsb' is the default database name when running ycsb on mongodb
		container = utils.CleanMongoDatabaseContainer(cr.Spec.Target, "ycsb")
	}

	if container == nil {
		return nil
	}

	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		*container,
	)

	return []*batchv1.Job{job}
}

func NewYcsbPrepareJobs(cr *v1alpha1.Ycsb) []*batchv1.Job {
	cmd := "/go-ycsb"
	cmd = fmt.Sprintf("%s load %s --interval 1", cmd, getYcsbDriver(cr.Spec.Target.Driver))
	cmd = fmt.Sprintf("%s %s", cmd, NewYcsbWorkloadParams(cr))
	cmd = fmt.Sprintf("%s -p recordcount=%d", cmd, cr.Spec.RecordCount)
	cmd = fmt.Sprintf("%s -p operationcount=%d", cmd, cr.Spec.OperationCount)
	cmd = fmt.Sprintf("%s -p threadcount=%d", cmd, cr.Spec.Threads[0])
	cmd = fmt.Sprintf("%s -p requestdistribution=%s", cmd, cr.Spec.RequestDistribution)
	cmd = fmt.Sprintf("%s -p scanlengthdistribution=%s", cmd, cr.Spec.ScanLengthDistribution)
	cmd = fmt.Sprintf("%s -p fieldlengthdistribution=%s", cmd, cr.Spec.FieldLengthDistribution)
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

	// add init containers to create database for prepare job
	// only mysql and postgresql need init containers
	initContainer := YcsbInitContainers(cr)
	if initContainer != nil {
		job.Spec.Template.Spec.InitContainers = append(
			job.Spec.Template.Spec.InitContainers,
			*initContainer,
		)
	}

	return []*batchv1.Job{job}
}

func NewYcsbRunJobs(cr *v1alpha1.Ycsb) []*batchv1.Job {
	cmd := "/go-ycsb"
	cmd = fmt.Sprintf("%s run %s --interval 1", cmd, getYcsbDriver(cr.Spec.Target.Driver))
	cmd = fmt.Sprintf("%s %s", cmd, NewYcsbWorkloadParams(cr))
	cmd = fmt.Sprintf("%s -p recordcount=%d", cmd, cr.Spec.RecordCount)
	cmd = fmt.Sprintf("%s -p operationcount=%d", cmd, cr.Spec.OperationCount)
	cmd = fmt.Sprintf("%s -p requestdistribution=%s", cmd, cr.Spec.RequestDistribution)
	cmd = fmt.Sprintf("%s -p scanlengthdistribution=%s", cmd, cr.Spec.ScanLengthDistribution)
	cmd = fmt.Sprintf("%s -p fieldlengthdistribution=%s", cmd, cr.Spec.FieldLengthDistribution)
	cmd = fmt.Sprintf("%s %s", cmd, strings.Join(cr.Spec.ExtraArgs, " "))

	totalProportion := cr.Spec.ReadProportion + cr.Spec.UpdateProportion + cr.Spec.InsertProportion + cr.Spec.ReadModifyWriteProportion + cr.Spec.ScanProportion
	readProportion, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(cr.Spec.ReadProportion)/float64(totalProportion)), 64)
	cmd = fmt.Sprintf("%s -p readproportion=%f", cmd, readProportion)

	updateProportion, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(cr.Spec.UpdateProportion)/float64(totalProportion)), 64)
	cmd = fmt.Sprintf("%s -p updateproportion=%f", cmd, updateProportion)

	insertProportion, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(cr.Spec.InsertProportion)/float64(totalProportion)), 64)
	cmd = fmt.Sprintf("%s -p insertproportion=%f", cmd, insertProportion)

	readModifyWriteProportion, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(cr.Spec.ReadModifyWriteProportion)/float64(totalProportion)), 64)
	cmd = fmt.Sprintf("%s -p readmodifywriteproportion=%f", cmd, readModifyWriteProportion)
	scanProportion, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(cr.Spec.ScanProportion)/float64(totalProportion)), 64)
	cmd = fmt.Sprintf("%s -p scanproportion=%f", cmd, scanProportion)

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
	case constants.MySqlDriver:
		return NewYcsbMysqlParams(cr)
	case constants.RedisDriver:
		return NewYcsbRedisParams(cr)
	case constants.PostgreSqlDriver:
		return NewYcsbPostgresParams(cr)
	case constants.MongoDbDriver:
		return NewYcsbMongodbParams(cr)
	case constants.MinioDriver:
		return NewYcsbMinioParams(cr)
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
	result := ""
	if cr.Spec.RedisAddr != "" {
		result = fmt.Sprintf("-p redis.addr='%s'", cr.Spec.RedisAddr)
	} else {
		result = fmt.Sprintf("-p redis.addr='%s'", fmt.Sprintf("%s:%d", cr.Spec.Target.Host, cr.Spec.Target.Port))
	}
	if cr.Spec.Target.User != "" {
		result = fmt.Sprintf("%s -p redis.username=%s", result, cr.Spec.Target.User)
	}
	if cr.Spec.Target.Password != "" {
		result = fmt.Sprintf("%s -p redis.password=%s", result, cr.Spec.Target.Password)
	}
	if cr.Spec.Target.Database != "" {
		result = fmt.Sprintf("%s -p redis.db=%s", result, cr.Spec.Target.Database)
	}
	if cr.Spec.RedisMode == "sentinel" {
		result = fmt.Sprintf("%s -p redis.mode=sentinel", result)
		result = fmt.Sprintf("%s -p redis.sentinel_master_name=%s", result, cr.Spec.MasterName)
		result = fmt.Sprintf("%s -p redis.sentinel_username=%s", result, cr.Spec.RedisSentinelUsername)
		result = fmt.Sprintf("%s -p redis.sentinel_password=%s", result, cr.Spec.RedisSentinelPassword)
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

func NewYcsbMinioParams(cr *v1alpha1.Ycsb) string {
	accessKey := cr.Spec.Target.User
	secretKey := cr.Spec.Target.Password
	endpoint := fmt.Sprintf("%s:%d", cr.Spec.Target.Host, cr.Spec.Target.Port)
	return fmt.Sprintf("-p minio.access-key=%s -p minio.secret-key=%s -p minio.endpoint=%s -p table=%s", accessKey, secretKey, endpoint, cr.Spec.Target.Database)
}

func YcsbInitContainers(cr *v1alpha1.Ycsb) *corev1.Container {
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

func getYcsbDriver(driver string) string {
	switch driver {
	case constants.MySqlDriver:
		return "mysql"
	case constants.PostgreSqlDriver:
		return "postgresql"
	case constants.MongoDbDriver:
		return "mongodb"
	case constants.RedisDriver:
		return "redis"
	case constants.MinioDriver:
		return "minio"
	default:
		return driver
	}
}
