/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sysbench

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	benchmarkv1alpha1 "github.com/apecloud/kubebench/api/v1alpha1"
	intctrlutil "github.com/apecloud/kubebench/internal/controllerutil"
	"github.com/apecloud/kubebench/internal/utils"
)

// SysbenchReconciler reconciles a Sysbench object
type SysbenchReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RestConfig *rest.Config
}

//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=sysbenches,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=sysbenches/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=sysbenches/finalizers,verbs=update

// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;delete;deletecollection
// +kubebuilder:rbac:groups=core,resources=pods/log,verbs=get;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Sysbench object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.1controllerutil.RequeueDuration.0/pkg/reconcile
func (r *SysbenchReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var sysbench benchmarkv1alpha1.Sysbench
	var err error
	if err = r.Get(ctx, req.NamespacedName, &sysbench); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Run to one completion
	if sysbench.Status.Phase == benchmarkv1alpha1.Complete || sysbench.Status.Phase == benchmarkv1alpha1.Failed {
		return intctrlutil.Reconciled()
	}

	l.Info("reconciling sysbench", "sysbench mode", sysbench.Spec.Mode)

	jobs := NewJobs(&sysbench)

	sysbench.Status.Phase = benchmarkv1alpha1.Running
	sysbench.Status.Total = len(jobs)
	sysbench.Status.Completions = fmt.Sprintf("%d/%d", sysbench.Status.Succeeded, sysbench.Status.Total)
	if err := r.Status().Update(ctx, &sysbench); err != nil {
		return intctrlutil.RequeueWithError(err, l, "unable to update sysbench status")
	}

	if err = r.Status().Update(ctx, &sysbench); err != nil {
		return intctrlutil.RequeueWithError(err, l, "update to update sysbench status")
	}

	for _, job := range jobs {
		if err = controllerutil.SetOwnerReference(&sysbench, job, r.Scheme); err != nil {
			return intctrlutil.RequeueWithError(err, l, "failed to set owner reference for job")
		}

		if err = r.Create(ctx, job); err != nil {
			return intctrlutil.RequeueWithError(err, l, "failed to create job")
		}

		l.Info("created job", "job", job.Name)
		// wait for the job to complete, then update the sysbench status

		for {
			// sleep for a while to avoid too many requests
			time.Sleep(time.Second)

			status, err := utils.GetJobStatus(r.Client, ctx, job.Name, job.Namespace)
			if err != nil {
				l.Error(err, "failed to get job status")
				break
			}

			// job is still running
			if status.Active > 0 {
				l.Info("job is still running", "job", job.Name)
				continue
			}

			// job is failed
			if status.Failed > 0 {
				l.Info("job is failed", "job", job.Name)
				sysbench.Status.Phase = benchmarkv1alpha1.Failed
			}

			// job is completed
			if status.Succeeded > 0 {
				l.Info("job is succeeded", "jobName", job.Name)
				sysbench.Status.Succeeded += 1
				sysbench.Status.Completions = fmt.Sprintf("%d/%d", sysbench.Status.Succeeded, sysbench.Status.Total)
			}

			// record the result
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, job.Name, sysbench.Namespace, &sysbench.Status.Conditions, ParseSysBench); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to record the log")
			}

			break
		}

		r.Status().Update(ctx, &sysbench)

		if err != nil {
			return intctrlutil.RequeueWithError(err, l, "")
		}
	}

	sysbench.Status.Phase = benchmarkv1alpha1.Complete
	r.Status().Update(ctx, &sysbench)
	return intctrlutil.Reconciled()
}

// SetupWithManager sets up the controller with the Manager.
func (r *SysbenchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&benchmarkv1alpha1.Sysbench{}).
		Complete(r)
}

func ParseSysBench(msg string) string {
	result := ""
	lines := strings.Split(msg, "\n")
	index := len(lines)

	for i, l := range lines {
		if strings.Contains(l, "SQL statistics") {
			index = i
			result += fmt.Sprintf("%s\n", l)
			break
		}
	}

	for i := index + 1; i < len(lines); i++ {
		if lines[i] != "" {
			// align the output
			result += fmt.Sprintf("%*s\n", len(lines[i])+27, lines[i])
		}
	}

	// delete the last \n
	return strings.TrimSpace(result)
}
