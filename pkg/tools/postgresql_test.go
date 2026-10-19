package tools

import (
	"testing"

	"github.com/lib/pq"
)

func TestMaintenanceCandidatesSkipTargetAndTemplate1(t *testing.T) {
	got := maintenanceCandidates("test", "", false)
	for _, name := range got {
		if name == "test" || name == "template1" {
			t.Fatalf("candidate %s must not be used to create itself or as template1", name)
		}
	}
	if len(got) == 0 || got[0] != "postgres" {
		t.Fatalf("postgres should be the first fallback, got %#v", got)
	}

	if got := maintenanceCandidates("mydb", "postgres", true); len(got) != 1 || got[0] != "postgres" {
		t.Fatalf("explicit maintenance database not honored: %#v", got)
	}
	if got := maintenanceCandidates("mydb", "template1", true); len(got) == 1 && got[0] == "template1" {
		t.Fatalf("explicit template1 must not be the maintenance connection: %#v", got)
	}
	if got := maintenanceCandidates("mydb", "mydb", true); len(got) == 1 && got[0] == "mydb" {
		t.Fatalf("must not connect to the database being created: %#v", got)
	}
}

func TestIsMissingDatabase(t *testing.T) {
	if !isMissingDatabase(&pq.Error{Code: "3D000"}) {
		t.Fatal("expected invalid catalog name to be missing")
	}
	if isMissingDatabase(&pq.Error{Code: "55006", Message: `database "test" is being accessed by other users`}) {
		t.Fatal("in-use database is not missing")
	}
}

func TestSafeIdent(t *testing.T) {
	if err := safeIdent("usertable"); err != nil {
		t.Fatal(err)
	}
	if err := safeIdent("test;drop"); err == nil {
		t.Fatal("expected invalid identifier")
	}
}
