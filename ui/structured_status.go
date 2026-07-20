package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var sectionProgressKeys = map[string]string{
	"basics":        "progress.basic_security",
	"cpu":           "progress.cpu",
	"memory":        "progress.memory",
	"disk":          "progress.disk",
	"deep_hardware": "progress.deep_hardware",
	"media":         "progress.unlock",
	"security":      "progress.ip_quality",
	"email":         "progress.email",
	"backtrace":     "progress.backtrace",
	"routes":        "progress.nt3",
	"ping":          "progress.ping",
	"tgdc":          "progress.tgdc",
	"web":           "progress.web",
	"speed":         "progress.speed",
	"nat":           "progress.nat",
	"tcp":           "progress.tcp",
	"analysis":      "progress.summary",
	"upload":        "progress.upload",
}

// ApplyStructuredReport consumes the goecs.report/v1 envelope without
// requiring the GUI to import a newer, unreleased goecs module.
func (ui *TestUI) ApplyStructuredReport(report StructuredRunResult) {
	ui.Mu.Lock()
	copy := report
	copy.Sections = append([]StructuredSection(nil), report.Sections...)
	copy.DataFiles = append([]StructuredDataFile(nil), report.DataFiles...)
	copy.Components = append([]StructuredComponent(nil), report.Components...)
	ui.StructuredResult = &copy
	ui.Mu.Unlock()
	ui.updateStructuredDetails(report)
	dataState, reason := summarizeStructuredRun(report)
	ui.updateDataStatus(dataState)
	ui.updatePartialReason(reason)
	ui.updateStructuredProgress(report)
	if key := structuredStatusKey(report.Status); key != "" {
		ui.setStatus(key)
	}
}

func (ui *TestUI) updateStructuredDetails(report StructuredRunResult) {
	if ui.StructuredDetailsView == nil {
		return
	}
	ui.StructuredDetailsView.SetText(formatStructuredDetails(report, ui.uiLang))
}

func (ui *TestUI) ApplyStructuredReportJSON(data []byte) error {
	report, err := decodeStructuredRun(data)
	if err != nil {
		return err
	}
	ui.ApplyStructuredReport(report)
	return nil
}

func (ui *TestUI) updatePartialReason(reason string) {
	if ui.PartialReasonLabel == nil {
		return
	}
	if strings.TrimSpace(reason) == "" {
		ui.PartialReasonLabel.SetText("")
		ui.PartialReasonLabel.Hide()
		return
	}
	ui.PartialReasonLabel.SetText(reason)
	ui.PartialReasonLabel.Show()
}

func (ui *TestUI) updateStructuredProgress(report StructuredRunResult) {
	if ui.CurrentItem == nil || ui.ProgressBar == nil {
		return
	}
	enabled, completed := 0, 0
	current := ""
	for _, section := range report.Sections {
		if !section.Enabled || section.Status == "skipped" {
			continue
		}
		enabled++
		if section.Status == "ok" || section.Status == "unavailable" || section.Status == "error" || section.Status == "timeout" || section.Status == "canceled" || section.Status == "partial" {
			completed++
		}
		if current == "" && section.Status != "ok" && section.Status != "skipped" {
			current = section.Name
		}
	}
	if enabled > 0 {
		ui.ProgressBar.SetValue(float64(completed) / float64(enabled))
	}
	if current != "" {
		if key, ok := sectionProgressKeys[current]; ok {
			current = ui.tr(key)
		}
		ui.CurrentItem.SetText(fmt.Sprintf("%s (%d/%d)", current, completed, enabled))
	}
}

func structuredStatusKey(status string) string {
	switch status {
	case "ok":
		return "status.done"
	case "partial":
		return "status.partial"
	case "timeout":
		return "status.timeout"
	case "canceled":
		return "status.stopped"
	case "error":
		return "status.failed"
	default:
		return ""
	}
}

func structuredReportJSON(report StructuredRunResult) ([]byte, error) {
	return json.Marshal(report)
}

func buildGUIStructuredReport(config ExecutionConfig, connected bool, tracker *progressTracker, runErr error, ctx context.Context, startedAt, finishedAt time.Time) StructuredRunResult {
	status, reason := "partial", "structured API unavailable in legacy build"
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		status, reason = "timeout", context.DeadlineExceeded.Error()
	} else if errors.Is(ctx.Err(), context.Canceled) {
		status, reason = "canceled", context.Canceled.Error()
	} else if runErr != nil {
		status, reason = "error", runErr.Error()
	}
	sections := buildGUIStructuredSections(config, connected, tracker, status, reason)
	if status == "partial" {
		// Legacy execution has text output only. Mark even completed-looking
		// sections as partial so the GUI never presents synthetic data as an
		// authoritative structured result.
		for index := range sections {
			if sections[index].Enabled && sections[index].Status == "ok" {
				sections[index].Status = "partial"
				sections[index].Reason = reason
			}
		}
	}
	return StructuredRunResult{
		SchemaVersion: structuredReportSchema,
		ECSVersion:    ecsVersion,
		Status:        status,
		StartedAt:     startedAt,
		FinishedAt:    finishedAt,
		DurationMS:    finishedAt.Sub(startedAt).Milliseconds(),
		Sections:      sections,
	}
}

func buildGUIStructuredSections(config ExecutionConfig, connected bool, tracker *progressTracker, runStatus, runReason string) []StructuredSection {
	selected := config.SelectedOptions
	mediaEnabled := selected["unlock"] && !config.ChinaModeEnabled
	pingEnabled := selected["ping"] || config.ChinaModeEnabled
	tgdcEnabled := config.PingTgdc && !config.ChinaModeEnabled
	webEnabled := config.PingWeb && !config.ChinaModeEnabled
	definitions := []struct {
		name, progress string
		enabled        bool
		network        bool
	}{
		{"basics", "progress.basic_security", selected["basic"], false},
		{"cpu", "progress.cpu", selected["cpu"], false},
		{"memory", "progress.memory", selected["memory"], false},
		{"disk", "progress.disk", selected["disk"], false},
		{"media", "progress.unlock", mediaEnabled, true},
		{"security", "progress.ip_quality", selected["security"], true},
		{"email", "progress.email", selected["email"], true},
		{"backtrace", "progress.backtrace", selected["backtrace"], true},
		{"routes", "progress.nt3", selected["nt3"], true},
		{"ping", "progress.ping", pingEnabled, true},
		{"tgdc", "progress.tgdc", tgdcEnabled, true},
		{"web", "progress.web", webEnabled, true},
		{"speed", "progress.speed", selected["speed"], true},
	}
	sections := make([]StructuredSection, 0, len(definitions))
	for _, definition := range definitions {
		section := StructuredSection{Name: definition.name, Enabled: definition.enabled}
		switch {
		case !definition.enabled:
			section.Status, section.Reason = "skipped", "disabled"
		case definition.network && !connected:
			section.Status, section.Reason = "unavailable", "network unavailable"
		case tracker != nil && tracker.finished[definition.progress]:
			section.Status = "ok"
		case runStatus != "ok":
			section.Status, section.Reason = runStatus, runReason
		default:
			section.Status, section.Reason = "partial", "no structured section result"
		}
		sections = append(sections, section)
	}
	return sections
}
