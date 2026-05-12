package v1alpha1

import "testing"

func TestEsrallySpecEffectiveDefaults(t *testing.T) {
	spec := EsrallySpec{}

	if got := spec.EffectiveTrack(); got != DefaultEsrallyTrack {
		t.Fatalf("expected default track %q, got %q", DefaultEsrallyTrack, got)
	}
	if got := spec.EffectiveOnError(); got != DefaultEsrallyOnError {
		t.Fatalf("expected default onError %q, got %q", DefaultEsrallyOnError, got)
	}
	if got := spec.EffectiveReportFormat(); got != DefaultEsrallyReportFormat {
		t.Fatalf("expected default reportFormat %q, got %q", DefaultEsrallyReportFormat, got)
	}
	if got := spec.EffectiveReportFile(); got != DefaultEsrallyReportFile {
		t.Fatalf("expected default reportFile %q, got %q", DefaultEsrallyReportFile, got)
	}
	if !spec.MetricsRequested() {
		t.Fatal("expected metrics to default to requested for direct Go objects")
	}
}

func TestEsrallySpecMetricsRequestedRespectsExplicitValues(t *testing.T) {
	falseValue := false
	trueValue := true

	tests := []struct {
		name    string
		metrics *bool
		want    bool
	}{
		{name: "omitted", metrics: nil, want: true},
		{name: "explicit false", metrics: &falseValue, want: false},
		{name: "explicit true", metrics: &trueValue, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := EsrallySpec{Metrics: tt.metrics}
			if got := spec.MetricsRequested(); got != tt.want {
				t.Fatalf("expected metrics requested=%t, got %t", tt.want, got)
			}
		})
	}
}

func TestEsrallySpecEffectiveClientOptions(t *testing.T) {
	spec := EsrallySpec{}
	spec.Target.User = "elastic"
	spec.Target.Password = "secret"

	if got := spec.EffectiveClientOptions(); got != "basic_auth_user:'elastic',basic_auth_password:'secret'" {
		t.Fatalf("unexpected synthesized client options: %q", got)
	}

	spec.ClientOptions = "use_ssl:true,verify_certs:false"
	if got := spec.EffectiveClientOptions(); got != spec.ClientOptions {
		t.Fatalf("expected explicit client options to pass through, got %q", got)
	}
}
