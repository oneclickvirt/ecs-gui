//go:build !ecs_structured

package ui

import "context"

type legacyExecutionRunner struct{}

func newExecutionRunner() executionRunner {
	return legacyExecutionRunner{}
}

func (legacyExecutionRunner) Run(ctx context.Context, config ExecutionConfig, output func(string), progress func(ProgressUpdate)) executionOutcome {
	executor := NewCommandExecutor(output)
	executor.SetProgressCallback(progress)
	executor.SetContext(ctx)
	err := executor.Execute(config)
	report, _ := executor.StructuredResult()
	return executionOutcome{Err: err, Report: report, Structured: false}
}
