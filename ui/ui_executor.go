package ui

import (
	"fmt"
	"time"
)

// runTestsWithExecutor 使用命令执行器运行测试
func (ui *TestUI) runTestsWithExecutor(config ExecutionConfig) {
	startTime := time.Now()

	// 添加错误恢复
	defer func() {
		if r := recover(); r != nil {
			// 安全地更新UI
			errorMsg := fmt.Sprintf("%s%s\n", ui.tr("log.fatal_prefix"), ui.tr("error.generic"))
			ui.Terminal.AppendText(errorMsg)
			ui.runOnUI(func() {
				ui.setStatus("status.failed")
			})
			ui.notifyTestFinished("status.failed", time.Since(startTime))
		}
		// 确保UI状态被重置
		ui.resetUIState()
	}()

	// The build-specific runner owns the single execution. Legacy builds wrap
	// CommandExecutor; ecs_structured builds call ecs/api directly.
	output := func(text string) {
		// 这个回调会从 executor 的 goroutine 调用
		// TerminalOutput 的 AppendText 已经是线程安全的
		ui.Terminal.AppendText(text)
	}
	progress := func(update ProgressUpdate) {
		ui.runOnUI(func() {
			ui.setProgress(update)
		})
	}

	// 更新进度
	ui.runOnUI(func() {
		ui.ProgressBar.SetValue(0.02)
		ui.setStatus("status.executing")
	})

	// Execute exactly once through the selected build backend. Structured
	// builds receive the same cancellation context all the way into goecs/api.
	outcome := executeWithRunner(ui.CancelCtx, newExecutionRunner(), config, output, progress)
	err := outcome.Err
	var reportReason string
	structuredStatus := ""
	if outcome.Report != nil {
		structuredStatus = outcome.Report.Status
		_, reportReason = summarizeStructuredRun(*outcome.Report)
		report := *outcome.Report
		ui.runOnUI(func() { ui.ApplyStructuredReport(report) })
	}
	if err != nil && reportReason == "" {
		ui.runOnUI(func() { ui.updatePartialReason(ui.friendlyErrorMessage(err)) })
	}
	if err != nil {
		ui.Terminal.AppendText(fmt.Sprintf("%s%s\n", ui.tr("log.error_prefix"), ui.friendlyErrorMessage(err)))
	}

	// A structured status is authoritative. Do not replace a partial report
	// with a generic "done" status merely because the API returned no error.
	if structuredStatus != "" {
		switch structuredStatus {
		case "timeout":
			ui.runOnUI(func() { ui.setStatus("status.timeout") })
			ui.notifyTestFinished("status.failed", durationSince(startTime))
		case "canceled":
			ui.runOnUI(func() { ui.setStatus("status.stopped") })
			ui.notifyTestFinished("status.stopped", durationSince(startTime))
		case "error":
			ui.runOnUI(func() { ui.setStatus("status.failed") })
			ui.notifyTestFinished("status.failed", durationSince(startTime))
		case "partial", "unavailable":
			ui.runOnUI(func() {
				ui.setStatus("status.partial")
				ui.ProgressBar.SetValue(1)
			})
			ui.notifyTestFinished("status.done", durationSince(startTime))
		default:
			ui.runOnUI(func() {
				ui.setStatus("status.done")
				ui.ProgressBar.SetValue(1.0)
			})
			ui.notifyTestFinished("status.done", durationSince(startTime))
		}
	} else if err != nil {
		// Legacy execution normally supplies a partial/error report. This branch
		// remains for startup failures before a compatibility report exists.
		// 区分手工取消、15 分钟硬超时和普通失败。
		if ui.isTimedOut() {
			ui.runOnUI(func() {
				ui.setStatus("status.timeout")
			})
			ui.notifyTestFinished("status.failed", durationSince(startTime))
		} else if ui.isCancelled() {
			ui.runOnUI(func() {
				ui.setStatus("status.stopped")
			})
			ui.notifyTestFinished("status.stopped", durationSince(startTime))
		} else {
			ui.runOnUI(func() {
				ui.setStatus("status.failed")
			})
			ui.notifyTestFinished("status.failed", durationSince(startTime))
		}
	} else if ui.isTimedOut() {
		ui.runOnUI(func() {
			ui.setStatus("status.timeout")
		})
		ui.notifyTestFinished("status.failed", durationSince(startTime))
	} else if ui.isCancelled() {
		ui.Terminal.AppendText(ui.tr("log.interrupted_short"))
		ui.runOnUI(func() {
			ui.setStatus("status.stopped")
		})
		ui.notifyTestFinished("status.stopped", durationSince(startTime))
	} else {
		ui.runOnUI(func() {
			ui.setStatus("status.done")
			ui.ProgressBar.SetValue(1.0)
		})
		ui.notifyTestFinished("status.done", durationSince(startTime))
	}

	// Structured and legacy backends use the same component log file. Refresh
	// after every terminal state so partial and failed runs remain inspectable.
	if config.LogEnabled {
		time.Sleep(500 * time.Millisecond)
		ui.refreshLogFromFile()
	}
}

func durationSince(start time.Time) time.Duration {
	if start.IsZero() {
		return 0
	}
	return time.Since(start)
}
