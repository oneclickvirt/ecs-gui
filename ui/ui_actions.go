package ui

import (
	"context"
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// onPresetChanged 预设模式改变时的处理
func (ui *TestUI) onPresetChanged(preset string) {
	switch preset {
	case "1. 融合怪完全体(能测全测)":
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
	case "2. 极简版(系统信息+CPU+内存+磁盘+测速节点5个)":
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
	case "3. 精简版(系统信息+CPU+内存+磁盘+常用流媒体+路由+测速节点5个)":
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
	case "4. 精简网络版(系统信息+CPU+内存+磁盘+回程+路由+测速节点5个)":
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
	case "5. 精简解锁版(系统信息+CPU+内存+磁盘IO+御三家+常用流媒体+测速节点5个)":
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
	case "6. 网络单项(IP质量检测+上游及三网回程+广州三网回程详细路由+全国延迟+TGDC+网站延迟+测速节点11个)":
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
	case "7. 解锁单项(御三家解锁+常用流媒体解锁)":
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
	case "8. 硬件单项(系统信息+CPU+dd磁盘测试+fio磁盘测试)":
		// 对应原goecs.go的选项8: SetHardwareOnlyTestStatus
		ui.setAllChecks(false)
		ui.BasicCheck.Checked = true
		ui.CpuCheck.Checked = true
		ui.MemoryCheck.Checked = true
		ui.DiskCheck.Checked = true
		ui.DiskMethodSelect.Selected = "auto" // 使用auto让系统自动选择dd和fio
		ui.PingTgdcCheck.Checked = false
		ui.PingWebCheck.Checked = false
		ui.ChinaModeCheck.Checked = false
		// 测速配置：禁用
		ui.SpTestUploadCheck.Checked = false
		ui.SpTestDownloadCheck.Checked = false
	case "9. IP质量检测(15个数据库的IP质量检测+邮件端口检测)":
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
	case "10. 三网回程线路检测+三网回程详细路由(北京上海广州成都)+全国延迟+TGDC+网站延迟":
		// 对应原goecs.go的选项10: SetRouteTestStatus + nt3Location = "ALL"
		ui.setAllChecks(false)
		ui.BasicCheck.Checked = false // 10号选项不包括基础信息
		ui.BacktraceCheck.Checked = true
		ui.Nt3Check.Checked = true
		ui.PingCheck.Checked = true
		ui.PingTgdcCheck.Checked = true
		ui.PingWebCheck.Checked = true
		ui.Nt3LocationSelect.Selected = "ALL" // 设置为全部地点
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
	ui.BasicCheck.Checked = checked
	ui.CpuCheck.Checked = checked
	ui.MemoryCheck.Checked = checked
	ui.DiskCheck.Checked = checked
	ui.CommCheck.Checked = checked
	ui.UnlockCheck.Checked = checked
	ui.SecurityCheck.Checked = checked
	ui.EmailCheck.Checked = checked
	ui.BacktraceCheck.Checked = checked
	ui.Nt3Check.Checked = checked
	ui.SpeedCheck.Checked = checked
	ui.PingCheck.Checked = checked
	ui.refreshAllChecks()
}

// refreshAllChecks 刷新所有测试项的显示
func (ui *TestUI) refreshAllChecks() {
	ui.BasicCheck.Refresh()
	ui.CpuCheck.Refresh()
	ui.MemoryCheck.Refresh()
	ui.DiskCheck.Refresh()
	ui.CommCheck.Refresh()
	ui.UnlockCheck.Refresh()
	ui.SecurityCheck.Refresh()
	ui.EmailCheck.Refresh()
	ui.BacktraceCheck.Refresh()
	ui.Nt3Check.Refresh()
	ui.SpeedCheck.Refresh()
	ui.PingCheck.Refresh()
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
		dialog.ShowInformation("提示", "请至少选择一项测试！", ui.Window)
		ui.Mu.Lock()
		ui.IsRunning = false
		ui.Mu.Unlock()
		return
	}

	ui.StartButton.Disable()
	ui.StopButton.Enable()
	ui.ProgressBar.Show()
	ui.StatusLabel.SetText("测试运行中...")

	// 清空终端输出
	if ui.Terminal != nil {
		ui.Terminal.Clear()
	}

	ui.CancelCtx, ui.CancelFn = context.WithCancel(context.Background())
	go ui.runTestsWithExecutor()
} // stopTests 停止正在执行的测试
func (ui *TestUI) stopTests() {
	ui.Mu.Lock()
	defer ui.Mu.Unlock()

	if ui.CancelFn != nil {
		ui.CancelFn()
	}
	ui.StatusLabel.SetText("测试已停止")
	ui.Terminal.AppendText("\n\n========== 测试被用户中断 ==========\n")
	ui.resetUIState()
}

// clearResults 清空测试结果
func (ui *TestUI) clearResults() {
	if ui.Terminal != nil {
		ui.Terminal.Clear()
	}
	ui.StatusLabel.SetText("就绪")
	ui.ProgressBar.SetValue(0)
}

// copyResults 复制测试结果到剪贴板
func (ui *TestUI) copyResults() {
	var content string
	if ui.Terminal != nil {
		content = ui.Terminal.GetText()
	}

	if content == "" {
		dialog.ShowInformation("提示", "没有可复制的内容", ui.Window)
		return
	}

	// 复制到剪贴板
	ui.App.Clipboard().SetContent(content)
	dialog.ShowInformation("成功", "测试结果已复制到剪贴板", ui.Window)
}

// exportResults 导出测试结果
func (ui *TestUI) exportResults() {
	var content string
	if ui.Terminal != nil {
		content = ui.Terminal.GetText()
	}

	if content == "" {
		dialog.ShowInformation("提示", "没有可导出的结果", ui.Window)
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

		dialog.ShowInformation("成功", "结果已导出到: "+writer.URI().Path(), ui.Window)
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
	ui.LogViewer.SetPlaceHolder("日志内容将在测试运行时显示...")
	ui.LogViewer.Wrapping = fyne.TextWrapWord
	// 不使用 Disable()，让文字颜色保持正常
	// ui.LogViewer.Disable() // 只读

	// 刷新日志按钮
	refreshButton := widget.NewButton("刷新日志", func() {
		ui.refreshLogFromFile()
	})

	// 清空日志按钮
	clearLogButton := widget.NewButton("清空日志", func() {
		ui.LogContent = ""
		ui.LogViewer.SetText("")
	})

	// 导出日志按钮
	exportLogButton := widget.NewButton("导出日志", ui.exportLogContent)

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
	ui.LogTab = container.NewTabItem("日志", logContent)
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
		ui.LogViewer.SetText(ui.LogContent)
	} else {
		ui.LogViewer.SetText("暂无日志内容\n\n日志将在测试运行时自动更新。")
	}
}

// refreshLogFromFile 从 ecs.log 文件读取日志内容
func (ui *TestUI) refreshLogFromFile() {
	if ui.LogViewer == nil {
		return
	}

	// ecs.log 文件应该在当前工作目录下
	logFilePath := "ecs.log"

	// 尝试读取日志文件
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		// 如果文件不存在或无法读取，显示错误信息
		if os.IsNotExist(err) {
			ui.LogViewer.SetText("日志文件 ecs.log 不存在\n\n可能测试未生成日志文件，或文件已被删除。")
		} else {
			ui.LogViewer.SetText(fmt.Sprintf("无法读取日志文件: %v", err))
		}
		return
	}

	// 更新日志内容
	ui.LogContent = string(content)
	ui.LogViewer.SetText(ui.LogContent)
}

// exportLogContent 导出日志内容
func (ui *TestUI) exportLogContent() {
	if ui.LogViewer == nil || ui.LogViewer.Text == "" {
		dialog.ShowInformation("提示", "没有可导出的日志内容", ui.Window)
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

		dialog.ShowInformation("成功", "日志已成功导出", ui.Window)
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
	ui.LogViewer.SetText(ui.LogContent)
}
