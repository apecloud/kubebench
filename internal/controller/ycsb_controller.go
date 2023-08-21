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

// YcsbReconciler reconciles a Ycsb object
type YcsbReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RestConfig *rest.Config
}

var ycsbToJobController = make(map[string]JobsController)

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

	jobs := NewYcsbJobs(&ycsb)

	if ycsb.Status.Phase == "" {
		l.Info("start ycsb", "ycsb", ycsb.Name)
		ycsb.Status.Phase = benchmarkv1alpha1.Running
		ycsb.Status.Total = len(jobs)
	}

	if _, ok := ycsbToJobController[ycsb.Name]; !ok {
		ycsbToJobController[ycsb.Name] = NewJobsController(r.Client, jobs)
	}
	jc := ycsbToJobController[ycsb.Name]

	if jc.Completed() {
		ycsb.Status.Phase = benchmarkv1alpha1.Complete
	} else {
		job := jc.GetCurJob()

		existed, err := utils.IsJobExisted(r.Client, ctx, job.Name, ycsb.Namespace)
		if err != nil {
			return intctrlutil.RequeueWithError(err, l, "failed to check if job exists")
		}

		if !existed {
			if err = controllerutil.SetOwnerReference(&ycsb, job, r.Scheme); err != nil {
				return intctrlutil.RequeueWithError(err, l, "failed to set owner reference for job")
			}

			if err := jc.StartJob(); err != nil {
				l.Error(err, "failed to start job")
				return intctrlutil.RequeueWithError(err, l, "failed to start job")
			}

			// wait for the job to be created
			l.Info("created job", "job", job.Name)
			return intctrlutil.RequeueAfter(intctrlutil.RequeueDuration)
		}

		if status, err := jc.CurJobStatus(); err != nil {
			return intctrlutil.RequeueWithError(err, l, "failed to get job status")
		} else {
			switch status {
			case Complete:
				ycsb.Status.Succeeded++
				// record the result
				if err := utils.LogJobPodToCond(r.Client, r.RestConfig, ctx, job.Name, ycsb.Namespace, &ycsb.Status.Conditions, ParseYcsb); err != nil {
					return intctrlutil.RequeueWithError(err, l, "unable to record the log")
				}

				jc.NextJob()
			case Failed:
				ycsb.Status.Phase = benchmarkv1alpha1.Failed
			}
		}
	}

	ycsb.Status.Completions = fmt.Sprintf("%d/%d", ycsb.Status.Succeeded, ycsb.Status.Total)
	if err := r.Status().Update(ctx, &ycsb); err != nil {
		return intctrlutil.RequeueWithError(err, l, "unable to update ycsb status")
	}

	return intctrlutil.RequeueAfter(intctrlutil.RequeueDuration)
}

// SetupWithManager sets up the controller with the Manager.
func (r *YcsbReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&benchmarkv1alpha1.Ycsb{}).
		Complete(r)
}

func ParseYcsb(msg string) string {
	result := ""
	lines := strings.Split(msg, "\n")
	index := len(lines)

	for i, l := range lines {
		if strings.Contains(l, "Run finished, takes") {
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
