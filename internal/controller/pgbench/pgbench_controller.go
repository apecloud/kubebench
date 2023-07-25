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

package pgbench

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
	PgbenchJobNamePrefix = "pgbench-"
)

// PgbenchReconciler reconciles a Pgbench object
type PgbenchReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RestConfig *rest.Config
}

//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=pgbenches,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=pgbenches/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=pgbenches/finalizers,verbs=update

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

	if pgbench.Status.Phase == benchmarkv1alpha1.Complete || pgbench.Status.Phase == benchmarkv1alpha1.Failed {
		return controllerutil.Reconciled()
	}

	if pgbench.Status.Phase == "" {
		pgbench.Status.Phase = benchmarkv1alpha1.Running
		pgbench.Status.Total = len(pgbench.Spec.Clients)
		pgbench.Status.Completions = fmt.Sprintf("%d/%d", pgbench.Status.Succeeded, pgbench.Status.Total)
		if err := r.Status().Update(ctx, &pgbench); err != nil {
			return controllerutil.RequeueWithError(err, l, "update to update pgbench status")
		}
	}

	var jobName string
	if pgbench.Status.Ready {
		jobName = PgbenchJobNamePrefix + fmt.Sprintf("%s-%d", pgbench.Name, pgbench.Status.Succeeded)
	} else {
		jobName = PgbenchJobNamePrefix + fmt.Sprintf("%s-init", pgbench.Name)
	}

	// check if the job already exists
	existed, err := utils.IsJobExisted(r.Client, ctx, jobName, pgbench.Namespace)
	if err != nil {
		return controllerutil.RequeueWithError(err, l, "unable to check if the Job already exists")
	}
	if existed {
		l.Info("Job already exists", "jobName", jobName)
		// get the job status
		status, err := utils.GetJobStatus(r.Client, ctx, jobName, pgbench.Namespace)
		if err != nil {
			return controllerutil.RequeueWithError(err, l, "unable to get Job status")
		}
		l.Info("Job status", "jobName", jobName, "status", status)

		// job is still running
		if status.Active > 0 {
			l.Info("Job is still running", "jobName", jobName)
			return controllerutil.RequeueAfter(controllerutil.RequeueDuration)
		}

		// job is failed
		if status.Failed > 0 {
			l.Info("Job is failed", "jobName", jobName)
			if err := r.Get(ctx, types.NamespacedName{Name: pgbench.Name, Namespace: pgbench.Namespace}, &pgbench); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to update pgbench status")
			}

			// update the status
			pgbench.Status.Phase = benchmarkv1alpha1.Failed

			// record the fail log
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, jobName, pgbench.Namespace, &pgbench.Status.Conditions, nil); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to record the fail log")
			}

			// delete the job
			l.V(1).Info("delete the Job", "jobName", jobName)
			if err := utils.DelteJob(r.Client, ctx, jobName, pgbench.Namespace); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to delete Job")
			}

			// update the pgbench status
			if err := r.Status().Update(ctx, &pgbench); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to update pgbench status")
			}

			return ctrl.Result{}, nil
		}

		if status.Succeeded > 0 {
			l.Info("Job is succeeded", "jobName", jobName)
			if err := r.Get(ctx, types.NamespacedName{Name: pgbench.Name, Namespace: pgbench.Namespace}, &pgbench); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to update pgbench status")
			}

			if !pgbench.Status.Ready {
				pgbench.Status.Ready = true
			} else {
				pgbench.Status.Succeeded += 1
			}
			pgbench.Status.Completions = fmt.Sprintf("%d/%d", pgbench.Status.Succeeded, pgbench.Status.Total)

			// record the result
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, jobName, pgbench.Namespace, &pgbench.Status.Conditions, ParsePgbench); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to record the fail log")
			}

			// delete the job
			l.V(1).Info("delete the Job", "jobName", jobName)
			if err := utils.DelteJob(r.Client, ctx, jobName, pgbench.Namespace); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to delete Job")
			}

			// update the pgbench status
			if err := r.Status().Update(ctx, &pgbench); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to update pgbench status")
			}
			return controllerutil.RequeueAfter(controllerutil.RequeueDuration)
		}

		// status is empty, job is creating
		return controllerutil.RequeueAfter(controllerutil.RequeueDuration)
	} else {
		// check the success number
		if pgbench.Status.Succeeded >= pgbench.Status.Total {
			if err := updatePgbenchStatus(r, ctx, &pgbench, benchmarkv1alpha1.Complete); err != nil {
				return controllerutil.RequeueWithError(err, l, "unable to update pgbench status")
			}
			return controllerutil.Reconciled()
		}

		l.Info("Job isn't existed", "jobName", jobName)

		// don't have job, and the pgbench is not complete
		// create a new job
		job := NewJob(&pgbench, jobName)
		l.Info("create a new Job", "jobName", job.Name)
		if err := r.Create(ctx, job); err != nil {
			return controllerutil.RequeueWithError(err, l, "unable to create Job")
		}
		return controllerutil.RequeueAfter(controllerutil.RequeueDuration)
	}
}

func updatePgbenchStatus(r *PgbenchReconciler, ctx context.Context, pgbench *benchmarkv1alpha1.Pgbench, status benchmarkv1alpha1.BenchmarkPhase) error {
	// The pgbench could have been modified since the last time we got it
	if err := r.Get(ctx, types.NamespacedName{Name: pgbench.Name, Namespace: pgbench.Namespace}, pgbench); err != nil {
		return err
	}
	pgbench.Status.Phase = status
	if err := r.Status().Update(ctx, pgbench); err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PgbenchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&benchmarkv1alpha1.Pgbench{}).
		Complete(r)
}

func ParsePgbench(msg string) string {
	result := ""
	lines := strings.Split(msg, "\n")
	index := len(lines)

	for i, l := range lines {
		if strings.Contains(l, "transaction type") {
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
