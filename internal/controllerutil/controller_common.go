package controllerutil

import (
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	RequeueDuration = time.Second
)

// Reconciled returns an empty result with nil error to signal a successful reconcile
// to the controller manager
func Reconciled() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

// RequeueAfter returns an empty result with nil error to signal a successful reconcile
// to the controller manager, but requests that the reconcile be run again after the
// given duration
func RequeueAfter(duration time.Duration) (reconcile.Result, error) {
	return reconcile.Result{
		Requeue:      true,
		RequeueAfter: duration,
	}, nil
}

// RequeueWithError returns an empty result with the given error to signal a failed
func RequeueWithError(err error, logger logr.Logger, msg string, keysAndValues ...interface{}) (reconcile.Result, error) {
	if msg == "" {
		logger.Info(err.Error())
	} else {
		// Info log the error message and then let the reconciler dump the stacktrace
		logger.Info(msg, keysAndValues...)
	}
	return reconcile.Result{}, err
}
