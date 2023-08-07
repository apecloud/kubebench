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
	var err error
	if err = r.Get(ctx, req.NamespacedName, &pgbench); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if pgbench.Status.Phase == benchmarkv1alpha1.Complete || pgbench.Status.Phase == benchmarkv1alpha1.Failed {
		return intctrlutil.Reconciled()
	}

	jobs := NewPgbenchJobs(&pgbench)

	pgbench.Status.Phase = benchmarkv1alpha1.Running
	pgbench.Status.Total = len(jobs)
	pgbench.Status.Completions = fmt.Sprintf("%d/%d", pgbench.Status.Succeeded, pgbench.Status.Total)

	if err = r.Status().Update(ctx, &pgbench); err != nil {
		return intctrlutil.RequeueWithError(err, l, "update to update pgbench status")
	}

	for _, job := range jobs {
		if err = controllerutil.SetOwnerReference(&pgbench, job, r.Scheme); err != nil {
			return intctrlutil.RequeueWithError(err, l, "failed to set owner reference for job")
		}

		if err = r.Create(ctx, job); err != nil {
			return intctrlutil.RequeueWithError(err, l, "failed to create job")
		}

		l.Info("created job", "job", job.Name)
		// wait for the job to complete, then update the pgbench status

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
				pgbench.Status.Phase = benchmarkv1alpha1.Failed
			}

			// job is completed
			if status.Succeeded > 0 {
				l.Info("job is succeeded", "jobName", job.Name)
				pgbench.Status.Succeeded += 1
				pgbench.Status.Completions = fmt.Sprintf("%d/%d", pgbench.Status.Succeeded, pgbench.Status.Total)
			}

			// record the result
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, job.Name, pgbench.Namespace, &pgbench.Status.Conditions, ParsePgbench); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to record the log")
			}

			break
		}

		// update the pgbench status
		if err := r.Status().Update(ctx, &pgbench); err != nil {
			return intctrlutil.RequeueWithError(err, l, "unable to update pgbench status")
		}

		if err != nil {
			return intctrlutil.RequeueWithError(err, l, "")
		}
	}

	pgbench.Status.Phase = benchmarkv1alpha1.Complete
	r.Status().Update(ctx, &pgbench)
	return intctrlutil.Reconciled()
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
