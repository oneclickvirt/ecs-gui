package ui

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestFormatStructuredDetailsUsesCompactLocalizedOverview(t *testing.T) {
	details := formatStructuredDetails(StructuredRunResult{
		SchemaVersion: "goecs.report/v1",
		Status:        "partial",
		Sections:      []StructuredSection{{Name: "basics", Enabled: true, Status: "ok"}},
		Components: []StructuredComponent{{
			Name: "basics", SchemaVersion: "goecs.system/v1", Status: "partial",
			Reason: "temperature unavailable", DurationMS: 10,
			Payload: []byte(`{"availability":"partial","devices":["nvme0"]}`),
		}},
	}, langZH)
	for _, want := range []string{"运行概览", "部分完成", "章节进度", "组件结果", "系统基础信息", "tempera..."} {
		if !strings.Contains(details, want) {
			t.Fatalf("details missing %q: %s", want, details)
		}
	}
	for _, forbidden := range []string{"schema=", "schema_version", "payload:", "\"devices\"", "{"} {
		if strings.Contains(details, forbidden) {
			t.Fatalf("details exposed machine field %q: %s", forbidden, details)
		}
	}
	for _, line := range strings.Split(details, "\n") {
		if width := runewidth.StringWidth(line); width > structuredOverviewWidth {
			t.Fatalf("overview line width %d exceeds %d: %q", width, structuredOverviewWidth, line)
		}
	}
}

func TestFormatStructuredDetailsSupportsEnglish(t *testing.T) {
	details := formatStructuredDetails(StructuredRunResult{
		Status: "ok", DurationMS: 1200,
		Sections: []StructuredSection{{Name: "speed", Enabled: true, Status: "ok"}},
	}, langEN)
	for _, want := range []string{"Run Overview", "Overall Status", "Section Progress", "Speed"} {
		if !strings.Contains(details, want) {
			t.Fatalf("English details missing %q: %s", want, details)
		}
	}
}
