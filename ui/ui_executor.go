package ui

import (
	"fmt"
	"time"
)

// runTestsWithExecutor 使用命令执行器运行测试
func (ui *TestUI) runTestsWithExecutor(config ExecutionConfig) {
	// 添加错误恢复
	defer func() {
		if r := recover(); r != nil {
			// 安全地更新UI
			errorMsg := fmt.Sprintf("%s%v\n", ui.tr("log.fatal_prefix"), r)
			ui.Terminal.AppendText(errorMsg)
			ui.runOnUI(func() {
				ui.setStatus("status.failed")
			})
		}
		// 确保UI状态被重置
		ui.resetUIState()
	}()

	startTime := time.Now()

	// 创建命令执行器
	executor := NewCommandExecutor(func(text string) {
		// 这个回调会从 executor 的 goroutine 调用
		// TerminalOutput 的 AppendText 已经是线程安全的
		ui.Terminal.AppendText(text)
	})

	// 设置执行上下文
	executor.SetContext(ui.CancelCtx)

	// 更新进度
	ui.runOnUI(func() {
		ui.ProgressBar.SetValue(0.1)
		ui.setStatus("status.executing")
	})

	// 执行测试（输出会实时显示在terminal widget中）
	err := executor.Execute(config)

	// 显示结束信息
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	_ = duration // 避免未使用警告

	if err != nil {
		ui.Terminal.AppendText(fmt.Sprintf("%s%v\n", ui.tr("log.error_prefix"), err))

		// 检查是否是取消导致的
		if ui.isCancelled() {
			ui.runOnUI(func() {
				ui.setStatus("status.stopped")
			})
		} else {
			ui.runOnUI(func() {
				ui.setStatus("status.failed")
			})
		}
	} else if ui.isCancelled() {
		ui.Terminal.AppendText(ui.tr("log.interrupted_short"))
		ui.runOnUI(func() {
			ui.setStatus("status.stopped")
		})
	} else {
		ui.runOnUI(func() {
			ui.setStatus("status.done")
			ui.ProgressBar.SetValue(1.0)
		})

		// 如果启用了日志，自动刷新日志内容
		if config.LogEnabled {
			time.Sleep(500 * time.Millisecond) // 等待日志文件写入完成
			ui.refreshLogFromFile()
		}
	}
}
