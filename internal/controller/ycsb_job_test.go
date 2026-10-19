package controller

import (
	"testing"

	benchmarkv1alpha1 "github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/pkg/constants"
)

func TestYcsbPGCleanupDropsTableNotDatabase(t *testing.T) {
	for _, driver := range []string{constants.PostgreSqlDriver, constants.KingbaseDriver, constants.VastbaseDriver} {
		cr := &benchmarkv1alpha1.Ycsb{}
		cr.Name = "ycsb"
		cr.Namespace = "default"
		cr.Spec.Step = constants.AllStep
		cr.Spec.Threads = []int{1}
		cr.Spec.Target.Driver = driver
		cr.Spec.Target.Host = "db.default.svc"
		cr.Spec.Target.Port = 54321
		cr.Spec.Target.User = "test"
		cr.Spec.Target.Password = "secret"
		cr.Spec.Target.Database = "test"
		cr.Spec.ExtraArgs = []string{"-p table=bench"}

		var cleanupArgs, createArgs []string
		for _, job := range NewYcsbJobs(cr) {
			if len(job.Spec.Template.Spec.Containers) > 0 {
				args := job.Spec.Template.Spec.Containers[0].Args
				if len(args) >= 2 && args[0] == "postgresql" && args[1] == "drop-table" {
					cleanupArgs = args
				}
			}
			if len(job.Spec.Template.Spec.InitContainers) > 0 {
				createArgs = job.Spec.Template.Spec.InitContainers[0].Args
			}
		}
		if !containsAll(cleanupArgs, []string{"postgresql", "drop-table", "bench", "--database", "test"}) {
			t.Fatalf("%s cleanup args: %#v", driver, cleanupArgs)
		}
		for _, arg := range cleanupArgs {
			if arg == "drop" {
				t.Fatalf("%s still drops the database: %#v", driver, cleanupArgs)
			}
		}
		if containsAll(createArgs, []string{"--database", "test"}) {
			t.Fatalf("%s create connects to the target database: %#v", driver, createArgs)
		}
	}
}

func TestYcsbTableDefault(t *testing.T) {
	if got := ycsbTable(nil); got != "usertable" {
		t.Fatalf("got %s", got)
	}
	if got := ycsbTable([]string{""}); got != "usertable" {
		t.Fatalf("got %s", got)
	}
}
