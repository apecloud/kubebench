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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	benchmarkv1alpha1 "github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/internal/controllerutil"
	"github.com/apecloud/kubebench/internal/utils"
)

const (
	SysbenchJobNamePrefix = "sysbench-"
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
	if err := r.Get(ctx, req.NamespacedName, &sysbench); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Run to one completion
	if sysbench.Status.Phase == benchmarkv1alpha1.Complete || sysbench.Status.Phase == benchmarkv1alpha1.Failed {
		return controllerutil.Reconciled()
	}

	if sysbench.Status.Phase == "" {
		sysbench.Status.Phase = benchmarkv1alpha1.Running
		sysbench.Status.Total = len(sysbench.Spec.Threads) * len(sysbench.Spec.Types)
		sysbench.Status.Completions = fmt.Sprintf("%d/%d", sysbench.Status.Succeeded, sysbench.Status.Total)
		if err := r.Status().Update(ctx, &sysbench); err != nil {
			return controllerutil.RequeueWithError(err, l, "unable to update sysbench status")
		}
	}

	// check if the job already exists
	jobName := SysbenchJobNamePrefix + fmt.Sprintf("%s-%d", sysbench.Name, sysbench.Status.Succeeded)
	existed, err := utils.IsJobExisted(r.Client, ctx, jobName, sysbench.Namespace)
	if err != nil {
		return controllerutil.RequeueWithError(err, l, "unable to check if job exists")
	}
	if existed {
		l.Info("job already exists", "job", jobName)
		// get the job status
		status, err := utils.GetJobStatus(r.Client, ctx, jobName, sysbench.Namespace)
		if err != nil {
			return controllerutil.RequeueWithError(err, l, "unable to get job status")
		}
		l.Info("job status", "job", jobName, "status", status)

		// job is still running
		if status.Active > 0 {
			l.Info("job is still running", "job", jobName)
			return controllerutil.RequeueAfter(controllerutil.RequeueDuration)
		}

		// job is failed
		if status.Failed > 0 {
			l.Info("job is failed", "job", jobName)
			if err := r.Get(ctx, types.NamespacedName{Name: sysbench.Name, Namespace: sysbench.Namespace}, &sysbench); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to get sysbench")
			}

			// record the fail log
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, jobName, sysbench.Namespace, &sysbench.Status.Conditions, nil); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to record the fail log")
			}

			// delete the job
			if err := utils.DelteJob(r.Client, ctx, jobName, sysbench.Namespace); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to delete Job")
			}

			// update the sysbench status
			if err := r.Status().Update(ctx, &sysbench); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to update sysbench status")
			}

			return controllerutil.Reconciled()
		}

		// job is completed
		if status.Succeeded > 0 {
			l.Info("job is succeeded", "jobName", jobName)
			if err := r.Get(ctx, types.NamespacedName{Name: sysbench.Name, Namespace: sysbench.Namespace}, &sysbench); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to get sysbench")
			}

			sysbench.Status.Succeeded += 1
			sysbench.Status.Completions = fmt.Sprintf("%d/%d", sysbench.Status.Succeeded, sysbench.Status.Total)

			// record the result
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, jobName, sysbench.Namespace, &sysbench.Status.Conditions, ParseSysBench); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to record the fail log")
			}

			// delete the job
			if err := utils.DelteJob(r.Client, ctx, jobName, sysbench.Namespace); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to delete Job")
			}

			// update the sysbench status
			if err := r.Status().Update(ctx, &sysbench); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to update sysbench status")
			}
			return controllerutil.RequeueAfter(controllerutil.RequeueDuration)
		}

		// status is empty, job is creating
		return controllerutil.RequeueAfter(controllerutil.RequeueDuration)
	} else {
		// check the success number
		if sysbench.Status.Succeeded >= sysbench.Status.Total {
			if err := updateSysbenchStatus(r, ctx, &sysbench, benchmarkv1alpha1.Complete); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to update sysbench status")
			}
			return controllerutil.Reconciled()
		}

		l.Info("Job isn't existed", "jobName", jobName)

		// create the job
		job := NewJob(&sysbench, jobName)
		l.Info("creating job", "job", job.Name)
		if err := r.Create(ctx, job); err != nil {
			return controllerutil.RequeueWithError(err, l, "unable to create job")
		}
		l.Info("job created", "job", job.Name)
		return controllerutil.RequeueAfter(controllerutil.RequeueDuration)
	}
}

func updateSysbenchStatus(r *SysbenchReconciler, ctx context.Context, sysbench *benchmarkv1alpha1.Sysbench, status benchmarkv1alpha1.BenchmarkPhase) error {
	// the sysbench could be deleted before the job is completed
	if err := r.Get(ctx, client.ObjectKeyFromObject(sysbench), sysbench); err != nil {
		return err
	}
	sysbench.Status.Phase = status
	return r.Status().Update(ctx, sysbench)
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
