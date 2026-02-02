package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NewTestUI 创建新的测试UI实例
func NewTestUI(app fyne.App) *TestUI {
	ui := &TestUI{
		App:    app,
		Window: app.NewWindow("融合怪测试 - 完整版"),
	}

	// 设置窗口大小
	ui.Window.Resize(fyne.NewSize(900, 800))
	ui.Window.SetPadded(true)
	ui.Window.CenterOnScreen()

	ui.buildUI()
	return ui
}

// buildUI 构建用户界面 - 使用Tab切换页面
func (ui *TestUI) buildUI() {
	// 创建终端输出组件
	ui.Terminal = NewTerminalOutput()

	// 创建状态栏
	ui.StatusLabel = widget.NewLabel("就绪")
	ui.ProgressBar = widget.NewProgressBar()
	ui.ProgressBar.Hide()

	// 创建Tab页面
	ui.MainTabs = container.NewAppTabs(
		container.NewTabItem("测试选项与配置", ui.createConfigTab()),
		container.NewTabItem("测试结果", ui.createResultTab()),
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
	scrollContent := container.NewScroll(optionsContent)

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
	// 状态栏
	statusBar := container.NewBorder(
		nil, nil,
		ui.StatusLabel,
		nil,
		ui.ProgressBar,
	)

	// 导出按钮
	copyButton := widget.NewButton("复制", ui.copyResults)
	exportButton := widget.NewButton("导出", ui.exportResults)
	clearButton := widget.NewButton("清空", ui.clearResults)

	topBar := container.NewBorder(
		nil, nil,
		statusBar,
		container.NewHBox(clearButton, copyButton, exportButton),
	)

	// 终端输出占据主要空间
	terminalScroll := container.NewScroll(ui.Terminal)

	return container.NewBorder(
		topBar,         // Top: 状态栏和操作按钮
		nil,            // Bottom
		nil,            // Left
		nil,            // Right
		terminalScroll, // Center: 终端输出
	)
}

// createControlButtons 创建控制按钮
func (ui *TestUI) createControlButtons() fyne.CanvasObject {
	ui.StartButton = widget.NewButton("开始测试", ui.startTests)
	ui.StartButton.Importance = widget.HighImportance

	ui.StopButton = widget.NewButton("停止测试", ui.stopTests)
	ui.StopButton.Disable()

	return container.NewCenter(
		container.NewHBox(
			ui.StartButton,
			ui.StopButton,
		),
	)
}
