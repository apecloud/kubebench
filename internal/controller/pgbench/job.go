package pgbench

import (
	"fmt"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/internal/utils"
)

const (
	PgbenchName  = "pgbench"
	PgbenchImage = "postgres:latest"
)

func NewJob(cr *v1alpha1.Pgbench, jobName string) *batchv1.Job {
	var cmds []string
	if cr.Status.Ready {
		cmds = []string{"pgbench", "-c", fmt.Sprintf("%d", cr.Spec.Clients[cr.Status.Succeeded])}

		// priority: transactions > time
		switch {
		case cr.Spec.Transactions > 0:
			cmds = append(cmds, "-t", fmt.Sprintf("%d", cr.Spec.Transactions))
		case cr.Spec.Duration > 0:
			cmds = append(cmds, "-T", fmt.Sprintf("%d", cr.Spec.Duration))
		}

		if cr.Spec.Connect {
			cmds = append(cmds, "-C")
		}

		if cr.Spec.SelectOnly {
			cmds = append(cmds, "-S")
		}

		// TODO add func to parse extra args
		cmds = append(cmds, strings.Join(cr.Spec.ExtraArgs, " "))
	} else {
		// TODO add func to parse extra args
		cmds = []string{"pgbench", "-i", fmt.Sprintf("-s%d", cr.Spec.Scale), strings.Join(cr.Spec.ExtraArgs, " ")}
	}

	objectMeta := metav1.ObjectMeta{
		Name:      jobName,
		Namespace: cr.Namespace,
	}

	image := v1alpha1.ImageSpec{
		Name:  PgbenchName,
		Image: PgbenchImage,
		Cmds:  cmds,
		Env: []corev1.EnvVar{
			{
				Name:  "PGHOST",
				Value: cr.Spec.Target.Host,
			},
			{
				Name:  "PGPORT",
				Value: fmt.Sprintf("%d", cr.Spec.Target.Port),
			},
			{
				Name:  "PGUSER",
				Value: cr.Spec.Target.User,
			},
			{
				Name:  "PGPASSWORD",
				Value: cr.Spec.Target.Password,
			},
			{
				Name:  "PGDATABASE",
				Value: cr.Spec.Target.Database,
			},
		},
	}

	return utils.NewJob(jobName, cr.Namespace, objectMeta, image)
}
