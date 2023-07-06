package sysbench

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
	SysbenchName  = "sysbench"
	SysbenchImage = "registry.cn-hangzhou.aliyuncs.com/apecloud/customsuites:latest"
)

func NewJob(cr *v1alpha1.Sysbench, jobName string) *batchv1.Job {
	value := fmt.Sprintf("mode:%s", "all")
	value = fmt.Sprintf("%s,driver:%s", value, cr.Spec.Target.Driver)
	value = fmt.Sprintf("%s,host:%s", value, cr.Spec.Target.Host)
	value = fmt.Sprintf("%s,port:%d", value, cr.Spec.Target.Port)
	value = fmt.Sprintf("%s,user:%s", value, cr.Spec.Target.User)
	value = fmt.Sprintf("%s,password:%s", value, cr.Spec.Target.Password)
	value = fmt.Sprintf("%s,db:%s", value, cr.Spec.Target.Database)
	value = fmt.Sprintf("%s,tables:%d", value, cr.Spec.Tables)
	value = fmt.Sprintf("%s,size:%d", value, cr.Spec.Size)
	value = fmt.Sprintf("%s,times:%d", value, cr.Spec.Duration)
	threads := make([]string, 0)
	for _, thread := range cr.Spec.Threads {
		threads = append(threads, fmt.Sprintf("%d", thread))
	}
	value = fmt.Sprintf("%s,threads:%s", value, strings.Join(threads, " "))
	value = fmt.Sprintf("%s,type:%s", value, strings.Join(cr.Spec.Types, " "))
	// TODO add func to parse extra args
	value = fmt.Sprintf("%s,others:%s", value, strings.Join(cr.Spec.ExtraArgs, " "))

	objectMeta := metav1.ObjectMeta{
		Name:      jobName,
		Namespace: cr.Namespace,
	}

	image := v1alpha1.ImageSpec{
		Name:  SysbenchName,
		Image: SysbenchImage,
		Env: []corev1.EnvVar{
			{
				Name:  "TYPE",
				Value: "2",
			},
			{
				Name:  "FLAG",
				Value: "0",
			},
			{
				Name:  "CONFIGS",
				Value: value,
			},
		},
	}

	return utils.NewJob(jobName, cr.Namespace, objectMeta, image)
}
