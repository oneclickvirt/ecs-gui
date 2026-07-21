package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	structuredReportSchema = "goecs.report/v1"
	maxStructuredJSONBytes = 16 << 20
)

type dataStatus struct {
	Schema      string
	GeneratedAt time.Time
	Source      string
	Fallback    bool
	FallbackTo  string
	File        string
	Count       int
	Error       string
}

func (ui *TestUI) updateDataStatus(status dataStatus) {
	if ui.DataStatusLabel == nil {
		return
	}
	if status.Source == "unavailable" {
		ui.DataStatusLabel.SetText(ui.tr("data.unavailable"))
		return
	}
	if status.GeneratedAt.IsZero() && strings.EqualFold(status.Source, "embedded") {
		ui.DataStatusLabel.SetText(ui.tr("data.embedded"))
		return
	}
	value := fmt.Sprintf(ui.tr("data.version"), status.GeneratedAt.Local().Format("2006-01-02 15:04"), displayDataSource(status.Source))
	if status.Fallback {
		value += " " + ui.tr("data.fallback")
	}
	ui.DataStatusLabel.SetText(value)
}

func displayDataSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "cdn":
		return "CDN"
	case "raw", "github raw":
		return "GitHub Raw"
	case "embedded":
		return "embedded"
	default:
		return source
	}
}

func summarizeStructuredRun(result StructuredRunResult) (dataStatus, string) {
	status := dataStatus{Source: "unavailable"}
	if len(result.DataFiles) > 0 {
		status.Source = "components"
		status.Schema = "component-registries/v1"
		for _, file := range result.DataFiles {
			if file.Status != "ok" {
				continue
			}
			status.Count += file.Count
			if file.GeneratedAt.After(status.GeneratedAt) {
				status.GeneratedAt = file.GeneratedAt
			}
			if file.Fallback != "" {
				status.Fallback = true
				status.FallbackTo = file.Fallback
			}
		}
	} else if result.Data != nil {
		status.Schema = result.Data.Schema
		status.GeneratedAt = result.Data.GeneratedAt
		status.Source = result.Data.Source
		status.FallbackTo = result.Data.Fallback
		status.Fallback = result.Data.Fallback != ""
		status.File = result.Data.File
		status.Count = result.Data.Count
	}
	var reasons []string
	for _, section := range result.Sections {
		if section.Status == "ok" || section.Status == "skipped" {
			continue
		}
		reasons = append(reasons, structuredReason(section.Name, section.Status, section.Reason))
	}
	for _, file := range result.DataFiles {
		if file.Status == "ok" || file.Status == "skipped" {
			continue
		}
		reasons = append(reasons, structuredReason("data "+file.File, file.Status, file.Reason))
	}
	for _, component := range result.Components {
		if component.Status == "ok" || component.Status == "skipped" {
			continue
		}
		reasons = append(reasons, structuredReason(component.Name, component.Status, component.Reason))
	}
	if len(reasons) == 0 && result.Status != "" && result.Status != "ok" {
		reasons = append(reasons, result.Status)
	}
	return status, strings.Join(uniqueStrings(reasons), "; ")
}

func structuredReason(name, status, reason string) string {
	if reason == "" {
		reason = status
	}
	if name == "" {
		return reason
	}
	return fmt.Sprintf("%s: %s", name, reason)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func decodeStructuredRun(data []byte) (StructuredRunResult, error) {
	if len(data) == 0 {
		return StructuredRunResult{}, fmt.Errorf("empty structured report")
	}
	if len(data) > maxStructuredJSONBytes {
		return StructuredRunResult{}, fmt.Errorf("structured report exceeds %d bytes", maxStructuredJSONBytes)
	}
	var result StructuredRunResult
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	if err := decoder.Decode(&result); err != nil {
		return StructuredRunResult{}, fmt.Errorf("decode structured report: %w", err)
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return StructuredRunResult{}, fmt.Errorf("structured report contains trailing data")
		}
		return StructuredRunResult{}, fmt.Errorf("structured report trailing data: %w", err)
	}
	if result.SchemaVersion != structuredReportSchema {
		return StructuredRunResult{}, fmt.Errorf("unsupported schema %q", result.SchemaVersion)
	}
	if !validReportStatus(result.Status) {
		return StructuredRunResult{}, fmt.Errorf("invalid report status %q", result.Status)
	}
	if result.DurationMS < 0 {
		return StructuredRunResult{}, fmt.Errorf("invalid negative duration")
	}
	for _, section := range result.Sections {
		if strings.TrimSpace(section.Name) == "" || !validReportStatus(section.Status) {
			return StructuredRunResult{}, fmt.Errorf("invalid section status for %q", section.Name)
		}
	}
	for _, file := range result.DataFiles {
		if strings.TrimSpace(file.File) == "" || !validReportStatus(file.Status) || file.Count < 0 {
			return StructuredRunResult{}, fmt.Errorf("invalid data file %q", file.File)
		}
		if file.Status == "ok" && (!supportedDataSchema(file.Schema) || file.GeneratedAt.IsZero() || strings.TrimSpace(file.Source) == "") {
			return StructuredRunResult{}, fmt.Errorf("incomplete data file %q", file.File)
		}
	}
	for _, component := range result.Components {
		if strings.TrimSpace(component.Name) == "" || component.SchemaVersion == "" || !validReportStatus(component.Status) {
			return StructuredRunResult{}, fmt.Errorf("invalid component %q", component.Name)
		}
	}
	return result, nil
}

func supportedDataSchema(schema string) bool {
	schema = strings.TrimSpace(schema)
	return schema != "" && strings.Contains(schema, "/")
}

func validReportStatus(status string) bool {
	switch status {
	case "ok", "partial", "unavailable", "timeout", "canceled", "error", "skipped":
		return true
	default:
		return false
	}
}

// extractStructuredRun accepts either a JSON report or a mixed terminal
// capture containing a fenced/inline goecs.report/v1 JSON object.
func extractStructuredRun(data []byte) (StructuredRunResult, error) {
	if result, err := decodeStructuredRun(data); err == nil {
		return result, nil
	}
	text := string(data)
	marker := strings.Index(text, `"schema_version"`)
	if marker < 0 {
		return StructuredRunResult{}, fmt.Errorf("structured report schema marker not found")
	}
	start := strings.LastIndex(text[:marker], "{")
	if start < 0 {
		return StructuredRunResult{}, fmt.Errorf("structured report object start not found")
	}
	decoder := json.NewDecoder(strings.NewReader(text[start:]))
	var result StructuredRunResult
	if err := decoder.Decode(&result); err != nil {
		return StructuredRunResult{}, fmt.Errorf("decode embedded structured report: %w", err)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return StructuredRunResult{}, err
	}
	return decodeStructuredRun(encoded)
}
