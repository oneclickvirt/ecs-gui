package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const (
	structuredReportSchema = "goecs.report/v1"
	maxStructuredJSONBytes = 16 << 20
)

func summarizeStructuredRun(result StructuredRunResult) string {
	sanitizeStructuredRunResult(&result)
	var reasons []string
	if summary := structuredDNSResolutionSummary(result.DNS); summary != "" {
		reasons = append(reasons, summary)
	}
	for _, section := range result.Sections {
		if section.Status == "ok" || section.Status == "skipped" {
			continue
		}
		reasons = append(reasons, structuredReason(section.Name, section.Status, section.Reason))
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
	return strings.Join(uniqueStrings(reasons), "; ")
}

func structuredDNSResolutionSummary(dns *StructuredDNSResolution) string {
	if dns == nil {
		return ""
	}
	provider := strings.TrimSpace(dns.Provider)
	suffix := ""
	if provider != "" {
		suffix = " (" + provider + ")"
	}
	switch strings.ToLower(strings.TrimSpace(dns.Active)) {
	case "system":
		if strings.TrimSpace(dns.Reason) != "" {
			return "DNS: system resolver retained after inconclusive probe"
		}
		return "DNS: system resolver"
	case "doh":
		if dns.Fallback {
			return "DNS: built-in DoH fallback" + suffix
		}
		return "DNS: built-in DoH" + suffix
	case "dot":
		if dns.Fallback {
			return "DNS: built-in DoT fallback" + suffix
		}
		return "DNS: built-in DoT" + suffix
	case "", "unavailable":
		return "DNS: unavailable"
	default:
		return "DNS: unavailable"
	}
}

func structuredReason(name, status, reason string) string {
	reason = safeStructuredReason(status, reason)
	if reason == "" {
		reason = safeStatusReason(status)
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
	sanitizeStructuredRunResult(&result)
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
