//go:build ecs_structured

package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	ecsapi "github.com/oneclickvirt/ecs/api"
)

type structuredAPIDeps struct {
	checkPublicAccess func(time.Duration) ecsapi.NetCheckResult
	runAllTests       func(context.Context, ecsapi.NetCheckResult, *ecsapi.Config, ecsapi.ProgressObserver) *ecsapi.RunResult
	finalize          func(context.Context, ecsapi.NetCheckResult, *ecsapi.Config, *ecsapi.RunResult) (ecsapi.FinalizeResult, error)
}

type structuredExecutionRunner struct {
	api structuredAPIDeps
}

func newExecutionRunner() executionRunner {
	return structuredExecutionRunner{api: structuredAPIDeps{
		checkPublicAccess: ecsapi.CheckPublicAccess,
		runAllTests:       ecsapi.RunAllTestsContextWithProgress,
		finalize:          ecsapi.FinalizeRunResultContext,
	}}
}

func (runner structuredExecutionRunner) Run(ctx context.Context, config ExecutionConfig, output func(string), progress func(ProgressUpdate)) executionOutcome {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return executionOutcome{Err: err, Structured: true}
	}
	if runner.api.checkPublicAccess == nil || runner.api.runAllTests == nil {
		return executionOutcome{Err: errExecutionRunnerUnavailable, Structured: true}
	}
	tracker := newProgressTracker(progress, nil)
	tracker.start("progress.precheck")
	preCheck := runner.api.checkPublicAccess(3 * time.Second)
	if err := ctx.Err(); err != nil {
		return executionOutcome{Err: err, Structured: true}
	}
	tracker.steps = buildProgressSteps(config, preCheck.Connected)
	tracker.finish("progress.precheck")

	apiConfig := structuredAPIConfig(config)
	observer := func(event ecsapi.ProgressEvent) {
		key, ok := sectionProgressKeys[event.Section]
		if !ok {
			return
		}
		if event.Phase == ecsapi.ProgressStarted {
			tracker.start(key)
			return
		}
		tracker.finish(key)
	}
	finalizeCtx := ecsapi.WithProgressObserver(ctx, observer)
	result := runner.api.runAllTests(finalizeCtx, preCheck, apiConfig, observer)
	if result == nil {
		return executionOutcome{Err: errors.New("structured API returned nil result"), Structured: true}
	}

	report, err := structuredReportFromRunResult(result)
	if err != nil {
		return executionOutcome{Err: err, Structured: true}
	}
	if output != nil {
		text := result.Output
		if text == "" && report.Text != "" {
			text = report.Text
		}
		if text != "" {
			output(text)
		}
	}
	var finalizeErr error
	if runner.api.finalize != nil {
		finalized, err := runner.api.finalize(finalizeCtx, preCheck, apiConfig, result)
		finalizeErr = err
		if output != nil && (finalized.HTTPURL != "" || finalized.HTTPSURL != "") {
			output(fmt.Sprintf("Http URL:  %s\nHttps URL: %s\n", finalized.HTTPURL, finalized.HTTPSURL))
		}
	}
	if progress != nil {
		progressFromStructuredReport(progress, *report)
	}
	tracker.finish("progress.finish")
	return executionOutcome{Err: finalizeErr, Report: report, Structured: true}
}

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

func structuredAPIConfig(config ExecutionConfig) *ecsapi.Config {
	selected := config.SelectedOptions
	apiConfig := ecsapi.NewConfig(ecsVersion)
	apiConfig.MenuMode = false
	apiConfig.Language = config.Language
	apiConfig.CpuTestMethod = config.CpuMethod
	apiConfig.CpuTestThreadMode = config.ThreadMode
	apiConfig.MemoryTestMethod = config.MemoryMethod
	apiConfig.DiskTestMethod = config.DiskMethod
	apiConfig.DiskTestPath = config.DiskPath
	apiConfig.DiskMultiCheck = config.DiskMulti
	apiConfig.AutoChangeDiskMethod = config.AutoDiskMethod
	apiConfig.Nt3Location = config.Nt3Location
	apiConfig.Nt3CheckType = config.Nt3Type
	apiConfig.SpNum = config.SpNum
	apiConfig.Width = config.OutputWidth
	apiConfig.BasicStatus = selected["basic"]
	apiConfig.CpuTestStatus = selected["cpu"]
	apiConfig.MemoryTestStatus = selected["memory"]
	apiConfig.DiskTestStatus = selected["disk"]
	apiConfig.UtTestStatus = selected["unlock"] && !config.ChinaModeEnabled
	apiConfig.SecurityTestStatus = selected["security"]
	apiConfig.EmailTestStatus = selected["email"]
	apiConfig.BacktraceStatus = selected["backtrace"] && !config.ChinaModeEnabled
	apiConfig.Nt3Status = selected["nt3"] && !config.ChinaModeEnabled
	apiConfig.SpeedTestStatus = selected["speed"]
	apiConfig.PingTestStatus = selected["ping"] || config.ChinaModeEnabled
	apiConfig.PingSortOrder = config.PingSortOrder
	apiConfig.PingScope = config.PingScope
	apiConfig.TCPSortOrder = config.TCPSortOrder
	apiConfig.TgdcTestStatus = config.PingTgdc && !config.ChinaModeEnabled
	apiConfig.WebTestStatus = config.PingWeb && !config.ChinaModeEnabled
	apiConfig.OnlyChinaTest = config.ChinaModeEnabled
	apiConfig.UnlockTestRegion = config.UnlockRegion
	apiConfig.UnlockTestIPVersion = config.UnlockIpVersion
	apiConfig.UnlockTestShowIP = config.UnlockShowIP
	apiConfig.UnlockTestInterface = config.UnlockInterface
	apiConfig.UnlockTestDNSServers = config.UnlockDNS
	apiConfig.UnlockTestHTTPProxy = config.UnlockHTTPProxy
	apiConfig.UnlockTestSOCKSProxy = config.UnlockSOCKSProxy
	apiConfig.UnlockTestConcurrency = config.UnlockConcurrency
	apiConfig.EnableLogger = config.LogEnabled
	apiConfig.FilePath = config.FilePath
	apiConfig.EnableUpload = config.EnableUpload && !config.PrivacyMode
	apiConfig.AnalyzeResult = config.AnalyzeResult
	apiConfig.PrivacyMode = config.PrivacyMode
	apiConfig.DataOffline = config.DataOffline
	apiConfig.DeepMode = config.DeepMode
	if config.DeepMode {
		apiConfig.DeepDiskPaths = config.DeepDiskPaths
		apiConfig.DeepSMARTDevices = config.DeepSMARTDevices
		apiConfig.DeepBurnDuration = config.DeepBurnDuration
		apiConfig.DeepGPUDevice = config.DeepGPUDevice
	}
	apiConfig.TCPProbeStatus = true
	apiConfig.MaxDuration = config.MaxDuration
	if apiConfig.MaxDuration <= 0 || apiConfig.MaxDuration > 15*time.Minute {
		apiConfig.MaxDuration = 15 * time.Minute
	}
	hardwareLimit := min(2*time.Minute, apiConfig.MaxDuration)
	if config.DeepMode {
		hardwareLimit = apiConfig.MaxDuration
	}
	apiConfig.HardwareBudget = config.HardwareBudget
	if apiConfig.HardwareBudget <= 0 || apiConfig.HardwareBudget > hardwareLimit {
		apiConfig.HardwareBudget = hardwareLimit
	}
	apiConfig.JSONPath = strings.TrimSpace(config.JSONPath)
	return apiConfig
}
