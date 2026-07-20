package ui

import (
	"context"
	"errors"
	"testing"
)

type fakeExecutionRunner struct {
	calls  int
	gotCtx context.Context
	result executionOutcome
}

func (fake *fakeExecutionRunner) Run(ctx context.Context, _ ExecutionConfig, output func(string), progress func(ProgressUpdate)) executionOutcome {
	fake.calls++
	fake.gotCtx = ctx
	if output != nil {
		output("fixture output\n")
	}
	if progress != nil {
		progress(ProgressUpdate{ItemKey: "progress.finish", Current: 1, Total: 1, Fraction: 1})
	}
	return fake.result
}

func TestExecuteWithRunnerInvokesSingleRunnerAndPropagatesContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fake := &fakeExecutionRunner{result: executionOutcome{Report: &StructuredRunResult{SchemaVersion: "goecs.report/v1", Status: "partial"}, Structured: true}}
	var output string
	var progress ProgressUpdate
	got := executeWithRunner(ctx, fake, ExecutionConfig{}, func(value string) { output += value }, func(value ProgressUpdate) { progress = value })
	if fake.calls != 1 {
		t.Fatalf("runner calls = %d, want 1", fake.calls)
	}
	if fake.gotCtx != ctx {
		t.Fatal("runner did not receive the caller context")
	}
	if output != "fixture output\n" || progress.Fraction != 1 || got.Report == nil || !got.Structured {
		t.Fatalf("unexpected outcome: output=%q progress=%#v outcome=%#v", output, progress, got)
	}
}

func TestExecuteWithRunnerRejectsNilRunner(t *testing.T) {
	got := executeWithRunner(context.Background(), nil, ExecutionConfig{}, nil, nil)
	if !errors.Is(got.Err, errExecutionRunnerUnavailable) {
		t.Fatalf("error = %v, want runner unavailable", got.Err)
	}
}
