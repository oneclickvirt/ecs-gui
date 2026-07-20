package ui

import "context"

// executionOutcome is the only result crossing the execution/UI boundary.
// A structured build fills Report from the ecs API; the legacy build returns
// an explicitly partial compatibility report.
type executionOutcome struct {
	Err        error
	Report     *StructuredRunResult
	Structured bool
}

type executionRunner interface {
	Run(context.Context, ExecutionConfig, func(string), func(ProgressUpdate)) executionOutcome
}

func executeWithRunner(ctx context.Context, runner executionRunner, config ExecutionConfig, output func(string), progress func(ProgressUpdate)) executionOutcome {
	if runner == nil {
		return executionOutcome{Err: errExecutionRunnerUnavailable}
	}
	return runner.Run(ctx, config, output, progress)
}
