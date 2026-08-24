package controller

import (
	"testing"

	benchmarkv1alpha1 "github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/pkg/constants"
)

func TestNewPgbenchJobsUseTargetDatabaseWithoutInitContainer(t *testing.T) {
	cr := &benchmarkv1alpha1.Pgbench{}
	cr.Name = "pgbench"
	cr.Namespace = "default"
	cr.Spec.Target.Driver = constants.PostgreSqlDriver
	cr.Spec.Target.Host = "pg.default.svc"
	cr.Spec.Target.Port = 5432
	cr.Spec.Target.User = "bench"
	cr.Spec.Target.Password = "secret"
	cr.Spec.Target.Database = "customer_db"

	jobs := NewPgbenchJobs(cr)
	if len(jobs) == 0 {
		t.Fatal("expected pgbench jobs")
	}
	for _, job := range jobs {
		if len(job.Spec.Template.Spec.InitContainers) != 0 {
			t.Fatalf("expected no init containers, got %#v", job.Spec.Template.Spec.InitContainers)
		}
	}

	precheckArgs := jobs[0].Spec.Template.Spec.Containers[0].Args
	if !containsAll(precheckArgs, []string{"postgresql", "ping", "--database", cr.Spec.Target.Database}) {
		t.Fatalf("expected precheck to use target database, got %#v", precheckArgs)
	}
}
