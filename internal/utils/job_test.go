package utils

import (
	"strings"
	"testing"

	"github.com/apecloud/kubebench/api/v1alpha1"
)

func TestNewPgbenchPreCheckJobUsesTargetDatabase(t *testing.T) {
	target := v1alpha1.Target{
		Host:     "pg.default.svc",
		Port:     5432,
		User:     "bench",
		Password: "secret",
		Database: "customer_db",
	}

	job := NewPgbenchPreCheckJob("bench", "default", target)
	args := strings.Join(job.Spec.Template.Spec.Containers[0].Args, " ")
	for _, want := range []string{"postgresql", "ping", "--database", target.Database} {
		if !strings.Contains(args, want) {
			t.Fatalf("expected precheck args to contain %q, got %q", want, args)
		}
	}
}
