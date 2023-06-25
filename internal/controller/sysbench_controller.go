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
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	benchmarkv1alpha1 "github.com/apecloud/kubebench/api/v1alpha1"
)

// SysbenchReconciler reconciles a Sysbench object
type SysbenchReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=benchmark.kubebench.io,resources=sysbenches,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=benchmark.kubebench.io,resources=sysbenches/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=benchmark.kubebench.io,resources=sysbenches/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Sysbench object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *SysbenchReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var sysbench benchmarkv1alpha1.Sysbench
	if err := r.Get(ctx, req.NamespacedName, &sysbench); err != nil {
		l.Error(err, "unable to fetch Sysbench")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Run to one completion
	if sysbench.Status.Phase == benchmarkv1alpha1.Complete {
		return ctrl.Result{}, nil
	}

	sysbench.Status.Phase = benchmarkv1alpha1.Running
	if err := r.Status().Update(ctx, &sysbench); err != nil {
		l.Error(err, "unable to update Sysbench status")
		return ctrl.Result{}, err
	}

	constructJobForSysbench := func(sysbench *benchmarkv1alpha1.Sysbench) (*batchv1.Job, error) {
		job := &batchv1.Job{
			ObjectMeta: ctrl.ObjectMeta{
				Labels:      map[string]string{},
				Annotations: map[string]string{},
				Name:        fmt.Sprintf("%s", sysbench.Name),
				Namespace:   sysbench.Namespace,
			},
			Spec: *sysbench.Spec.JobTemplate.Spec.DeepCopy(),
		}
		for k, v := range sysbench.Spec.JobTemplate.Annotations {
			job.Annotations[k] = v
		}
		for k, v := range sysbench.Spec.JobTemplate.Labels {
			job.Labels[k] = v
		}
		if err := ctrl.SetControllerReference(sysbench, job, r.Scheme); err != nil {
			return nil, err
		}

		return job, nil
	}

	isJobExisted := func(name, namespace string) (existed bool, err error) {
		var job batchv1.Job
		if err := r.Get(ctx, req.NamespacedName, &job); err != nil {
			return false, client.IgnoreNotFound(err)
		}

		return true, nil
	}

	isJobFinished := func(name, namespace string) (finished bool, err error) {
		var job batchv1.Job
		if err := r.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &job); err != nil {
			return false, client.IgnoreNotFound(err)
		}

		finished = job.Status.CompletionTime != nil
		return finished, nil
	}

	// check if the job already exists
	existed, err := isJobExisted(sysbench.Name, sysbench.Namespace)
	if err != nil {
		l.Error(err, "unable to check if the Job already exists")
		return ctrl.Result{}, err
	}
	if existed {
		l.V(1).Info("Job already exists", "job", sysbench.Name)
	} else {
		// construct the job from the template
		job, err := constructJobForSysbench(&sysbench)
		if err != nil {
			l.Error(err, "unable to construct Job from template")
			return ctrl.Result{}, err
		}

		// actually create the job on the cluster
		if err := r.Create(ctx, job); err != nil {
			l.Error(err, "unable to create Job for Sysbench")
			return ctrl.Result{}, err
		}
	}

	// check if the job is finished
	finished, err := isJobFinished(sysbench.Name, sysbench.Namespace)
	if err != nil {
		l.Error(err, "unable to check if the Job is finished")
		return ctrl.Result{}, err
	}
	if !finished {
		l.V(1).Info("Job is not finished", "job", sysbench.Name)
		return ctrl.Result{Requeue: true}, nil
	}

	// The sysbench could have been modified since the last time we got it
	if err := r.Get(ctx, req.NamespacedName, &sysbench); err != nil {
		l.Error(err, "unable to fetch Sysbench")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	sysbench.Status.Phase = benchmarkv1alpha1.Complete
	if err := r.Status().Update(ctx, &sysbench); err != nil {
		l.Error(err, "unable to update Sysbench status")
		return ctrl.Result{}, err
	}
	l.Info("Sysbench benchmark completed", "sysbench", sysbench.Name)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SysbenchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&benchmarkv1alpha1.Sysbench{}).
		Complete(r)
}
