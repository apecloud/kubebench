package redisbenchmark

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/internal/utils"
)

const (
	RedisBenchmark  = "redis-benchmark"
	RedisBenchImage = "redis:latest"
)

func NewJob(cr *v1alpha1.Redis, jobName string) *batchv1.Job {
	cmds := []string{"redis-benchmark",
		"-h", cr.Spec.Target.Host,
		"-p", fmt.Sprintf("%d", cr.Spec.Target.Port),
		"-a", cr.Spec.Target.Password,
		"-c", fmt.Sprintf("%d", cr.Spec.Clients[cr.Status.Succeeded]),
		"-n", fmt.Sprintf("%d", cr.Spec.Requests),
		"-d", fmt.Sprintf("%d", cr.Spec.DataSize),
		"-P", fmt.Sprintf("%d", cr.Spec.Pipeline),
		"--csv",
	}

	if !cr.Spec.KeepAlive {
		cmds = append(cmds, "-k", "0")
	}

	if cr.Spec.KeySpace != 0 {
		cmds = append(cmds, "-r", fmt.Sprintf("%d", cr.Spec.KeySpace))
	}

	objectMeta := metav1.ObjectMeta{
		// add redis prefix to job name to avoid name conflict in different benchmarks
		Name:      jobName,
		Namespace: cr.Namespace,
	}

	image := v1alpha1.ImageSpec{
		Name:  RedisBenchmark,
		Image: RedisBenchImage,
		Cmds:  cmds,
	}

	return utils.NewJob(jobName, cr.Namespace, objectMeta, image)
}
