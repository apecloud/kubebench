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

package tpcc

import (
	"context"
	"fmt"
	"time"

	benchmarkv1alpha1 "github.com/apecloud/kubebench/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	intctrlutil "github.com/apecloud/kubebench/internal/controllerutil"
	"github.com/apecloud/kubebench/internal/utils"
)

// TpccReconciler reconciles a Tpcc object
type TpccReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RestConfig *rest.Config
}

//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=tpccs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=tpccs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=tpccs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Tpcc object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *TpccReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var tpcc benchmarkv1alpha1.Tpcc
	var err error
	if err = r.Get(ctx, req.NamespacedName, &tpcc); err != nil {
		l.Error(err, "unable to fetch Tpcc")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	l.Info("reconciling Tpcc", "name", tpcc.Name)
	if tpcc.Status.Phase == benchmarkv1alpha1.Complete || tpcc.Status.Phase == benchmarkv1alpha1.Failed {
		return intctrlutil.Reconciled()
	}

	l.Info("reconciling tpcc", "tpcc mode", tpcc.Spec.Mode)

	jobs := NewJobs(&tpcc)

	tpcc.Status.Phase = benchmarkv1alpha1.Running
	tpcc.Status.Total = len(jobs)
	tpcc.Status.Completions = fmt.Sprintf("%d/%d", tpcc.Status.Succeeded, tpcc.Status.Total)
	if err := r.Status().Update(ctx, &tpcc); err != nil {
		return intctrlutil.RequeueWithError(err, l, "unable to update tpcc status")
	}

	if err = r.Status().Update(ctx, &tpcc); err != nil {
		return intctrlutil.RequeueWithError(err, l, "update to update tpcc status")
	}

	for _, job := range jobs {
		if err = controllerutil.SetOwnerReference(&tpcc, job, r.Scheme); err != nil {
			return intctrlutil.RequeueWithError(err, l, "failed to set owner reference for job")
		}

		if err = r.Create(ctx, job); err != nil {
			return intctrlutil.RequeueWithError(err, l, "failed to create job")
		}

		l.Info("created job", "job", job.Name)
		// wait for the job to complete, then update the tpcc status

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
				tpcc.Status.Phase = benchmarkv1alpha1.Failed
			}

			// job is completed
			if status.Succeeded > 0 {
				l.Info("job is succeeded", "jobName", job.Name)
				tpcc.Status.Succeeded += 1
				tpcc.Status.Completions = fmt.Sprintf("%d/%d", tpcc.Status.Succeeded, tpcc.Status.Total)
			}

			// record the result
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, job.Name, tpcc.Namespace, &tpcc.Status.Conditions, nil); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to record the log")
			}

			break
		}

		r.Status().Update(ctx, &tpcc)

		if err != nil {
			return intctrlutil.RequeueWithError(err, l, "")
		}
	}

	tpcc.Status.Phase = benchmarkv1alpha1.Complete
	r.Status().Update(ctx, &tpcc)
	return intctrlutil.Reconciled()
}

// SetupWithManager sets up the controller with the Manager.
func (r *TpccReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&benchmarkv1alpha1.Tpcc{}).
		Complete(r)
}
