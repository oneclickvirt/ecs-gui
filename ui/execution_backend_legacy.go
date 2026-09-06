//go:build !ecs_structured

package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	ecsapi "github.com/oneclickvirt/ecs/api"
)

type legacyExecutionRunner struct{}

type legacyFullConcurrentAPIDeps struct {
	checkPublicAccess func(time.Duration) ecsapi.NetCheckResult
	runAllTests       func(context.Context, ecsapi.NetCheckResult, *ecsapi.Config, ecsapi.ProgressObserver) *ecsapi.RunResult
	finalize          func(context.Context, ecsapi.NetCheckResult, *ecsapi.Config, *ecsapi.RunResult) (ecsapi.FinalizeResult, error)
}

func newExecutionRunner() executionRunner {
	return legacyExecutionRunner{}
}

func (legacyExecutionRunner) Run(ctx context.Context, config ExecutionConfig, output func(string), progress func(ProgressUpdate)) executionOutcome {
	if config.PresetKey == "full_concurrent" {
		return runLegacyFullConcurrent(ctx, config, output, progress, legacyFullConcurrentAPIDeps{
			checkPublicAccess: ecsapi.CheckPublicAccess,
			runAllTests:       ecsapi.RunAllTestsContextWithProgress,
			finalize:          ecsapi.FinalizeRunResultContext,
		})
	}
	executor := NewCommandExecutor(output)
	executor.SetProgressCallback(progress)
	executor.SetContext(ctx)
	err := executor.Execute(config)
	report, _ := executor.StructuredResult()
	return executionOutcome{Err: err, Report: report, Structured: false}
}

// runLegacyFullConcurrent keeps the normal GUI build on its legacy executor
// for ordinary profiles, but delegates option 2 to the upstream runner. That
// runner owns the buffered, fully concurrent schedule and emits the same
// option-1 chapter order after all tasks have completed.
func runLegacyFullConcurrent(ctx context.Context, config ExecutionConfig, output func(string), progress func(ProgressUpdate), api legacyFullConcurrentAPIDeps) executionOutcome {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return executionOutcome{Err: err, Structured: true}
	}
	if api.checkPublicAccess == nil || api.runAllTests == nil {
		return executionOutcome{Err: errExecutionRunnerUnavailable, Structured: true}
	}
	defer ecsapi.ShutdownDNS()

	tracker := newProgressTracker(progress, nil)
	tracker.start("progress.precheck")
	preCheck := api.checkPublicAccess(3 * time.Second)
	if err := ctx.Err(); err != nil {
		return executionOutcome{Err: err, Structured: true}
	}
	tracker.steps = buildProgressSteps(config, preCheck.Connected)
	tracker.finish("progress.precheck")

	apiConfig := apiConfigForExecution(config)
	apiConfig.Choice = "2"
	if apiConfig.EnableUpload {
		apiConfig.FilePath = safeUploadFilePath(config.FilePath)
		defer os.Remove(apiConfig.FilePath)
	} else {
		apiConfig.FilePath = ""
	}
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
	callCtx := ecsapi.WithProgressObserver(ctx, observer)
	result := api.runAllTests(callCtx, preCheck, apiConfig, observer)
	if result == nil {
		return executionOutcome{Err: errors.New("full concurrent API returned nil result"), Structured: true}
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
			output(sanitizeGUIText(text))
		}
	}

	var finalizeErr error
	if api.finalize != nil {
		finalized, err := api.finalize(callCtx, preCheck, apiConfig, result)
		finalizeErr = err
		if output != nil && finalized.HTTPSURL != "" {
			output(fmt.Sprintf("Share URL: %s\n", finalized.HTTPSURL))
		}
	}
	if progress != nil {
		progressFromStructuredReport(progress, *report)
	}
	tracker.finish("progress.finish")
	return executionOutcome{Err: finalizeErr, Report: report, Structured: true}
}
