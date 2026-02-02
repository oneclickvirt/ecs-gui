package ui

import (
	"fmt"
	"time"
)

// runTestsWithExecutor 使用命令执行器运行测试
func (ui *TestUI) runTestsWithExecutor() {
	defer func() {
		if r := recover(); r != nil {
			ui.Terminal.AppendText(fmt.Sprintf("\n错误: %v\n", r))
		}
		ui.resetUIState()
	}()

	startTime := time.Now()

	// 创建命令执行器
	executor := NewCommandExecutor(func(text string) {
		ui.Terminal.AppendText(text)
	})

	// 获取选择的测试选项
	selectedOptions := ui.GetSelectedOptions()

	// 获取选择的语言
	language := "zh"
	if ui.LanguageSelect.Selected == "English" {
		language = "en"
	}

	// 获取速度测试配置
	testUpload := ui.SpTestUploadCheck.Checked
	testDownload := ui.SpTestDownloadCheck.Checked

	// 获取中国模式配置
	chinaModeEnabled := ui.ChinaModeCheck.Checked

	// 获取CPU配置
	cpuMethod := ui.CpuMethodSelect.Selected
	if cpuMethod == "" {
		cpuMethod = "sysbench"
	}
	threadMode := ui.ThreadModeSelect.Selected
	if threadMode == "" {
		threadMode = "multi"
	}

	// 获取内存配置
	memoryMethod := ui.MemoryMethodSelect.Selected
	if memoryMethod == "" {
		memoryMethod = "auto"
	}

	// 获取磁盘配置
	diskMethod := ui.DiskMethodSelect.Selected
	if diskMethod == "" {
		diskMethod = "auto"
	}
	diskPath := ui.DiskPathEntry.Text
	diskMulti := ui.DiskMultiCheck.Checked

	// 获取NT3配置
	nt3Location := ui.Nt3LocationSelect.Selected
	if nt3Location == "" {
		nt3Location = "GZ"
	}
	nt3Type := ui.Nt3TypeSelect.Selected
	if nt3Type == "" {
		nt3Type = "ipv4"
	}

	// 获取测速节点数
	spNum := 2 // 默认值
	if ui.SpNumEntry.Text != "" {
		// 尝试解析spNum，如果失败则使用默认值
		fmt.Sscanf(ui.SpNumEntry.Text, "%d", &spNum)
	}

	// 获取PING测试配置
	pingTgdc := ui.PingTgdcCheck.Checked
	pingWeb := ui.PingWebCheck.Checked

	// 更新进度
	ui.ProgressBar.SetValue(0.1)
	ui.StatusLabel.SetText("正在执行测试...")

	// 执行测试（输出会实时显示在terminal widget中）
	err := executor.Execute(selectedOptions, language, testUpload, testDownload, chinaModeEnabled,
		cpuMethod, threadMode, memoryMethod, diskMethod, diskPath, diskMulti,
		nt3Location, nt3Type, spNum, pingTgdc, pingWeb)

	// 显示结束信息
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	_ = duration // 避免未使用警告

	if err != nil {
		ui.Terminal.AppendText(fmt.Sprintf("\n错误: %v\n", err))
		ui.StatusLabel.SetText("测试失败")
	} else if ui.isCancelled() {
		ui.Terminal.AppendText("\n测试被用户中断\n")
		ui.StatusLabel.SetText("测试已停止")
	} else {
		ui.StatusLabel.SetText("测试完成")
		ui.ProgressBar.SetValue(1.0)

		// 如果启用了日志，自动刷新日志内容
		if ui.LogCheck != nil && ui.LogCheck.Checked {
			time.Sleep(500 * time.Millisecond) // 等待日志文件写入完成
			ui.refreshLogFromFile()
		}
	}
}
