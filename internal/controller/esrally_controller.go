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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	benchmarkv1alpha1 "github.com/apecloud/kubebench/api/v1alpha1"
	intctrlutil "github.com/apecloud/kubebench/internal/controllerutil"
	"github.com/apecloud/kubebench/internal/exporter"
	"github.com/apecloud/kubebench/internal/utils"
)

// EsrallyReconciler reconciles an Esrally object.
type EsrallyReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RestConfig *rest.Config
}

//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=esrallies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=esrallies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=benchmark.apecloud.io,resources=esrallies/finalizers,verbs=update

func (r *EsrallyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var esrally benchmarkv1alpha1.Esrally
	if err := r.Get(ctx, req.NamespacedName, &esrally); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	old := esrally.DeepCopy()

	if esrally.Status.Phase == benchmarkv1alpha1.Completed || esrally.Status.Phase == benchmarkv1alpha1.Failed {
		return intctrlutil.Reconciled()
	}

	jobs := NewEsrallyJobs(&esrally)

	if esrally.Status.Phase == "" {
		l.Info("start esrally", "esrally", esrally.Name)
		esrally.Status.Phase = benchmarkv1alpha1.Running
		esrally.Status.Total = len(jobs)
	}

	if esrally.Status.Succeeded >= esrally.Status.Total {
		l.Info("esrally complete", "esrally", esrally.Name)
		esrally.Status.Phase = benchmarkv1alpha1.Completed
		esrally.Status.CompletionTimestamp = &metav1.Time{Time: time.Now()}
	} else {
		job := jobs[esrally.Status.Succeeded]

		existed, err := utils.IsJobExisted(r.Client, ctx, job.Name, esrally.Namespace)
		if err != nil {
			l.Error(err, "failed to check job existence", "job", job.Name)
			return intctrlutil.RequeueWithError(err, l, "failed to check job existence")
		}

		if !existed {
			if err = controllerutil.SetOwnerReference(&esrally, job, r.Scheme); err != nil {
				l.Error(err, "failed to set owner reference", "job", job.Name)
				return intctrlutil.RequeueWithError(err, l, "failed to set owner reference")
			}

			if err = r.Create(ctx, job); err != nil {
				l.Error(err, "failed to create job", "job", job.Name)
				return intctrlutil.RequeueWithError(err, l, "failed to create job")
			}

			l.Info("created job", "job", job.Name)
			return intctrlutil.RequeueAfter(intctrlutil.RequeueDuration)
		}

		status, err := utils.GetJobStatus(r.Client, ctx, job.Name, job.Namespace)
		if err != nil {
			l.Error(err, "failed to get job status", "job", job.Name)
			return intctrlutil.RequeueWithError(err, l, "failed to get job status")
		}

		if status.Succeeded > 0 {
			l.Info("job completed", "job", job.Name)
			esrally.Status.Succeeded++
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, job.Name, esrally.Namespace, &esrally.Status.Conditions, ParseEsrally); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to record the log")
			}
		} else if status.Failed > 0 {
			l.Info("job failed", "job", job.Name)
			esrally.Status.Phase = benchmarkv1alpha1.Failed
			if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, job.Name, esrally.Namespace, &esrally.Status.Conditions, nil); err != nil {
				return intctrlutil.RequeueWithError(err, l, "unable to record the log")
			}
		} else {
			l.Info("job running", "job", job.Name)
		}
	}

	esrally.Status.Completions = fmt.Sprintf("%d/%d", esrally.Status.Succeeded, esrally.Status.Total)
	if err := r.Status().Patch(ctx, &esrally, client.MergeFrom(old)); err != nil {
		l.Error(err, "failed to patch esrally status")
		return intctrlutil.RequeueWithError(err, l, "failed to patch esrally status")
	}

	return intctrlutil.RequeueAfter(intctrlutil.RequeueDuration)
}

func (r *EsrallyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&benchmarkv1alpha1.Esrally{}).
		Complete(r)
}

func ParseEsrally(msg string) string {
	return exporter.SummarizeEsrallyCSV(msg, 12)
}
