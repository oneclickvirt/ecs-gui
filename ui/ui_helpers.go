package ui

// hasSelectedTests 检查是否有选中的测试项
func (ui *TestUI) hasSelectedTests() bool {
	return ui.BasicCheck.Checked ||
		ui.CpuCheck.Checked ||
		ui.MemoryCheck.Checked ||
		ui.DiskCheck.Checked ||
		ui.CommCheck.Checked ||
		ui.UnlockCheck.Checked ||
		ui.SecurityCheck.Checked ||
		ui.EmailCheck.Checked ||
		ui.BacktraceCheck.Checked ||
		ui.Nt3Check.Checked ||
		ui.SpeedCheck.Checked ||
		ui.PingCheck.Checked
}

// isCancelled 检查测试是否被取消
func (ui *TestUI) isCancelled() bool {
	select {
	case <-ui.CancelCtx.Done():
		return true
	default:
		return false
	}
}

// resetUIState 重置UI状态
// 注意：此方法可能从 goroutine 中调用
// 已在 main.go 中设置 FYNE_DISABLE_DRIVER_THREAD_CHECK=1 来处理线程安全
func (ui *TestUI) resetUIState() {
	ui.Mu.Lock()
	ui.IsRunning = false
	ui.Mu.Unlock()

	// UI更新 - Fyne 会自动处理线程调度
	ui.StartButton.Enable()
	ui.StopButton.Disable()
	ui.ProgressBar.Hide()
	ui.ProgressBar.SetValue(0)
}

// GetSelectedOptions 获取所有选中的测试选项
func (ui *TestUI) GetSelectedOptions() map[string]bool {
	return map[string]bool{
		"basic":     ui.BasicCheck.Checked,
		"cpu":       ui.CpuCheck.Checked,
		"memory":    ui.MemoryCheck.Checked,
		"disk":      ui.DiskCheck.Checked,
		"comm":      ui.CommCheck.Checked,
		"unlock":    ui.UnlockCheck.Checked,
		"security":  ui.SecurityCheck.Checked,
		"email":     ui.EmailCheck.Checked,
		"backtrace": ui.BacktraceCheck.Checked,
		"nt3":       ui.Nt3Check.Checked,
		"speed":     ui.SpeedCheck.Checked,
		"ping":      ui.PingCheck.Checked,
	}
}
