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

// RedisbenchReconciler reconciles a Redisbench object
type RedisbenchReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RestConfig *rest.Config
}

//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=redisbenches,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=redisbenches/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=redisbenches/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Redisbench object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *RedisbenchReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var redisbench benchmarkv1alpha1.Redisbench
	if err := r.Get(ctx, req.NamespacedName, &redisbench); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	old := redisbench.DeepCopy()

	if redisbench.Status.Phase == benchmarkv1alpha1.Complete || redisbench.Status.Phase == benchmarkv1alpha1.Failed {
		return intctrlutil.Reconciled()
	}

	jobs := NewRedisBenchJobs(&redisbench)

	if redisbench.Status.Phase == "" {
		l.Info("start redisbench", "redisbench", redisbench.Name)
		redisbench.Status.Phase = benchmarkv1alpha1.Running
		redisbench.Status.Total = len(jobs)
	}

	if redisbench.Status.Succeeded >= redisbench.Status.Total {
		l.Info("redisbench complete", "redisbench", redisbench.Name)
		redisbench.Status.Phase = benchmarkv1alpha1.Complete
	} else {
		job := jobs[redisbench.Status.Succeeded]

		existed, err := utils.IsJobExisted(r.Client, ctx, job.Name, redisbench.Namespace)
		if err != nil {
			l.Error(err, "failed to check job existence", "job", job.Name)
			return intctrlutil.RequeueWithError(err, l, "failed to check job existence")
		}

		if !existed {
			if err = controllerutil.SetOwnerReference(&redisbench, job, r.Scheme); err != nil {
				l.Error(err, "failed to set owner reference", "job", job.Name)
				return intctrlutil.RequeueWithError(err, l, "failed to set owner reference")
			}

			if err = r.Create(ctx, job); err != nil {
				l.Error(err, "failed to create job", "job", job.Name)
				return intctrlutil.RequeueWithError(err, l, "failed to create job")
			}

			// wait for the job to be created
			l.Info("created job", "job", job.Name)
			return intctrlutil.RequeueAfter(intctrlutil.RequeueDuration)
		}

		// check if the job is completed
		status, err := utils.GetJobStatus(r.Client, ctx, job.Name, job.Namespace)
		if err != nil {
			l.Error(err, "failed to get job status", "job", job.Name)
			return intctrlutil.RequeueWithError(err, l, "failed to get job status")
		}

		if status.Succeeded > 0 {
			l.Info("job completed", "job", job.Name)
			redisbench.Status.Succeeded++
			// record the result
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, job.Name, redisbench.Namespace, &redisbench.Status.Conditions, ParseRedisBench); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to record the log")
			}
		} else if status.Failed > 0 {
			l.Info("job failed", "job", job.Name)
			redisbench.Status.Phase = benchmarkv1alpha1.Failed
		} else {
			l.Info("job running", "job", job.Name)
		}
	}

	redisbench.Status.Completions = fmt.Sprintf("%d/%d", redisbench.Status.Succeeded, redisbench.Status.Total)
	if err := r.Status().Patch(ctx, &redisbench, client.MergeFrom(old)); err != nil {
		l.Error(err, "failed to patch redisbench status")
		return intctrlutil.RequeueWithError(err, l, "failed to patch redisbench status")
	}

	return intctrlutil.RequeueAfter(intctrlutil.RequeueDuration)
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisbenchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&benchmarkv1alpha1.Redisbench{}).
		Complete(r)
}

func ParseRedisBench(msg string) string {
	result := ""

	lines := strings.Split(msg, "\r")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// save the result query/sev value
		if strings.Contains(line, "per second") {
			result += fmt.Sprintf("%s\n", line)
		}
	}

	return strings.TrimSpace(result)
}
