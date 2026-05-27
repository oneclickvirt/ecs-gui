package ui

import (
	"context"
	"fmt"
	"io"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

const maxLogViewBytes int64 = 2 * 1024 * 1024

// onPresetChanged 预设模式改变时的处理
func (ui *TestUI) onPresetChanged(preset string) {
	if ui.suppressPresetChange {
		return
	}

	key, ok := ui.presetLabelToKey[preset]
	if !ok {
		key = "custom"
	}
	ui.selectedPresetKey = key

	switch key {
	case "full":
		// 对应原goecs.go的选项1: SetFullTestStatus
		ui.setAllChecks(true)
		// 注意：原goecs.go的完全体包括TGDC和Web测试，不包括三网ping测试
		ui.PingCheck.Checked = false
		ui.PingTgdcCheck.Checked = true
		ui.PingWebCheck.Checked = true
		ui.ChinaModeCheck.Checked = false
		// 测速配置：全部启用
		ui.SpTestUploadCheck.Checked = true
		ui.SpTestDownloadCheck.Checked = true
	case "minimal":
		// 对应原goecs.go的选项2: SetMinimalTestStatus
		ui.setAllChecks(false)
		ui.BasicCheck.Checked = true
		ui.CpuCheck.Checked = true
		ui.MemoryCheck.Checked = true
		ui.DiskCheck.Checked = true
		ui.SpeedCheck.Checked = true
		ui.PingTgdcCheck.Checked = false
		ui.PingWebCheck.Checked = false
		ui.ChinaModeCheck.Checked = false
		// 测速配置：全部启用
		ui.SpTestUploadCheck.Checked = true
		ui.SpTestDownloadCheck.Checked = true
	case "standard":
		// 对应原goecs.go的选项3: SetStandardTestStatus
		ui.setAllChecks(false)
		ui.BasicCheck.Checked = true
		ui.CpuCheck.Checked = true
		ui.MemoryCheck.Checked = true
		ui.DiskCheck.Checked = true
		ui.UnlockCheck.Checked = true
		ui.Nt3Check.Checked = true
		ui.SpeedCheck.Checked = true
		ui.PingTgdcCheck.Checked = false
		ui.PingWebCheck.Checked = false
		ui.ChinaModeCheck.Checked = false
		// 测速配置：全部启用
		ui.SpTestUploadCheck.Checked = true
		ui.SpTestDownloadCheck.Checked = true
	case "network_focus":
		// 对应原goecs.go的选项4: SetNetworkFocusedTestStatus
		ui.setAllChecks(false)
		ui.BasicCheck.Checked = true
		ui.CpuCheck.Checked = true
		ui.MemoryCheck.Checked = true
		ui.DiskCheck.Checked = true
		ui.BacktraceCheck.Checked = true
		ui.Nt3Check.Checked = true
		ui.SpeedCheck.Checked = true
		ui.PingTgdcCheck.Checked = false
		ui.PingWebCheck.Checked = false
		ui.ChinaModeCheck.Checked = false
		// 测速配置：全部启用
		ui.SpTestUploadCheck.Checked = true
		ui.SpTestDownloadCheck.Checked = true
	case "unlock_focus":
		// 对应原goecs.go的选项5: SetUnlockFocusedTestStatus
		ui.setAllChecks(false)
		ui.BasicCheck.Checked = true
		ui.CpuCheck.Checked = true
		ui.MemoryCheck.Checked = true
		ui.DiskCheck.Checked = true
		ui.CommCheck.Checked = true
		ui.UnlockCheck.Checked = true
		ui.SpeedCheck.Checked = true
		ui.PingTgdcCheck.Checked = false
		ui.PingWebCheck.Checked = false
		ui.ChinaModeCheck.Checked = false
		// 测速配置：全部启用
		ui.SpTestUploadCheck.Checked = true
		ui.SpTestDownloadCheck.Checked = true
	case "network_only":
		// 对应原goecs.go的选项6: SetNetworkOnlyTestStatus
		ui.setAllChecks(false)
		ui.BasicCheck.Checked = false // 6号选项不包括基础信息
		ui.SecurityCheck.Checked = true
		ui.SpeedCheck.Checked = true
		ui.BacktraceCheck.Checked = true
		ui.Nt3Check.Checked = true
		ui.PingCheck.Checked = true
		ui.PingTgdcCheck.Checked = true
		ui.PingWebCheck.Checked = true
		ui.ChinaModeCheck.Checked = false
		// 测速配置：全部启用
		ui.SpTestUploadCheck.Checked = true
		ui.SpTestDownloadCheck.Checked = true
	case "unlock_only":
		// 对应原goecs.go的选项7: SetUnlockOnlyTestStatus
		ui.setAllChecks(false)
		ui.CommCheck.Checked = true
		ui.UnlockCheck.Checked = true
		ui.PingTgdcCheck.Checked = false
		ui.PingWebCheck.Checked = false
		ui.ChinaModeCheck.Checked = false
		// 测速配置：禁用
		ui.SpTestUploadCheck.Checked = false
		ui.SpTestDownloadCheck.Checked = false
	case "hardware_only":
		// 对应原goecs.go的选项8: SetHardwareOnlyTestStatus
		ui.setAllChecks(false)
		ui.BasicCheck.Checked = true
		ui.CpuCheck.Checked = true
		ui.MemoryCheck.Checked = true
		ui.DiskCheck.Checked = true
		ui.DiskMethodSelect.SetSelected("auto") // 使用auto让系统自动选择dd和fio
		ui.PingTgdcCheck.Checked = false
		ui.PingWebCheck.Checked = false
		ui.ChinaModeCheck.Checked = false
		// 测速配置：禁用
		ui.SpTestUploadCheck.Checked = false
		ui.SpTestDownloadCheck.Checked = false
	case "ip_quality":
		// 对应原goecs.go的选项9: SetIPQualityTestStatus
		ui.setAllChecks(false)
		ui.BasicCheck.Checked = false // 9号选项不包括基础信息
		ui.SecurityCheck.Checked = true
		ui.EmailCheck.Checked = true
		ui.PingTgdcCheck.Checked = false
		ui.PingWebCheck.Checked = false
		ui.ChinaModeCheck.Checked = false
		// 测速配置：禁用
		ui.SpTestUploadCheck.Checked = false
		ui.SpTestDownloadCheck.Checked = false
	case "route_only":
		// 对应原goecs.go的选项10: SetRouteTestStatus + nt3Location = "ALL"
		ui.setAllChecks(false)
		ui.BasicCheck.Checked = false // 10号选项不包括基础信息
		ui.BacktraceCheck.Checked = true
		ui.Nt3Check.Checked = true
		ui.PingCheck.Checked = true
		ui.PingTgdcCheck.Checked = true
		ui.PingWebCheck.Checked = true
		ui.Nt3LocationSelect.SetSelected("ALL") // 设置为全部地点
		ui.ChinaModeCheck.Checked = false
		// 测速配置：禁用
		ui.SpTestUploadCheck.Checked = false
		ui.SpTestDownloadCheck.Checked = false
	default: // 自定义
		return
	}
	ui.refreshAllChecks()
	ui.refreshSpeedTestChecks()
}

// setAllChecks 设置所有测试项的选中状态
func (ui *TestUI) setAllChecks(checked bool) {
	for _, check := range ui.testChecks {
		if check != nil {
			check.Checked = checked
		}
	}
	ui.refreshAllChecks()
}

// refreshAllChecks 刷新所有测试项的显示
func (ui *TestUI) refreshAllChecks() {
	for _, check := range ui.testChecks {
		if check != nil {
			check.Refresh()
		}
	}
}

// refreshSpeedTestChecks 刷新测速配置的显示
func (ui *TestUI) refreshSpeedTestChecks() {
	if ui.SpTestUploadCheck != nil {
		ui.SpTestUploadCheck.Refresh()
	}
	if ui.SpTestDownloadCheck != nil {
		ui.SpTestDownloadCheck.Refresh()
	}
	if ui.PingTgdcCheck != nil {
		ui.PingTgdcCheck.Refresh()
	}
	if ui.PingWebCheck != nil {
		ui.PingWebCheck.Refresh()
	}
	if ui.ChinaModeCheck != nil {
		ui.ChinaModeCheck.Refresh()
	}
}

// startTests 开始执行测试
func (ui *TestUI) startTests() {
	ui.Mu.Lock()
	if ui.IsRunning {
		ui.Mu.Unlock()
		return
	}
	ui.IsRunning = true
	ui.Mu.Unlock()

	if !ui.hasSelectedTests() {
		dialog.ShowInformation(ui.tr("dialog.hint"), ui.tr("dialog.no_tests"), ui.Window)
		ui.Mu.Lock()
		ui.IsRunning = false
		ui.Mu.Unlock()
		return
	}

	config := ui.collectExecutionConfig()

	// 禁用开始按钮，启用停止按钮
	ui.StartButton.Disable()
	ui.StopButton.Enable()
	ui.ProgressBar.Show()
	ui.setStatus("status.running")

	// 清空终端输出
	if ui.Terminal != nil {
		ui.Terminal.Clear()
	}

	// 创建新的取消上下文
	ui.CancelCtx, ui.CancelFn = context.WithCancel(context.Background())

	// 在新 goroutine 中运行测试
	go ui.runTestsWithExecutor(config)
} // stopTests 停止正在执行的测试
func (ui *TestUI) stopTests() {
	ui.Mu.Lock()
	if !ui.IsRunning {
		ui.Mu.Unlock()
		return
	}

	// 调用取消函数
	if ui.CancelFn != nil {
		ui.CancelFn()
	}
	ui.Mu.Unlock()

	// 更新UI状态
	ui.runOnUI(func() {
		ui.setStatus("status.stopping")
		ui.StopButton.Disable()
	})
	ui.Terminal.AppendText(ui.tr("log.interrupted"))

	// resetUIState 会在 runTestsWithExecutor 的 defer 中调用
}

// clearResults 清空测试结果
func (ui *TestUI) clearResults() {
	if ui.Terminal != nil {
		ui.Terminal.Clear()
	}
	ui.runOnUI(func() {
		ui.setStatus("status.ready")
		ui.ProgressBar.SetValue(0)
	})
}

// copyResults 复制测试结果到剪贴板
func (ui *TestUI) copyResults() {
	var content string
	if ui.Terminal != nil {
		content = ui.Terminal.GetText()
	}

	if content == "" {
		dialog.ShowInformation(ui.tr("dialog.hint"), ui.tr("dialog.no_copy"), ui.Window)
		return
	}

	// 复制到剪贴板
	ui.App.Clipboard().SetContent(content)
	dialog.ShowInformation(ui.tr("dialog.success"), ui.tr("dialog.copy_ok"), ui.Window)
}

// exportResults 导出测试结果
func (ui *TestUI) exportResults() {
	var content string
	if ui.Terminal != nil {
		content = ui.Terminal.GetText()
	}

	if content == "" {
		dialog.ShowInformation(ui.tr("dialog.hint"), ui.tr("dialog.no_export"), ui.Window)
		return
	}

	// 直接导出为文本文件
	// 设置默认文件名
	defaultFilename := "goecs.txt"

	// 创建保存对话框，设置默认文件名
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.Window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		_, err = writer.Write([]byte(content))
		if err != nil {
			dialog.ShowError(err, ui.Window)
			return
		}

		dialog.ShowInformation(ui.tr("dialog.success"), ui.tr("dialog.exported")+writer.URI().Path(), ui.Window)
	}, ui.Window)

	// 设置默认文件名
	saveDialog.SetFileName(defaultFilename)

	// 尝试设置默认位置到用户主目录
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		if lister, err := storage.ListerForURI(storage.NewFileURI(homeDir)); err == nil {
			saveDialog.SetLocation(lister)
		}
	}

	saveDialog.Show()
}

