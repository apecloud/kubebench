package v1alpha1

import "fmt"

const (
	DefaultEsrallyTrack        = "geonames"
	DefaultEsrallyOnError      = "abort"
	DefaultEsrallyReportFormat = "csv"
	DefaultEsrallyReportFile   = "/var/log/esrally-report.csv"
)

func (spec EsrallySpec) EffectiveTrack() string {
	return stringOrDefault(spec.Track, DefaultEsrallyTrack)
}

func (spec EsrallySpec) EffectiveOnError() string {
	return stringOrDefault(spec.OnError, DefaultEsrallyOnError)
}

func (spec EsrallySpec) EffectiveReportFormat() string {
	return stringOrDefault(spec.ReportFormat, DefaultEsrallyReportFormat)
}

func (spec EsrallySpec) EffectiveReportFile() string {
	return stringOrDefault(spec.ReportFile, DefaultEsrallyReportFile)
}

func (spec EsrallySpec) MetricsRequested() bool {
	return spec.Metrics == nil || *spec.Metrics
}

func (spec EsrallySpec) EffectiveClientOptions() string {
	if spec.ClientOptions != "" {
		return spec.ClientOptions
	}
	if spec.Target.User == "" && spec.Target.Password == "" {
		return ""
	}
	return fmt.Sprintf("basic_auth_user:'%s',basic_auth_password:'%s'", spec.Target.User, spec.Target.Password)
}

func stringOrDefault(value string, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}
