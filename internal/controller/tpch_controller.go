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

// TpchReconciler reconciles a Tpch object
type TpchReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RestConfig *rest.Config
}

//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=tpches,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=tpches/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=tpches/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Tpch object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *TpchReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	// TODO(user): your logic here
	var tpch benchmarkv1alpha1.Tpch
	var err error
	if err = r.Get(ctx, req.NamespacedName, &tpch); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Run to one completion
	if tpch.Status.Phase == benchmarkv1alpha1.Complete ||
		tpch.Status.Phase == benchmarkv1alpha1.Failed ||
		tpch.Status.Phase == benchmarkv1alpha1.Running {
		return intctrlutil.Reconciled()
	}

	jobs := NewTpchJobs(&tpch)

	tpch.Status.Phase = benchmarkv1alpha1.Running
	tpch.Status.Total = len(jobs)
	tpch.Status.Completions = fmt.Sprintf("%d/%d", tpch.Status.Succeeded, tpch.Status.Total)

	if err := r.Status().Update(ctx, &tpch); err != nil {
		return intctrlutil.RequeueWithError(err, l, "unable to update tpch status")
	}

	go func() {
		for _, job := range jobs {
			if err = controllerutil.SetOwnerReference(&tpch, job, r.Scheme); err != nil {
				l.Info("unable to set owner reference", "job", job.Name)
				return
			}

			if err = r.Create(ctx, job); err != nil {
				l.Info("unable to create job", "job", job.Name)
				return
			}

			l.Info("created job", "job", job.Name)
			// wait for the job to complete, then update the tpch status

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
					tpch.Status.Phase = benchmarkv1alpha1.Failed
				}

				// job is completed
				if status.Succeeded > 0 {
					l.Info("job is succeeded", "jobName", job.Name)
					tpch.Status.Succeeded += 1
					tpch.Status.Completions = fmt.Sprintf("%d/%d", tpch.Status.Succeeded, tpch.Status.Total)
				}

				// record the result
				if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, job.Name, tpch.Namespace, &tpch.Status.Conditions, ParseTpch); err != nil {
					l.Info("unable to record the log", "job", job.Name)
				}

				break
			}

			// update the tpch status
			if err := r.Status().Update(ctx, &tpch); err != nil {
				l.Info("unable to update tpch status", "job", job.Name)
			}

			// if the tpch is failed, return
			if tpch.Status.Phase == benchmarkv1alpha1.Failed {
				return
			}
		}

		tpch.Status.Phase = benchmarkv1alpha1.Complete
		r.Status().Update(ctx, &tpch)
	}()

	return intctrlutil.Reconciled()
}

func ParseTpch(msg string) string {
	result := ""
	lines := strings.Split(msg, "\n")
	index := len(lines)

	for i, l := range lines {
		if strings.Contains(l, "run tpch query") {
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

// SetupWithManager sets up the controller with the Manager.
func (r *TpchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&benchmarkv1alpha1.Tpch{}).
		Complete(r)
}