// onLogCheckChanged 当日志复选框状态改变时调用
func (ui *TestUI) onLogCheckChanged(checked bool) {
	if checked {
		// 勾选时添加日志标签页
		ui.addLogTab()
	} else {
		// 取消勾选时移除日志标签页
		ui.removeLogTab()
	}
}

// addLogTab 添加日志标签页
func (ui *TestUI) addLogTab() {
	// 如果日志标签页已存在，不重复添加
	if ui.LogTab != nil {
		return
	}

	// 创建日志查看器
	ui.LogViewer = widget.NewMultiLineEntry()
	ui.LogViewer.SetPlaceHolder(ui.tr("placeholder.log_viewer"))
	ui.LogViewer.Wrapping = fyne.TextWrapWord
	// 不使用 Disable()，让文字颜色保持正常
	// ui.LogViewer.Disable() // 只读

	// 刷新日志按钮
	refreshButton := widget.NewButton(ui.tr("button.log_refresh"), func() {
		ui.refreshLogFromFileAsync()
	})

	// 清空日志按钮
	clearLogButton := widget.NewButton(ui.tr("button.log_clear"), func() {
		ui.LogContent = ""
		ui.LogViewer.SetText("")
	})

	// 导出日志按钮
	exportLogButton := widget.NewButton(ui.tr("button.log_export"), ui.exportLogContent)

	// 按钮栏
	buttonBar := container.NewHBox(
		refreshButton,
		clearLogButton,
		exportLogButton,
	)

	// 日志内容区域
	logScroll := container.NewScroll(ui.LogViewer)

	// 组合布局
	logContent := container.NewBorder(
		buttonBar, // Top: 按钮栏
		nil,       // Bottom
		nil,       // Left
		nil,       // Right
		logScroll, // Center: 日志内容
	)

	// 创建并添加日志标签页
	ui.LogTab = container.NewTabItem(ui.tr("tab.log"), logContent)
	ui.MainTabs.Append(ui.LogTab)

	// 初始化日志内容
	ui.LogContent = ""
}

