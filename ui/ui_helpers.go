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

// resetUIState 重置UI状态（线程安全）
func (ui *TestUI) resetUIState() {
	ui.Mu.Lock()
	ui.IsRunning = false
	ui.Mu.Unlock()

	// UI更新应该在主线程，但 Fyne 的组件更新已经处理了线程安全
	ui.StartButton.Enable()
	ui.StopButton.Disable()
	ui.ProgressBar.Hide()
	ui.ProgressBar.SetValue(0)

	// 如果状态还是"正在停止..."，更新为"就绪"
	if ui.StatusLabel.Text == "正在停止..." {
		ui.StatusLabel.SetText("测试已停止")
	}
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
