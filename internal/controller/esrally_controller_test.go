package controller

import (
	"strings"
	"testing"
)

func TestParseEsrallyReturnsBoundedSummaryWithoutRawLogs(t *testing.T) {
	msg := `esrally race finished
basic_auth_password:'top-secret'
api_key:'abc123'`

	summary := ParseEsrally(msg)
	if !strings.Contains(summary, "No numeric Rally CSV summary was found.") {
		t.Fatalf("expected bounded fallback summary, got %q", summary)
	}
	if strings.Contains(summary, "top-secret") || strings.Contains(summary, "abc123") {
		t.Fatalf("summary leaked sensitive log content: %q", summary)
	}
}