// removeLogTab 移除日志标签页
func (ui *TestUI) removeLogTab() {
	if ui.LogTab == nil {
		return
	}

	// 从标签页容器中移除
	ui.MainTabs.Remove(ui.LogTab)
	ui.LogTab = nil
	ui.LogViewer = nil
	ui.LogContent = ""
}

// refreshLogContent 刷新日志内容
func (ui *TestUI) refreshLogContent() {
	if ui.LogViewer == nil {
		return
	}

	// 显示存储的日志内容
	if ui.LogContent != "" {
		ui.runOnUI(func() {
			ui.LogViewer.SetText(ui.LogContent)
		})
	} else {
		ui.runOnUI(func() {
			ui.LogViewer.SetText(ui.tr("log.empty"))
		})
	}
}

// refreshLogFromFile 从 ecs.log 文件读取日志内容
func (ui *TestUI) refreshLogFromFile() {
	ui.refreshLogFromFileAsync()
}

func (ui *TestUI) refreshLogFromFileAsync() {
	if ui.LogViewer == nil {
		return
	}

	go func() {
		logFilePath := "ecs.log"
		content, err := tailFileContent(logFilePath, maxLogViewBytes)
		if err != nil {
			if os.IsNotExist(err) {
				ui.runOnUI(func() {
					if ui.LogViewer != nil {
						ui.LogViewer.SetText(ui.tr("log.not_found"))
					}
				})
			} else {
				ui.runOnUI(func() {
					if ui.LogViewer != nil {
						ui.LogViewer.SetText(ui.tr("log.read_failed") + err.Error())
					}
				})
			}
			return
		}

		ui.Mu.Lock()
		ui.LogContent = content
		ui.Mu.Unlock()

		ui.runOnUI(func() {
			if ui.LogViewer != nil {
				ui.LogViewer.SetText(content)
			}
		})
	}()
}

