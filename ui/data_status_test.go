package ui

import (
	"os"
	"strings"
	"testing"
)

func TestSummarizeStructuredRun(t *testing.T) {
	result := StructuredRunResult{Status: "partial", Sections: []StructuredSection{
		{Name: "tcp", Enabled: true, Status: "unavailable", Reason: "network unavailable"},
		{Name: "basics", Enabled: true, Status: "ok"},
	}}
	result.SchemaVersion = structuredReportSchema
	status, reason := summarizeStructuredRun(result)
	if status.Source != "unavailable" || status.Fallback || !strings.Contains(reason, "tcp") {
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
	if report.SchemaVersion != structuredReportSchema || report.PrivacyMode != true || status.Source != "unavailable" || status.Fallback || !strings.Contains(reason, "tcp") || !strings.Contains(reason, "basics") {
		t.Fatalf("fixture was not consumed correctly: report=%#v status=%#v reason=%q", report, status, reason)
	}
}

func TestStructuredRunOmitsDataFileReason(t *testing.T) {
	result := StructuredRunResult{Status: "partial", DataFiles: []StructuredDataFile{
		{File: "private.json", Schema: "private/v1", Source: "https://private.example/list", Status: "timeout", Reason: "fetch failed"},
	}}
	status, reason := summarizeStructuredRun(result)
	if strings.Contains(reason, "private.json") || strings.Contains(reason, "private.example") || strings.Contains(reason, "data ") {
		t.Fatalf("data file details were exposed: %q", reason)
	}
	if status.Source != "unavailable" || status.Fallback {
		t.Fatalf("data status exposed provenance: %#v", status)
	}
}

func TestStructuredPublicViewDropsProvenanceAndRedactsPayload(t *testing.T) {
	report := StructuredRunResult{
		Data:       &StructuredDataVersion{Source: "https://private.example/list", File: "private.json"},
		DataFiles:  []StructuredDataFile{{File: "private.json", Source: "git@private.example:owner/repo.git"}},
		Sections:   []StructuredSection{{Name: "speed", Status: "error", Reason: "fetch https://private.example/list?token=secret failed"}},
		Components: []StructuredComponent{{Name: "speed.registry", Status: "error", Payload: []byte(`{"private_registry":{"url":"https://private.example/list"},"url":"https://private.example/node","value":1}`)}},
	}
	sanitizeStructuredRunResult(&report)
	encoded, err := structuredReportJSON(report)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"private.example", "owner/repo.git", "private.json", "private_registry", "data_files"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("GUI public view exposed %q: %s", forbidden, encoded)
		}
	}
}

func TestGUIPublicSanitizerPreservesBusinessTargets(t *testing.T) {
	payload := sanitizeGUIPayload([]byte(`{
		"providers":[{"source":"ipregistry.co","status":"available"}],
		"selected":[{"host":"speed.example:443","url":"https://speed.example/upload?token=secret&mode=test","availability":"available"}],
		"registry":{"source":"private-loader","fallback":"embedded"},
		"private_registry":{"url":"https://private.example/list"},
		"error":"fetch https://private.example/list?key=secret failed"
	}`))
	encoded := string(payload)
	for _, forbidden := range []string{"private.example", "private-loader", `"fallback"`, `"private_registry"`, "token=secret", "key=secret"} {
		if strings.Contains(encoded, forbidden) {
			t.Fatalf("GUI public payload exposed %q: %s", forbidden, encoded)
		}
	}
	for _, preserved := range []string{"ipregistry.co", "speed.example:443", "https://speed.example/upload", "mode=test"} {
		if !strings.Contains(encoded, preserved) {
			t.Fatalf("GUI public payload removed business field %q: %s", preserved, encoded)
		}
	}
}
