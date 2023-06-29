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

package controller

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	benchmarkv1alpha1 "github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/internal/utils"
)

// PgbenchReconciler reconciles a Pgbench object
type PgbenchReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RestConfig *rest.Config
}

//+kubebuilder:rbac:groups=benchmark.kubebench.io,resources=pgbenches,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=benchmark.kubebench.io,resources=pgbenches/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=benchmark.kubebench.io,resources=pgbenches/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pgbench object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *PgbenchReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var pgbench benchmarkv1alpha1.Pgbench
	if err := r.Get(ctx, req.NamespacedName, &pgbench); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Run to one completion
	if pgbench.Status.Phase == benchmarkv1alpha1.Complete || pgbench.Status.Phase == benchmarkv1alpha1.Failed {
		return ctrl.Result{}, nil
	}

	pgbench.Status.Phase = benchmarkv1alpha1.Running
	if err := r.Status().Update(ctx, &pgbench); err != nil {
		l.Error(err, "unable to update pgbench status")
		return ctrl.Result{}, err
	}

	// check if the job already exists
	existed, err := utils.IsJobExisted(r.Client, ctx, pgbench.Name, pgbench.Namespace)
	if err != nil {
		l.Error(err, "unable to check if the Job already exists")
		return ctrl.Result{}, err
	}
	if !existed {
		// construct the job from the template
		job := utils.NewJob(pgbench.Name, pgbench.Namespace, pgbench.Spec.Image, pgbench.Spec.PodConfig)

		// actually create the job on the cluster
		if err := r.Create(ctx, job); err != nil {
			l.Error(err, "unable to create Job for pgbench")
			return ctrl.Result{}, err
		}
		l.V(1).Info("Job created", "job", pgbench.Name)
		return ctrl.Result{Requeue: true}, nil
	}

	// get the job status
	var phase benchmarkv1alpha1.BenchmarkPhase
	status, err := utils.GetJobStatus(r.Client, ctx, pgbench.Name, pgbench.Namespace)
	if err != nil {
		l.Error(err, "unable to get Job status")
		return ctrl.Result{}, err
	}

	switch {
	case status.Succeeded > 0:
		l.V(1).Info("Job succeeded", "job", pgbench.Name)
		phase = benchmarkv1alpha1.Complete
	case status.Failed > 0:
		l.V(1).Info("Job failed", "job", pgbench.Name)
		phase = benchmarkv1alpha1.Failed
	case status.Active > 0:
		return ctrl.Result{Requeue: true}, nil
	default:
		return ctrl.Result{}, nil
	}

	// The pgbench could have been modified since the last time we got it
	if err := r.Get(ctx, req.NamespacedName, &pgbench); err != nil {
		l.Error(err, "unable to fetch pgbench")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	podList, err := utils.GetPodListFromJob(r.Client, ctx, pgbench.Name, pgbench.Namespace)
	if err != nil {
		l.Error(err, "unable to get pods from Job")
		return ctrl.Result{}, err
	}
	for _, pod := range podList.Items {
		// get the logs from the pod
		msg, err := utils.GetLogFromPod(r.RestConfig, ctx, pod.Name, pod.Namespace)
		if err != nil {
			l.Error(err, "unable to get logs from pod")
			return ctrl.Result{}, err
		}

		// save the result to the status
		meta.SetStatusCondition(&pgbench.Status.Conditions, metav1.Condition{
			Type:               "Complete",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: pgbench.Generation,
			Reason:             "JobFinished",
			Message:            msg,
			LastTransitionTime: metav1.Now(),
		})
	}

	// delete the job
	if err := utils.DelteJob(r.Client, ctx, pgbench.Name, pgbench.Namespace); err != nil {
		l.Error(err, "unable to delete Job")
		return ctrl.Result{}, err
	}

	// update pgbench status
	pgbench.Status.Phase = phase
	if err := r.Status().Update(ctx, &pgbench); err != nil {
		l.Error(err, "unable to update pgbench status")
		return ctrl.Result{}, err
	}
	l.Info("pgbench benchmark completed", "pgbench", pgbench.Name)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PgbenchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&benchmarkv1alpha1.Pgbench{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
