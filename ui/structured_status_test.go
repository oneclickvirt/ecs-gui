package ui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestBuildGUIStructuredReportMarksOfflineNetworkSections(t *testing.T) {
	tracker := newProgressTracker(nil, []string{"progress.basic_security", "progress.cpu"})
	tracker.finish("progress.basic_security")
	report := buildGUIStructuredReport(ExecutionConfig{
		SelectedOptions: map[string]bool{"basic": true, "cpu": true, "speed": true, "unlock": true},
	}, false, tracker, nil, context.Background(), time.Now(), time.Now().Add(time.Second))
	statuses := make(map[string]string)
	for _, section := range report.Sections {
		statuses[section.Name] = section.Status
	}
	if report.SchemaVersion != structuredReportSchema || report.Status != "partial" || statuses["basics"] != "partial" || statuses["media"] != "unavailable" || statuses["speed"] != "unavailable" {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestBuildGUIStructuredReportPreservesTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	report := buildGUIStructuredReport(ExecutionConfig{SelectedOptions: map[string]bool{"cpu": true}}, true, nil, errors.New("cancelled"), ctx, time.Now(), time.Now())
	if report.Status != "canceled" {
		t.Fatalf("status=%q, want canceled", report.Status)
	}
}

func TestApplyStructuredReportUpdatesPartialReason(t *testing.T) {
	ui := newTestUIForTest(t)
	ui.ApplyStructuredReport(StructuredRunResult{
		SchemaVersion: structuredReportSchema,
		Status:        "partial",
		Sections:      []StructuredSection{{Name: "tcp", Enabled: true, Status: "unavailable", Reason: "network unavailable"}},
	})
	if !strings.Contains(ui.PartialReasonLabel.Text, "tcp") || ui.StatusLabel.Text != ui.tr("status.partial") {
		t.Fatalf("reason=%q status=%q", ui.PartialReasonLabel.Text, ui.StatusLabel.Text)
	}
}

func TestBuildGUIStructuredSectionsKeepsPingExtensionsSeparate(t *testing.T) {
	sections := buildGUIStructuredSections(ExecutionConfig{
		SelectedOptions: map[string]bool{"ping": false},
		PingTgdc:        true,
		PingWeb:         true,
	}, true, nil, "partial", "running")
	statuses := make(map[string]StructuredSection, len(sections))
	for _, section := range sections {
		statuses[section.Name] = section
	}
	if statuses["ping"].Enabled || !statuses["tgdc"].Enabled || !statuses["web"].Enabled {
		t.Fatalf("unexpected ping extension sections: %#v", statuses)
	}
}