func tailFileContent(path string, maxBytes int64) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", err
	}

	size := info.Size()
	if size == 0 {
		return "", nil
	}

	start := int64(0)
	truncated := false
	if size > maxBytes {
		start = size - maxBytes
		truncated = true
	}

	buf := make([]byte, size-start)
	_, err = f.ReadAt(buf, start)
	if err != nil && err != io.EOF {
		return "", err
	}

	content := string(buf)
	if truncated {
		content = fmt.Sprintf("[log truncated: showing last %d bytes]\n%s", maxBytes, content)
	}
	return content, nil
}

// exportLogContent 导出日志内容
func (ui *TestUI) exportLogContent() {
	if ui.LogViewer == nil || ui.LogViewer.Text == "" {
		dialog.ShowInformation(ui.tr("dialog.hint"), ui.tr("dialog.no_log_export"), ui.Window)
		return
	}

	// 使用文件保存对话框
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.Window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		// 写入日志内容
		_, err = writer.Write([]byte(ui.LogViewer.Text))
		if err != nil {
			dialog.ShowError(err, ui.Window)
			return
		}

		dialog.ShowInformation(ui.tr("dialog.success"), ui.tr("dialog.log_export_ok"), ui.Window)
	}, ui.Window)
}

// AppendLog 向日志内容追加文本
func (ui *TestUI) AppendLog(text string) {
	if !ui.LogCheck.Checked || ui.LogViewer == nil {
		return
	}

	ui.Mu.Lock()
	defer ui.Mu.Unlock()

	ui.LogContent += text
	content := ui.LogContent
	ui.runOnUI(func() {
		if ui.LogViewer != nil {
			ui.LogViewer.SetText(content)
		}
	})
}
