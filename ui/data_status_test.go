package ui

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestSummarizeStructuredRun(t *testing.T) {
	result := StructuredRunResult{Status: "partial", Sections: []StructuredSection{
		{Name: "tcp", Enabled: true, Status: "unavailable", Reason: "network unavailable"},
		{Name: "basics", Enabled: true, Status: "ok"},
	}}
	result.Data = &StructuredDataVersion{
		Schema: "pingtest.tcp-targets/v1", GeneratedAt: time.Date(2026, 7, 19, 0, 0, 0, 0, time.UTC),
		Source: "raw", Fallback: "raw", File: "tcp-targets.json", Count: 1,
	}
	result.SchemaVersion = structuredReportSchema
	status, reason := summarizeStructuredRun(result)
	if status.Source != "raw" || !status.Fallback || !strings.Contains(reason, "tcp") {
		t.Fatalf("status=%+v reason=%q", status, reason)
	}
}

func TestDecodeStructuredRunV1AndRejectsTrailingData(t *testing.T) {
	report := []byte(`{"schema_version":"goecs.report/v1","ecs_version":"v0.1.139","status":"partial","duration_ms":12,"data":{"schema":"ecs-data/v1","generated_at":"2026-07-19T00:00:00Z","source":"raw","fallback":"raw","file":"tcp-targets.json","count":2},"sections":[{"name":"basics","enabled":true,"status":"ok"},{"name":"tcp","enabled":true,"status":"unavailable","reason":"network unavailable"}]}`)
	decoded, err := decodeStructuredRun(report)
	if err != nil || decoded.SchemaVersion != structuredReportSchema || len(decoded.Sections) != 2 {
		t.Fatalf("decode failed: %#v %v", decoded, err)
	}
	if _, err := decodeStructuredRun(append(report, []byte(` {}`)...)); err == nil {
		t.Fatal("expected trailing JSON to be rejected")
	}
}

func TestExtractStructuredRunFromMixedOutput(t *testing.T) {
	data := []byte("plain output\n{" + `"schema_version":"goecs.report/v1","status":"ok","sections":[]` + "}\n")
	decoded, err := extractStructuredRun(data)
	if err != nil || decoded.Status != "ok" {
		t.Fatalf("extract failed: %#v %v", decoded, err)
	}
}

func TestDecodeStructuredReportOfflineFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/goecs_report_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	report, err := decodeStructuredRun(data)
	if err != nil {
		t.Fatal(err)
	}
	status, reason := summarizeStructuredRun(report)
	if report.SchemaVersion != structuredReportSchema || report.PrivacyMode != true || status.Source != "components" || !status.Fallback || !strings.Contains(reason, "tcp") || !strings.Contains(reason, "basics") {
		t.Fatalf("fixture was not consumed correctly: report=%#v status=%#v reason=%q", report, status, reason)
	}
}

func TestStructuredRunIncludesDataFileReason(t *testing.T) {
	older := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	newer := older.Add(24 * time.Hour)
	result := StructuredRunResult{Status: "partial", DataFiles: []StructuredDataFile{
		{File: "tcp-targets.json", Schema: "pingtest.tcp-targets/v1", GeneratedAt: older, Source: "embedded", Count: 64, Status: "ok"},
		{File: "speedtest-servers.json", Schema: "speedtest.servers/v1", GeneratedAt: newer, Source: "raw", Fallback: "raw", Count: 278, Status: "ok"},
		{File: "dnsbl-zones.json", Count: 0, Status: "timeout", Reason: "deadline exceeded"},
	}}
	status, reason := summarizeStructuredRun(result)
	if !strings.Contains(reason, "data dnsbl-zones.json: deadline exceeded") {
		t.Fatalf("missing data file reason: %q", reason)
	}
	if status.Source != "components" || !status.GeneratedAt.Equal(newer) || !status.Fallback || status.Count != 342 {
		t.Fatalf("component data summary used the wrong file metadata: %#v", status)
	}
}
