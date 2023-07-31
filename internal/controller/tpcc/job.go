package tpcc

import (
	"fmt"

	"github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/pkg/constants"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewJob(cr *v1alpha1.Tpcc, jobName string) *batchv1.Job {
	backoffLimit := int32(0) // no retry

	cmd := "python3 main.py"
	cmd = fmt.Sprintf("%s --db %s", cmd, cr.Spec.Target.Driver)
	cmd = fmt.Sprintf("%s --user %s", cmd, cr.Spec.Target.User)
	cmd = fmt.Sprintf("%s --password %s", cmd, cr.Spec.Target.Password)
	cmd = fmt.Sprintf("%s %s", cmd, NewWorkLoadParams(cr))
	cmd = fmt.Sprintf("%s --warehouses %d", cmd, cr.Spec.WareHouses)
	cmd = fmt.Sprintf("%s --threads %d", cmd, cr.Spec.Threads[cr.Status.Succeeded])
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
							Image:           constants.TpccImage,
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
	result := fmt.Sprintf("--driver %s", "com.mysql.jdbc.Driver")
	result = fmt.Sprintf("%s --conn jdbc:mysql://%s:%d/%s?useSSL=false", result, cr.Spec.Target.Host, cr.Spec.Target.Port, cr.Spec.Target.Database)
	return result
}

func NewPostgresParams(cr *v1alpha1.Tpcc) string {
	result := fmt.Sprintf("--driver %s", "org.postgresql.Driver")
	result = fmt.Sprintf("%s --conn jdbc:postgresql://%s:%d/%s", result, cr.Spec.Target.Host, cr.Spec.Target.Port, cr.Spec.Target.Database)
	return result
}
