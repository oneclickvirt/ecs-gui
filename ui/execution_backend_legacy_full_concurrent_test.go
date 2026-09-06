//go:build !ecs_structured

package ui

import (
	"context"
	"os"
	"testing"
	"time"

	ecsapi "github.com/oneclickvirt/ecs/api"
)

func TestLegacyFullConcurrentDelegatesToUpstreamChoiceTwo(t *testing.T) {
	fixture, err := os.ReadFile("testdata/goecs_report_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	runCalls := 0
	var output string
	var progress []ProgressUpdate
	outcome := runLegacyFullConcurrent(context.Background(), ExecutionConfig{
		PresetKey: "full_concurrent", Language: "zh",
		SelectedOptions: map[string]bool{
			"basic": true, "cpu": true, "memory": true, "disk": true, "unlock": true,
			"security": true, "email": true, "backtrace": true, "nt3": true, "speed": true, "ping": true,
		},
	}, func(text string) { output += text }, func(update ProgressUpdate) { progress = append(progress, update) }, legacyFullConcurrentAPIDeps{
		checkPublicAccess: func(time.Duration) ecsapi.NetCheckResult {
			return ecsapi.NetCheckResult{Connected: true, StackType: "DualStack"}
		},
		runAllTests: func(_ context.Context, _ ecsapi.NetCheckResult, config *ecsapi.Config, observer ecsapi.ProgressObserver) *ecsapi.RunResult {
			runCalls++
			if config.Choice != "2" {
				t.Fatalf("full concurrent upstream choice = %q, want 2", config.Choice)
			}
			if !config.CpuTestStatus || !config.MemoryTestStatus || !config.DiskTestStatus || !config.SpeedTestStatus {
				t.Fatalf("full concurrent coverage was not mapped: %#v", config)
			}
			observer(ecsapi.ProgressEvent{Section: "cpu", Phase: ecsapi.ProgressStarted})
			return &ecsapi.RunResult{Output: "fixture output\n", JSON: fixture}
		},
	})
	if outcome.Err != nil || !outcome.Structured || outcome.Report == nil || runCalls != 1 {
		t.Fatalf("unexpected outcome: %#v calls=%d", outcome, runCalls)
	}
	if output != "fixture output\n" {
		t.Fatalf("output = %q", output)
	}
	foundCPU := false
	for _, update := range progress {
		if update.ItemKey == "progress.cpu" {
			foundCPU = true
			break
		}
	}
	if !foundCPU {
		t.Fatalf("upstream progress was not forwarded: %#v", progress)
	}
}
