package ui

import (
	"net/url"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewTestUI 创建新的测试UI实例
func NewTestUI(app fyne.App) *TestUI {
	themeMode := normalizeThemeMode(app.Preferences().StringWithFallback(themePreferenceKey, themeModeLight))
	ui := &TestUI{
		App:       app,
		uiLang:    langZH,
		themeMode: themeMode,
		Window:    app.NewWindow(""),
	}
	ui.applyThemeMode(themeMode)
	ui.Window.SetTitle(ui.tr("app.title"))

	// 移动端使用系统默认窗口行为，桌面端提供较舒适的初始尺寸
	if runtime.GOOS != "android" && runtime.GOOS != "ios" {
		ui.Window.Resize(fyne.NewSize(980, 820))
	}
	ui.Window.SetPadded(true)
	ui.Window.CenterOnScreen()

	ui.buildUI()
	ui.registerLifecycleHooks()

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

func (ui *TestUI) registerLifecycleHooks() {
	if ui.App == nil {
		return
	}
	ui.App.Lifecycle().SetOnExitedForeground(func() {
		ui.Mu.Lock()
		ui.inBackground = true
		ui.Mu.Unlock()
	})
	ui.App.Lifecycle().SetOnEnteredForeground(func() {
		ui.Mu.Lock()
		ui.inBackground = false
		ui.Mu.Unlock()
	})
}

// buildUI 构建用户界面 - 使用Tab切换页面
func (ui *TestUI) buildUI() {
	// 创建终端输出组件
	ui.Terminal = NewTerminalOutput()

	// 创建状态栏
	ui.StatusLabel = widget.NewLabel(ui.tr("status.ready"))
	ui.StatusBadge = widget.NewLabel(ui.tr("badge.ready"))
	ui.CurrentItem = widget.NewLabel(ui.tr("progress.idle"))
	ui.ProgressBar = widget.NewProgressBar()
	ui.ProgressBar.Hide()

	// 创建Tab页面
	launchTab := container.NewTabItem(ui.tr("tab.launch"), ui.createLaunchTab())
	configTab := container.NewTabItem(ui.tr("tab.config"), ui.createConfigTab())
	resultTab := container.NewTabItem(ui.tr("tab.result"), ui.createResultTab())
	ui.MainTabs = container.NewAppTabs(
		launchTab,
		configTab,
		resultTab,
	)

	ui.Window.SetContent(ui.createRootContent())
}

func (ui *TestUI) createRootContent() fyne.CanvasObject {
	return container.NewBorder(nil, ui.createFooter(), nil, nil, ui.MainTabs)
}

func (ui *TestUI) createFooter() fyne.CanvasObject {
	link := func(label, rawURL string) *widget.Hyperlink {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			parsed = &url.URL{Scheme: "https", Host: "github.com"}
		}
		return widget.NewHyperlink(label, parsed)
	}

	links := []fyne.CanvasObject{
		link(ui.tr("footer.gui"), "https://github.com/oneclickvirt/ecs-gui"),
		link(ui.tr("footer.upstream"), "https://github.com/oneclickvirt/ecs"),
		link(ui.tr("footer.guide"), "https://bash.spiritlhl.net/ecsguide"),
	}
	if isMobilePlatform() {
		return container.NewPadded(container.NewAdaptiveGrid(3, links...))
	}
	return container.NewPadded(container.NewHBox(layout.NewSpacer(), links[0], links[1], links[2]))
}

func (ui *TestUI) showConfigTab() {
	if ui.MainTabs != nil && len(ui.MainTabs.Items) > 1 {
		ui.MainTabs.SelectIndex(1)
	}
}

func (ui *TestUI) showResultTab() {
	if ui.MainTabs != nil && len(ui.MainTabs.Items) > 2 {
		ui.MainTabs.SelectIndex(2)
	}
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

func (ui *TestUI) createLaunchTab() fyne.CanvasObject {
	presetButton := func(label string, icon fyne.Resource, presetKey string) fyne.CanvasObject {
		return widget.NewButtonWithIcon(label, icon, func() {
			ui.applyPresetAndStart(presetKey)
		})
	}
	singleButton := func(label string, icon fyne.Resource, keys ...string) fyne.CanvasObject {
		return widget.NewButtonWithIcon(label, icon, func() {
			ui.applySingleSelection(keys...)
			ui.startTests()
			ui.showResultTab()
		})
	}

	presets := container.NewAdaptiveGrid(optionGridColumns(),
		presetButton(ui.tr("button.start_standard"), theme.MediaPlayIcon(), "standard"),
		presetButton(ui.tr("button.start_full"), theme.ViewFullScreenIcon(), "full"),
		presetButton(ui.tr("preset.network_only"), theme.SearchIcon(), "network_only"),
		presetButton(ui.tr("preset.hardware_only"), theme.ComputerIcon(), "hardware_only"),
		presetButton(ui.tr("preset.unlock_only"), theme.InfoIcon(), "unlock_only"),
		presetButton(ui.tr("preset.route_only"), theme.NavigateNextIcon(), "route_only"),
	)

	singles := container.NewAdaptiveGrid(optionGridColumns(),
		singleButton(ui.tr("single.basic"), theme.SettingsIcon(), "basic"),
		singleButton(ui.tr("single.cpu"), theme.ComputerIcon(), "cpu"),
		singleButton(ui.tr("single.memory"), theme.StorageIcon(), "memory"),
		singleButton(ui.tr("single.disk"), theme.FolderIcon(), "disk"),
		singleButton(ui.tr("single.unlock"), theme.InfoIcon(), "unlock"),
		singleButton(ui.tr("single.security"), theme.VisibilityIcon(), "security"),
		singleButton(ui.tr("single.email"), theme.MailComposeIcon(), "email"),
		singleButton(ui.tr("single.backtrace"), theme.SearchIcon(), "backtrace"),
		singleButton(ui.tr("single.nt3"), theme.NavigateNextIcon(), "nt3"),
		singleButton(ui.tr("single.speed"), theme.DownloadIcon(), "speed"),
		singleButton(ui.tr("single.ping"), theme.ViewRefreshIcon(), "ping"),
		singleButton(ui.tr("single.tgdc"), theme.UploadIcon(), "tgdc"),
		singleButton(ui.tr("single.web"), theme.HomeIcon(), "web"),
	)

	manage := container.NewAdaptiveGrid(optionGridColumns(),
		widget.NewButtonWithIcon(ui.tr("button.open_config"), theme.SettingsIcon(), ui.showConfigTab),
		widget.NewButtonWithIcon(ui.tr("tab.result"), theme.DocumentIcon(), ui.showResultTab),
	)

	content := container.NewVBox(
		widget.NewCard(ui.tr("launch.card.title"), ui.tr("launch.card.sub"), container.NewVBox(
			widget.NewLabelWithStyle(ui.tr("launch.presets"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			presets,
			widget.NewSeparator(),
			widget.NewLabelWithStyle(ui.tr("launch.single"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			singles,
			widget.NewSeparator(),
			widget.NewLabelWithStyle(ui.tr("launch.manage"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			manage,
		)),
	)

	return container.NewScroll(container.NewPadded(content))
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
