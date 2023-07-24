package ycsb

import (
	"fmt"
	"strconv"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/pkg/constants"
)

func NewJob(cr *v1alpha1.Ycsb, jobName string) *batchv1.Job {
	backoffLimit := int32(0) // no retry

	cmd := "/go-ycsb"
	if cr.Status.Ready {
		cmd = fmt.Sprintf("%s run %s", cmd, cr.Spec.Target.Driver)
	} else {
		cmd = fmt.Sprintf("%s load %s -p dropdata=true", cmd, cr.Spec.Target.Driver)
	}

	totalProportion := cr.Spec.ReadProportion + cr.Spec.UpdateProportion
	readProportion, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(cr.Spec.ReadProportion)/float64(totalProportion)), 64)
	updateProportion, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(cr.Spec.UpdateProportion)/float64(totalProportion)), 64)
	cmd = fmt.Sprintf("%s %s", cmd, NewWorkloadParams(cr))
	cmd = fmt.Sprintf("%s -p recordcount=%d", cmd, cr.Spec.RecordCount)
	cmd = fmt.Sprintf("%s -p operationcount=%d", cmd, cr.Spec.OperationCount)
	cmd = fmt.Sprintf("%s -p readproportion=%f", cmd, readProportion)
	cmd = fmt.Sprintf("%s -p updateproportion=%f", cmd, updateProportion)
	cmd = fmt.Sprintf("%s -p threadcount=%d", cmd, cr.Spec.Threads[cr.Status.Succeeded])
	cmd = fmt.Sprintf("%s %s", cmd, strings.Join(cr.Spec.ExtraArgs, " "))

	cmd = fmt.Sprintf("%s 2>&1 | tee /var/log/ycsb.log", cmd)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: cr.Namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/managed-by": "kubebench",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            constants.ContainerName,
							Image:           constants.YcsbImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"/bin/sh", "-c", cmd},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	return job
}

func NewWorkloadParams(cr *v1alpha1.Ycsb) string {
	switch cr.Spec.Target.Driver {
	case "mysql":
		return NewMysqlParams(cr)
	case "redis":
		return NewRedisParams(cr)
	case "postgresql":
		return NewPostgresParams(cr)
	case "mongodb":
		return NewMongodbParams(cr)
	default:
		return ""
	}
}

func NewMysqlParams(cr *v1alpha1.Ycsb) string {
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

func NewRedisParams(cr *v1alpha1.Ycsb) string {
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

func NewPostgresParams(cr *v1alpha1.Ycsb) string {
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

func NewMongodbParams(cr *v1alpha1.Ycsb) string {
	mongdbUri := "mongodb://%s:%s@%s:%d/admin?replicaset=test-mongo-mongodb"
	result := fmt.Sprintf("-p mongodb.url=%s", fmt.Sprintf(mongdbUri, cr.Spec.Target.User, cr.Spec.Target.Password, cr.Spec.Target.Host, cr.Spec.Target.Port))
	return result
}
