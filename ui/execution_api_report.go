package ui

import (
	"errors"
	"fmt"

	ecsapi "github.com/oneclickvirt/ecs/api"
)

func structuredReportFromRunResult(result *ecsapi.RunResult) (*StructuredRunResult, error) {
	data := result.JSON
	if len(data) == 0 && result.Report != nil {
		encoded, err := result.Report.JSON()
		if err != nil {
			return nil, fmt.Errorf("encode structured API report: %w", err)
		}
		data = encoded
	}
	if len(data) == 0 {
		return nil, errors.New("structured API returned no report payload")
	}
	report, err := decodeStructuredRun(data)
	if err != nil {
		return nil, fmt.Errorf("decode structured API report: %w", err)
	}
	return &report, nil
}

func progressFromStructuredReport(progress func(ProgressUpdate), report StructuredRunResult) {
	if progress == nil {
		return
	}
	enabled, completed := 0, 0
	for _, section := range report.Sections {
		if !section.Enabled || section.Status == "skipped" {
			continue
		}
		enabled++
		if section.Status != "" && section.Status != "running" {
			completed++
		}
	}
	if enabled == 0 {
		enabled = 1
	}
	progress(ProgressUpdate{ItemKey: "progress.finish", Current: completed, Total: enabled, Fraction: float64(completed) / float64(enabled)})
}
