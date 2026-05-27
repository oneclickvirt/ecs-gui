package ui

import (
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewTestUI 创建新的测试UI实例
func NewTestUI(app fyne.App) *TestUI {
	ui := &TestUI{
		App:    app,
		uiLang: langZH,
		Window: app.NewWindow(""),
	}
	ui.Window.SetTitle(ui.tr("app.title"))

	// 移动端使用系统默认窗口行为，桌面端提供较舒适的初始尺寸
	if runtime.GOOS != "android" && runtime.GOOS != "ios" {
		ui.Window.Resize(fyne.NewSize(980, 820))
	}
	ui.Window.SetPadded(true)
	ui.Window.CenterOnScreen()

	ui.buildUI()

	// 设置窗口关闭时的清理操作
	ui.Window.SetOnClosed(func() {
		// 如果测试正在运行，取消它
		if ui.CancelFn != nil {
			ui.CancelFn()
		}

		// 清理 Terminal 资源
		if ui.Terminal != nil {
			ui.Terminal.Destroy()
		}
	})

	return ui
}

// buildUI 构建用户界面 - 使用Tab切换页面
func (ui *TestUI) buildUI() {
	// 创建终端输出组件
	ui.Terminal = NewTerminalOutput()

	// 创建状态栏
	ui.StatusLabel = widget.NewLabel(ui.tr("status.ready"))
	ui.StatusBadge = widget.NewLabel(ui.tr("badge.ready"))
	ui.ProgressBar = widget.NewProgressBar()
	ui.ProgressBar.Hide()

	// 创建Tab页面
	configTab := container.NewTabItem(ui.tr("tab.config"), ui.createConfigTab())
	resultTab := container.NewTabItem(ui.tr("tab.result"), ui.createResultTab())
	ui.MainTabs = container.NewAppTabs(
		configTab,
		resultTab,
	)

	ui.Window.SetContent(ui.MainTabs)
}

// createConfigTab 创建测试选项与配置页面（支持滚动）
func (ui *TestUI) createConfigTab() fyne.CanvasObject {
	// 创建选项面板内容
	optionsContent := ui.createOptionsPanel()

	// 创建控制按钮区域
	controlButtons := ui.createControlButtons()

	// 将选项放在滚动容器中
	scrollContent := container.NewScroll(container.NewPadded(optionsContent))

	// 使用Border布局，控制按钮固定在底部
	return container.NewBorder(
		nil,            // Top
		controlButtons, // Bottom: 控制按钮固定在底部
		nil,            // Left
		nil,            // Right
		scrollContent,  // Center: 可滚动的选项内容
	)
}

// createResultTab 创建测试结果页面
func (ui *TestUI) createResultTab() fyne.CanvasObject {
	statusRow := container.NewHBox(
		ui.StatusLabel,
		layout.NewSpacer(),
		ui.StatusBadge,
	)
	statusBar := container.NewVBox(statusRow, ui.ProgressBar)

	copyButton := widget.NewButtonWithIcon(ui.tr("button.copy"), theme.ContentCopyIcon(), ui.copyResults)
	exportButton := widget.NewButtonWithIcon(ui.tr("button.export"), theme.DownloadIcon(), ui.exportResults)
	clearButton := widget.NewButtonWithIcon(ui.tr("button.clear"), theme.DeleteIcon(), ui.clearResults)

	actions := []fyne.CanvasObject{clearButton, copyButton, exportButton}
	actionsBar := container.NewHBox(actions...)
	if isMobilePlatform() {
		actionsBar = container.NewAdaptiveGrid(2, actions...)
	} else {
		actionsBar = container.NewHBox(layout.NewSpacer(), clearButton, copyButton, exportButton)
	}

	header := widget.NewCard("", "", container.NewVBox(
		statusBar,
		actionsBar,
	))

	terminalScroll := container.NewScroll(container.NewPadded(ui.Terminal))

	return container.NewBorder(
		header,         // Top: 状态栏和操作按钮
		nil,            // Bottom
		nil,            // Left
		nil,            // Right
		terminalScroll, // Center: 终端输出
	)
}

// createControlButtons 创建控制按钮
func (ui *TestUI) createControlButtons() fyne.CanvasObject {
	ui.StartButton = widget.NewButton(ui.tr("button.start"), ui.startTests)
	ui.StartButton.Importance = widget.HighImportance

	ui.StopButton = widget.NewButton(ui.tr("button.stop"), ui.stopTests)
	ui.StopButton.Disable()
	ui.StopButton.Importance = widget.MediumImportance

	if isMobilePlatform() {
		return container.NewPadded(container.NewVBox(ui.StartButton, ui.StopButton))
	}

	return container.NewPadded(container.NewCenter(
		container.NewHBox(
			ui.StartButton,
			ui.StopButton,
		),
	))
}
