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

package ycsb

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	benchmarkv1alpha1 "github.com/apecloud/kubebench/api/v1alpha1"
	intctrlutil "github.com/apecloud/kubebench/internal/controllerutil"
	"github.com/apecloud/kubebench/internal/utils"
)

// YcsbReconciler reconciles a Ycsb object
type YcsbReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RestConfig *rest.Config
}

//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=ycsbs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=ycsbs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=ycsbs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ycsb object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *YcsbReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var ycsb benchmarkv1alpha1.Ycsb
	if err := r.Get(ctx, req.NamespacedName, &ycsb); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// run if bench completion
	if ycsb.Status.Phase == benchmarkv1alpha1.Complete || ycsb.Status.Phase == benchmarkv1alpha1.Failed {
		return intctrlutil.Reconciled()
	}

	// run if bench not started
	if ycsb.Status.Phase == "" {
		ycsb.Status.Phase = benchmarkv1alpha1.Running
		ycsb.Status.Total = len(ycsb.Spec.Threads)
		ycsb.Status.Completions = fmt.Sprintf("%d/%d", ycsb.Status.Succeeded, ycsb.Status.Total)
		if err := r.Status().Update(ctx, &ycsb); err != nil {
			return intctrlutil.RequeueWithError(err, l, "unable to update ycsb status")
		}
	}

	var jobName string
	if ycsb.Status.Ready {
		jobName = fmt.Sprintf("%s-%d", ycsb.Name, ycsb.Status.Succeeded)
	} else {
		jobName = fmt.Sprintf("%s-init", ycsb.Name)
	}

	// check if the job already exists
	existed, err := utils.IsJobExisted(r.Client, ctx, jobName, ycsb.Namespace)
	if err != nil {
		return intctrlutil.RequeueWithError(err, l, "failed to check if job exists")
	}
	if existed {
		l.Info("job already exists", "job", jobName)
		// get the job status
		status, err := utils.GetJobStatus(r.Client, ctx, jobName, ycsb.Namespace)
		if err != nil {
			return intctrlutil.RequeueWithError(err, l, "failed to get job status")
		}
		l.Info("Job status", "jobName", jobName, "status", status)

		// job is still running
		if status.Active > 0 {
			l.Info("Job is still running", "jobName", jobName)
			return intctrlutil.RequeueAfter(intctrlutil.RequeueDuration)
		}

		// job is failed
		if status.Failed > 0 {
			l.Info("Job is failed", "jobName", jobName)
			if err := r.Get(ctx, types.NamespacedName{Name: ycsb.Name, Namespace: ycsb.Namespace}, &ycsb); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to update ycsb status")
			}

			// update the status
			ycsb.Status.Phase = benchmarkv1alpha1.Failed

			// record the fail log
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, jobName, ycsb.Namespace, &ycsb.Status.Conditions, nil); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to record the fail log")
			}

			// delete the job
			l.V(1).Info("delete the Job", "jobName", jobName)
			if err := utils.DelteJob(r.Client, ctx, jobName, ycsb.Namespace); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to delete Job")
			}

			// update the ycsb status
			if err := r.Status().Update(ctx, &ycsb); err != nil {
				l.Error(err, "unable to update ycsb status")
				return intctrlutil.RequeueWithError(err, l, "unable to update ycsb status")
			}

			return ctrl.Result{}, nil
		}

		if status.Succeeded > 0 {
			l.Info("Job is succeeded", "jobName", jobName)
			if err := r.Get(ctx, types.NamespacedName{Name: ycsb.Name, Namespace: ycsb.Namespace}, &ycsb); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to update ycsb status")
			}

			if !ycsb.Status.Ready {
				ycsb.Status.Ready = true
			} else {
				ycsb.Status.Succeeded += 1
			}
			ycsb.Status.Completions = fmt.Sprintf("%d/%d", ycsb.Status.Succeeded, ycsb.Status.Total)

			// record the result
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, jobName, ycsb.Namespace, &ycsb.Status.Conditions, nil); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to record the fail log")
			}

			// delete the job
			l.V(1).Info("delete the Job", "jobName", jobName)
			if err := utils.DelteJob(r.Client, ctx, jobName, ycsb.Namespace); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to delete Job")
			}

			// update the ycsb status
			if err := r.Status().Update(ctx, &ycsb); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to update ycsb status")
			}
			return intctrlutil.RequeueAfter(intctrlutil.RequeueDuration)
		}

		// status is empty, job is creating
		return intctrlutil.RequeueAfter(intctrlutil.RequeueDuration)
	} else {
		// check the success number
		if ycsb.Status.Succeeded >= ycsb.Status.Total {
			if err := updateYcsbStatus(r, ctx, &ycsb, benchmarkv1alpha1.Complete); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to update ycsb status")
			}
			return intctrlutil.Reconciled()
		}

		l.Info("Job isn't existed", "jobName", jobName)

		// don't have job, and the pgbench is not complete
		// create a new job
		job := NewJob(&ycsb, jobName)
		l.Info("create a new Job", "jobName", job.Name)

		if err := controllerutil.SetOwnerReference(&ycsb, job, r.Scheme); err != nil {
			return intctrlutil.RequeueWithError(err, l, "unable to set ownerReference for Job")
		}

		if err := r.Create(ctx, job); err != nil {
			return intctrlutil.RequeueWithError(err, l, "unable to create Job")
		}
		return intctrlutil.RequeueAfter(intctrlutil.RequeueDuration)
	}
}

func updateYcsbStatus(r *YcsbReconciler, ctx context.Context, ycsb *benchmarkv1alpha1.Ycsb, status benchmarkv1alpha1.BenchmarkPhase) error {
	// The ycsb could have been modified since the last time we got it
	if err := r.Get(ctx, types.NamespacedName{Name: ycsb.Name, Namespace: ycsb.Namespace}, ycsb); err != nil {
		return err
	}
	ycsb.Status.Phase = status
	if err := r.Status().Update(ctx, ycsb); err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *YcsbReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&benchmarkv1alpha1.Ycsb{}).
		Complete(r)
}
