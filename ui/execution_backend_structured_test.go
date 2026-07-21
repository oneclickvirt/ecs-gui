//go:build ecs_structured

package ui

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	ecsapi "github.com/oneclickvirt/ecs/api"
)

func TestStructuredExecutionRunnerUsesOfflineFixtureAndSingleAPICall(t *testing.T) {
	fixture, err := os.ReadFile("testdata/goecs_report_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	var gotCtx context.Context
	var gotConfig *ecsapi.Config
	checkCalls, runCalls := 0, 0
	runner := structuredExecutionRunner{api: structuredAPIDeps{
		checkPublicAccess: func(time.Duration) ecsapi.NetCheckResult {
			checkCalls++
			return ecsapi.NetCheckResult{Connected: false, StackType: "None"}
		},
		runAllTests: func(callCtx context.Context, _ ecsapi.NetCheckResult, config *ecsapi.Config, observer ecsapi.ProgressObserver) *ecsapi.RunResult {
			runCalls++
			gotCtx, gotConfig = callCtx, config
			observer(ecsapi.ProgressEvent{Section: "basics", Phase: ecsapi.ProgressStarted})
			observer(ecsapi.ProgressEvent{Section: "basics", Phase: ecsapi.ProgressCompleted, Status: ecsapi.ReportStatusOK})
			return &ecsapi.RunResult{Output: "fixture output\n", JSON: fixture}
		},
	}}

	var output string
	var progressUpdates []ProgressUpdate
	outcome := runner.Run(ctx, ExecutionConfig{
		Language: "zh", SelectedOptions: map[string]bool{"basic": true},
		CpuMethod: "sysbench", ThreadMode: "multi", MemoryMethod: "stream",
		DiskMethod: "fio", DeepMode: true,
		DeepDiskPaths: "/mnt/a,/mnt/b", DeepSMARTDevices: "/dev/sda",
		DeepBurnDuration: 45 * time.Second, DeepGPUDevice: "gpu0",
	}, func(value string) { output += value }, func(update ProgressUpdate) { progressUpdates = append(progressUpdates, update) })
	if outcome.Err != nil || outcome.Report == nil {
		t.Fatalf("unexpected outcome: %#v", outcome)
	}
	if checkCalls != 1 || runCalls != 1 || gotCtx == nil || gotCtx.Err() != nil {
		t.Fatalf("calls/context: check=%d run=%d ctx=%v", checkCalls, runCalls, gotCtx)
	}
	if gotConfig == nil || gotConfig.MaxDuration != 15*time.Minute || gotConfig.HardwareBudget != 15*time.Minute || !gotConfig.BasicStatus || gotConfig.CpuTestStatus {
		t.Fatalf("unexpected mapped config: %#v", gotConfig)
	}
	if !gotConfig.DeepMode || gotConfig.DeepDiskPaths != "/mnt/a,/mnt/b" || gotConfig.DeepSMARTDevices != "/dev/sda" || gotConfig.DeepBurnDuration != 45*time.Second || gotConfig.DeepGPUDevice != "gpu0" {
		t.Fatalf("deep config was not mapped: %#v", gotConfig)
	}
	if output != "fixture output\n" || len(outcome.Report.Components) != 1 || outcome.Report.Components[0].Name != "basics" {
		t.Fatalf("fixture was not propagated: output=%q report=%#v", output, outcome.Report)
	}
	foundRunningBasic := false
	for _, update := range progressUpdates {
		if update.ItemKey == "progress.basic_security" {
			foundRunningBasic = true
			break
		}
	}
	if !foundRunningBasic {
		t.Fatalf("structured API progress was not forwarded: %#v", progressUpdates)
	}
}

func TestStructuredAPIConfigKeepsStandardHardwareBudget(t *testing.T) {
	config := structuredAPIConfig(ExecutionConfig{
		SelectedOptions: map[string]bool{"disk": true},
		DeepDiskPaths:   "/must/not/run", DeepSMARTDevices: "/dev/must-not-run",
		DeepBurnDuration: time.Minute, DeepGPUDevice: "must-not-run",
	})
	if config.DeepMode || config.HardwareBudget != 2*time.Minute || config.MaxDuration != 15*time.Minute {
		t.Fatalf("unexpected standard budgets: %#v", config)
	}
	if config.DeepDiskPaths != "" || config.DeepSMARTDevices != "" || config.DeepBurnDuration != 0 || config.DeepGPUDevice != "" {
		t.Fatalf("standard mode retained deep targets: %#v", config)
	}
}

func TestStructuredAPIConfigMapsUnlockNetworkInputs(t *testing.T) {
	config := structuredAPIConfig(ExecutionConfig{
		SelectedOptions: map[string]bool{"unlock": true},
		UnlockInterface: "eth0", UnlockDNS: "1.1.1.1",
		UnlockHTTPProxy: "http://127.0.0.1:8080", UnlockSOCKSProxy: "socks5://127.0.0.1:1080",
		UnlockConcurrency: 12,
	})
	if config.UnlockTestInterface != "eth0" || config.UnlockTestDNSServers != "1.1.1.1" || config.UnlockTestConcurrency != 12 {
		t.Fatalf("unlock network config was not mapped: %#v", config)
	}
	if config.UnlockTestHTTPProxy == "" || config.UnlockTestSOCKSProxy == "" {
		t.Fatalf("unlock proxy config was not mapped: %#v", config)
	}
}

func TestStructuredAPIConfigMapsRuntimeDataAndPrivacyInputs(t *testing.T) {
	config := structuredAPIConfig(ExecutionConfig{
		SelectedOptions: map[string]bool{"basic": true},
		MaxDuration:     9 * time.Minute, HardwareBudget: 90 * time.Second,
		DataOffline: true,
		PrivacyMode: true, JSONPath: "result.json", EnableUpload: true,
	})
	if config.MaxDuration != 9*time.Minute || config.HardwareBudget != 90*time.Second {
		t.Fatalf("runtime budgets were not mapped: %#v", config)
	}
	if !config.DataOffline || !config.PrivacyMode || config.JSONPath != "result.json" || config.EnableUpload {
		t.Fatalf("data/privacy options were not mapped: %#v", config)
	}
}

func TestStructuredExecutionRunnerHonorsCanceledContextBeforePrecheck(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	checkCalls := 0
	runner := structuredExecutionRunner{api: structuredAPIDeps{
		checkPublicAccess: func(time.Duration) ecsapi.NetCheckResult {
			checkCalls++
			return ecsapi.NetCheckResult{}
		},
		runAllTests: func(context.Context, ecsapi.NetCheckResult, *ecsapi.Config, ecsapi.ProgressObserver) *ecsapi.RunResult {
			t.Fatal("run API must not be called for a canceled context")
			return nil
		},
	}}
	outcome := runner.Run(ctx, ExecutionConfig{}, nil, nil)
	if outcome.Err != context.Canceled || checkCalls != 0 {
		t.Fatalf("unexpected canceled outcome: %#v calls=%d", outcome, checkCalls)
	}
}

func TestStructuredExecutionRunnerFinalizesFilesAndUpload(t *testing.T) {
	fixture, err := os.ReadFile("testdata/goecs_report_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	finalizeCalls := 0
	runner := structuredExecutionRunner{api: structuredAPIDeps{
		checkPublicAccess: func(time.Duration) ecsapi.NetCheckResult {
			return ecsapi.NetCheckResult{Connected: true}
		},
		runAllTests: func(context.Context, ecsapi.NetCheckResult, *ecsapi.Config, ecsapi.ProgressObserver) *ecsapi.RunResult {
			return &ecsapi.RunResult{Output: "fixture output\n", JSON: fixture}
		},
		finalize: func(_ context.Context, _ ecsapi.NetCheckResult, config *ecsapi.Config, result *ecsapi.RunResult) (ecsapi.FinalizeResult, error) {
			finalizeCalls++
			if config.FilePath != "result.txt" || !config.EnableUpload || result.Output == "" {
				t.Fatalf("finalize input was not mapped: config=%#v result=%#v", config, result)
			}
			return ecsapi.FinalizeResult{HTTPSURL: "https://example.test/result"}, nil
		},
	}}
	var output string
	outcome := runner.Run(context.Background(), ExecutionConfig{
		SelectedOptions: map[string]bool{}, FilePath: "result.txt", EnableUpload: true,
	}, func(value string) { output += value }, nil)
	if outcome.Err != nil || finalizeCalls != 1 || !strings.Contains(output, "https://example.test/result") {
		t.Fatalf("unexpected finalize outcome: %#v calls=%d output=%q", outcome, finalizeCalls, output)
	}
}
