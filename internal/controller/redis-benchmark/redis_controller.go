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

package redisbenchmark

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	benchmarkv1alpha1 "github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/internal/utils"
)

// RedisReconciler reconciles a Redis object
type RedisReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RestConfig *rest.Config
}

//+kubebuilder:rbac:groups=benchmark.kubebench.io,resources=redis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=benchmark.kubebench.io,resources=redis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=benchmark.kubebench.io,resources=redis/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Redis object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *RedisReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var bench benchmarkv1alpha1.Redis
	if err := r.Get(ctx, req.NamespacedName, &bench); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if bench.Status.Phase == benchmarkv1alpha1.Complete || bench.Status.Phase == benchmarkv1alpha1.Failed {
		return ctrl.Result{}, nil
	}

	if bench.Status.Phase == "" {
		bench.Status.Phase = benchmarkv1alpha1.Running
		bench.Status.Total = len(bench.Spec.Clients)
		bench.Status.Completions = fmt.Sprintf("%d/%d", bench.Status.Succeeded, bench.Status.Total)
		if err := r.Status().Update(ctx, &bench); err != nil {
			l.Error(err, "unable to update bench status")
			return ctrl.Result{}, err
		}
	}

	jobName := fmt.Sprintf("redis-%s-%d", bench.Name, bench.Status.Succeeded)

	// check if the job already exists
	existed, err := utils.IsJobExisted(r.Client, ctx, jobName, bench.Namespace)
	if err != nil {
		l.Error(err, "unable to check if the Job already exists")
		return ctrl.Result{}, err
	}
	if existed {
		l.Info("Job already exists", "jobName", jobName)
		// get the job status
		status, err := utils.GetJobStatus(r.Client, ctx, jobName, bench.Namespace)
		if err != nil {
			l.Error(err, "unable to get Job status")
			return ctrl.Result{}, err
		}
		l.Info("Job status", "jobName", jobName, "status", status)

		// job is still running
		if status.Active > 0 {
			l.Info("Job is still running", "jobName", jobName)
			return ctrl.Result{Requeue: true}, nil
		}

		// job is failed
		if status.Failed > 0 {
			l.Info("Job is failed", "jobName", jobName)
			if err := r.Get(ctx, types.NamespacedName{Name: bench.Name, Namespace: bench.Namespace}, &bench); err != nil {
				l.Error(err, "unable to update bench status")
				return ctrl.Result{}, err
			}

			// update the status
			bench.Status.Phase = benchmarkv1alpha1.Failed

			// record the fail log
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, jobName, bench.Namespace, &bench.Status.Conditions, nil); err != nil {
				l.Error(err, "unable to record the fail log")
				return ctrl.Result{}, err
			}

			// delete the job
			l.V(1).Info("delete the Job", "jobName", jobName)
			if err := utils.DelteJob(r.Client, ctx, jobName, bench.Namespace); err != nil {
				l.Error(err, "unable to delete Job")
				return ctrl.Result{}, err
			}

			// update the bench status
			if err := r.Status().Update(ctx, &bench); err != nil {
				l.Error(err, "unable to update bench status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}

		if status.Succeeded > 0 {
			l.Info("Job is succeeded", "jobName", jobName)
			if err := r.Get(ctx, types.NamespacedName{Name: bench.Name, Namespace: bench.Namespace}, &bench); err != nil {
				l.Error(err, "unable to update bench status")
				return ctrl.Result{}, err
			}

			bench.Status.Succeeded += 1
			bench.Status.Completions = fmt.Sprintf("%d/%d", bench.Status.Succeeded, bench.Status.Total)

			// TODO add func to process log
			// record the result
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, jobName, bench.Namespace, &bench.Status.Conditions, nil); err != nil {
				l.Error(err, "unable to record the fail log")
				return ctrl.Result{}, err
			}

			// delete the job
			l.V(1).Info("delete the Job", "jobName", jobName)
			if err := utils.DelteJob(r.Client, ctx, jobName, bench.Namespace); err != nil {
				l.Error(err, "unable to delete Job")
				return ctrl.Result{}, err
			}

			// update the bench status
			if err := r.Status().Update(ctx, &bench); err != nil {
				l.Error(err, "unable to update bench status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}

		// status is empty, job is creating
		return ctrl.Result{Requeue: true}, nil
	} else {
		// check the success number
		if bench.Status.Succeeded >= bench.Status.Total {
			if err := updatebenchStatus(r, ctx, &bench, benchmarkv1alpha1.Complete); err != nil {
				l.Error(err, "unable to update bench status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		l.Info("Job isn't existed", "jobName", jobName)

		// don't have job, and the bench is not complete
		// create a new job
		job := NewJob(&bench, jobName)
		l.Info("create a new Job", "jobName", job.Name)
		if err := r.Create(ctx, job); err != nil {
			l.Error(err, "unable to create Job")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}
}

func updatebenchStatus(r *RedisReconciler, ctx context.Context, bench *benchmarkv1alpha1.Redis, status benchmarkv1alpha1.BenchmarkPhase) error {
	// The bench could have been modified since the last time we got it
	if err := r.Get(ctx, types.NamespacedName{Name: bench.Name, Namespace: bench.Namespace}, bench); err != nil {
		return err
	}
	bench.Status.Phase = status
	if err := r.Status().Update(ctx, bench); err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&benchmarkv1alpha1.Redis{}).
		Complete(r)
}
