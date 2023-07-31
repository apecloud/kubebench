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

	benchmarkv1alpha1 "github.com/apecloud/kubebench/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
	if err := r.Get(ctx, req.NamespacedName, &tpcc); err != nil {
		l.Error(err, "unable to fetch Tpcc")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	l.Info("reconciling Tpcc", "name", tpcc.Name)
	if tpcc.Status.Phase == benchmarkv1alpha1.Complete || tpcc.Status.Phase == benchmarkv1alpha1.Failed {
		return intctrlutil.Reconciled()
	}

	if tpcc.Status.Phase == "" {
		tpcc.Status.Phase = benchmarkv1alpha1.Running
		tpcc.Status.Total = len(tpcc.Spec.Threads)
		tpcc.Status.Completions = fmt.Sprintf("%d/%d", tpcc.Status.Succeeded, tpcc.Status.Total)
		if err := r.Status().Update(ctx, &tpcc); err != nil {
			return intctrlutil.RequeueWithError(err, l, "unable to update Tpcc status")
		}
	}

	// check if the job is already exist
	jobName := fmt.Sprintf("%s-%d", tpcc.Name, tpcc.Status.Succeeded)
	existed, err := utils.IsJobExisted(r.Client, ctx, jobName, tpcc.Namespace)
	if err != nil {
		return intctrlutil.RequeueWithError(err, l, "unable to check if job exists")
	}
	if existed {
		l.Info("job already exists", "job", jobName)
		// get the job status
		status, err := utils.GetJobStatus(r.Client, ctx, jobName, tpcc.Namespace)
		if err != nil {
			return intctrlutil.RequeueWithError(err, l, "unable to get job status")
		}
		l.Info("job status", "job", jobName, "status", status)

		// job is still running
		if status.Active > 0 {
			l.Info("job is still running", "job", jobName)
			return intctrlutil.RequeueAfter(intctrlutil.RequeueDuration)
		}

		// job is failed
		if status.Failed > 0 {
			l.Info("job is failed", "job", jobName)
			if err := r.Get(ctx, types.NamespacedName{Name: tpcc.Name, Namespace: tpcc.Namespace}, &tpcc); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to get tpcc")
			}

			// record the fail log
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, jobName, tpcc.Namespace, &tpcc.Status.Conditions, nil); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to record the fail log")
			}

			// delete the job
			if err := utils.DelteJob(r.Client, ctx, jobName, tpcc.Namespace); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to delete Job")
			}

			// update the tpcc status
			if err := r.Status().Update(ctx, &tpcc); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to update tpcc status")
			}

			return intctrlutil.Reconciled()
		}

		// job is completed
		if status.Succeeded > 0 {
			l.Info("job is succeeded", "jobName", jobName)
			if err := r.Get(ctx, types.NamespacedName{Name: tpcc.Name, Namespace: tpcc.Namespace}, &tpcc); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to get tpcc")
			}

			tpcc.Status.Succeeded += 1
			tpcc.Status.Completions = fmt.Sprintf("%d/%d", tpcc.Status.Succeeded, tpcc.Status.Total)

			// record the result
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, jobName, tpcc.Namespace, &tpcc.Status.Conditions, nil); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to record the fail log")
			}

			// delete the job
			if err := utils.DelteJob(r.Client, ctx, jobName, tpcc.Namespace); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to delete Job")
			}

			// update the tpcc status
			if err := r.Status().Update(ctx, &tpcc); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to update tpcc status")
			}
			return intctrlutil.RequeueAfter(intctrlutil.RequeueDuration)
		}

		// status is empty, job is creating
		return intctrlutil.RequeueAfter(intctrlutil.RequeueDuration)
	} else {
		// check the success number
		if tpcc.Status.Succeeded >= tpcc.Status.Total {
			if err := updatetpccStatus(r, ctx, &tpcc, benchmarkv1alpha1.Complete); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to update tpcc status")
			}
			return intctrlutil.Reconciled()
		}

		l.Info("Job isn't existed", "jobName", jobName)

		// create the job
		job := NewJob(&tpcc, jobName)

		if err := controllerutil.SetControllerReference(&tpcc, job, r.Scheme); err != nil {
			return intctrlutil.RequeueWithError(err, l, "unable to set ownerReference for job")
		}

		l.Info("creating job", "job", job.Name)
		if err := r.Create(ctx, job); err != nil {
			return intctrlutil.RequeueWithError(err, l, "unable to create job")
		}
		l.Info("job created", "job", job.Name)
		return intctrlutil.RequeueAfter(intctrlutil.RequeueDuration)
	}
}

func updatetpccStatus(r *TpccReconciler, ctx context.Context, tpcc *benchmarkv1alpha1.Tpcc, phase benchmarkv1alpha1.BenchmarkPhase) error {
	if err := r.Get(ctx, client.ObjectKeyFromObject(tpcc), tpcc); err != nil {
		return err
	}
	tpcc.Status.Phase = phase
	return r.Status().Update(ctx, tpcc)
}

// SetupWithManager sets up the controller with the Manager.
func (r *TpccReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&benchmarkv1alpha1.Tpcc{}).
		Complete(r)
}
