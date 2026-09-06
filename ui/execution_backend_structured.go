//go:build ecs_structured

package ui

import (
	"context"
	"errors"
	"fmt"
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
	defer ecsapi.ShutdownDNS()
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
			output(sanitizeGUIText(text))
		}
	}
	var finalizeErr error
	if runner.api.finalize != nil {
		finalized, err := runner.api.finalize(finalizeCtx, preCheck, apiConfig, result)
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

func structuredAPIConfig(config ExecutionConfig) *ecsapi.Config {
	return apiConfigForExecution(config)
}
