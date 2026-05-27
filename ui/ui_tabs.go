package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (ui *TestUI) newIconCard(title, subtitle string, icon fyne.Resource, body fyne.CanvasObject) fyne.CanvasObject {
	head := container.NewHBox(
		widget.NewIcon(icon),
		widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)
	return widget.NewCard("", "", container.NewVBox(
		head,
		widget.NewLabel(subtitle),
		layout.NewSpacer(),
		body,
	))
}

// createOptionsPanel 创建选项面板（测试项目 + 配置选项整合在一起）
func (ui *TestUI) createOptionsPanel() fyne.CanvasObject {
	if ui.selectedPresetKey == "" {
		ui.selectedPresetKey = "custom"
	}

	// 预设模式选择
	presetOptions := ui.rebuildPresetMappings()
	ui.PresetSelect = widget.NewSelect(
		presetOptions,
		ui.onPresetChanged,
	)
	ui.suppressPresetChange = true
	ui.PresetSelect.SetSelected(ui.presetLabelByKey(ui.selectedPresetKey))
	ui.suppressPresetChange = false

	presetSection := widget.NewCard(ui.tr("preset.card.title"), ui.tr("preset.card.subtitle"), ui.PresetSelect)

	// === 测试项目复选框 ===
	ui.BasicCheck = widget.NewCheck(ui.tr("check.basic"), nil)
	ui.BasicCheck.Checked = true

	ui.CpuCheck = widget.NewCheck(ui.tr("check.cpu"), nil)
	ui.CpuCheck.Checked = true

	ui.MemoryCheck = widget.NewCheck(ui.tr("check.memory"), nil)
	ui.MemoryCheck.Checked = true

	ui.DiskCheck = widget.NewCheck(ui.tr("check.disk"), nil)
	ui.DiskCheck.Checked = true

	ui.CommCheck = widget.NewCheck(ui.tr("check.comm"), nil)
	ui.CommCheck.Checked = false

	ui.UnlockCheck = widget.NewCheck(ui.tr("check.unlock"), nil)
	ui.UnlockCheck.Checked = false

	ui.SecurityCheck = widget.NewCheck(ui.tr("check.security"), nil)
	ui.SecurityCheck.Checked = false

	ui.EmailCheck = widget.NewCheck(ui.tr("check.email"), nil)
	ui.EmailCheck.Checked = false

	ui.BacktraceCheck = widget.NewCheck(ui.tr("check.backtrace"), nil)
	ui.BacktraceCheck.Checked = false

	ui.Nt3Check = widget.NewCheck(ui.tr("check.nt3"), nil)
	ui.Nt3Check.Checked = false

	ui.SpeedCheck = widget.NewCheck(ui.tr("check.speed"), nil)
	ui.SpeedCheck.Checked = false

	ui.PingCheck = widget.NewCheck(ui.tr("check.ping"), nil)
	ui.PingCheck.Checked = false

	ui.LogCheck = widget.NewCheck(ui.tr("check.log"), ui.onLogCheckChanged)
	ui.LogCheck.Checked = false

	ui.testChecks = []*widget.Check{
		ui.BasicCheck,
		ui.CpuCheck,
		ui.MemoryCheck,
		ui.DiskCheck,
		ui.CommCheck,
		ui.UnlockCheck,
		ui.SecurityCheck,
		ui.EmailCheck,
		ui.BacktraceCheck,
		ui.Nt3Check,
		ui.SpeedCheck,
		ui.PingCheck,
	}

	// 全选/取消全选按钮
	selectAllBtn := widget.NewButton(ui.tr("button.select_all"), func() {
		ui.setAllChecks(true)
	})

	deselectAllBtn := widget.NewButton(ui.tr("button.deselect_all"), func() {
		ui.setAllChecks(false)
	})

	buttonRow := container.NewHBox(selectAllBtn, deselectAllBtn)

	// 测试项目分组
	basicTests := ui.newIconCard(ui.tr("tests.basic.title"), ui.tr("tests.basic.sub"), theme.SettingsIcon(), container.NewVBox(
		ui.BasicCheck,
		ui.CpuCheck,
		ui.MemoryCheck,
		ui.DiskCheck,
	))

	networkTests := ui.newIconCard(ui.tr("tests.network.title"), ui.tr("tests.network.sub"), theme.SearchIcon(), container.NewVBox(
		ui.SpeedCheck,
		ui.SecurityCheck,
		ui.EmailCheck,
		ui.BacktraceCheck,
		ui.Nt3Check,
		ui.PingCheck,
	))

	unlockTests := ui.newIconCard(ui.tr("tests.unlock.title"), ui.tr("tests.unlock.sub"), theme.InfoIcon(), container.NewVBox(
		ui.CommCheck,
		ui.UnlockCheck,
	))

	testsGrid := container.NewAdaptiveGrid(optionGridColumns(),
		basicTests,
		networkTests,
		unlockTests,
	)

	testsSection := widget.NewCard(ui.tr("tests.card.title"), ui.tr("tests.card.subtitle"), container.NewVBox(
		buttonRow,
		layout.NewSpacer(),
		testsGrid,
	))

	// === 配置选项 ===
	configSection := ui.createConfigSection()

	// 整合所有内容
	allContent := container.NewVBox(
		widget.NewCard("", "", presetSection),
		widget.NewSeparator(),
		testsSection,
		widget.NewSeparator(),
		configSection,
	)

	return allContent
}

// createConfigSection 创建配置选项区域
func (ui *TestUI) createConfigSection() fyne.CanvasObject {
	// 语言选择
	ui.LanguageSelect = widget.NewSelect(
		[]string{"中文", "English"},
		func(value string) {
			lang := langZH
			if value == "English" {
				lang = langEN
			}

			if lang == ui.uiLang {
				return
			}

			if ui.isRunning() {
				dialog.ShowInformation(ui.tr("dialog.hint"), ui.tr("dialog.running_no_switch"), ui.Window)
				if ui.uiLang == langEN {
					ui.LanguageSelect.SetSelected("English")
				} else {
					ui.LanguageSelect.SetSelected("中文")
				}
				return
			}

			state := ui.snapshotUIState()
			if ui.Terminal != nil {
				ui.Terminal.Destroy()
			}

			ui.uiLang = lang
			ui.buildUI()
			ui.restoreUIState(state)
			ui.Window.SetContent(ui.MainTabs)
			ui.Window.SetTitle(ui.tr("app.title"))
		},
	)
	if ui.uiLang == langEN {
		ui.LanguageSelect.SetSelected("English")
	} else {
		ui.LanguageSelect.SetSelected("中文")
	}

	// CPU 配置
	ui.CpuMethodSelect = widget.NewSelect(
		[]string{"sysbench", "geekbench", "winsat"},
		func(value string) {},
	)
	ui.CpuMethodSelect.Selected = "sysbench"

	ui.ThreadModeSelect = widget.NewSelect(
		[]string{"single", "multi"},
		func(value string) {},
	)
	ui.ThreadModeSelect.Selected = "multi"

	// 内存配置
	ui.MemoryMethodSelect = widget.NewSelect(
		[]string{"auto", "stream", "sysbench", "dd", "winsat"},
		func(value string) {},
	)
	ui.MemoryMethodSelect.Selected = "auto"

	// 磁盘配置
	ui.DiskMethodSelect = widget.NewSelect(
		[]string{"auto", "fio", "dd", "winsat"},
		func(value string) {},
	)
	ui.DiskMethodSelect.Selected = "auto"

	ui.DiskPathEntry = widget.NewEntry()
	ui.DiskPathEntry.SetPlaceHolder(ui.tr("placeholder.disk_path"))

	ui.DiskMultiCheck = widget.NewCheck(ui.tr("check.disk_multi"), nil)
	ui.DiskMultiCheck.Checked = false

	// NT3 配置
	ui.Nt3LocationSelect = widget.NewSelect(
		[]string{"GZ", "SH", "BJ", "CD", "ALL"},
		func(value string) {},
	)
	ui.Nt3LocationSelect.Selected = "GZ"

	ui.Nt3TypeSelect = widget.NewSelect(
		[]string{"ipv4", "ipv6", "both"},
		func(value string) {},
	)
	ui.Nt3TypeSelect.Selected = "ipv4"

	// 测速配置
	ui.SpNumEntry = widget.NewEntry()
	ui.SpNumEntry.SetText("2")
	ui.SpNumEntry.SetPlaceHolder(ui.tr("placeholder.sp_num"))

	// 速度测试上传下载控制
	ui.SpTestUploadCheck = widget.NewCheck(ui.tr("check.sp_up"), nil)
	ui.SpTestUploadCheck.Checked = true

	ui.SpTestDownloadCheck = widget.NewCheck(ui.tr("check.sp_down"), nil)
	ui.SpTestDownloadCheck.Checked = true

	// 中国模式
	ui.ChinaModeCheck = widget.NewCheck(ui.tr("check.china_mode"), nil)
	ui.ChinaModeCheck.Checked = false

	// PING测试配置
	ui.PingTgdcCheck = widget.NewCheck(ui.tr("check.ping_tgdc"), nil)
	ui.PingTgdcCheck.Checked = false

	ui.PingWebCheck = widget.NewCheck(ui.tr("check.ping_web"), nil)
	ui.PingWebCheck.Checked = false

	generalContent := container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabel(ui.tr("label.language")),
			ui.LanguageSelect,
		),
		ui.LogCheck,
	)

	chinaContent := container.NewVBox(
		ui.ChinaModeCheck,
	)

	cpuContent := container.NewGridWithColumns(2,
		widget.NewLabel(ui.tr("label.cpu_method")),
		ui.CpuMethodSelect,
		widget.NewLabel(ui.tr("label.thread_mode")),
		ui.ThreadModeSelect,
	)

	memoryContent := container.NewGridWithColumns(2,
		widget.NewLabel(ui.tr("label.cpu_method")),
		ui.MemoryMethodSelect,
	)

	diskContent := container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabel(ui.tr("label.cpu_method")),
			ui.DiskMethodSelect,
			widget.NewLabel(ui.tr("label.disk_path")),
			ui.DiskPathEntry,
		),
		ui.DiskMultiCheck,
	)

	routeContent := container.NewGridWithColumns(2,
		widget.NewLabel(ui.tr("label.nt3_location")),
		ui.Nt3LocationSelect,
		widget.NewLabel(ui.tr("label.nt3_type")),
		ui.Nt3TypeSelect,
	)

	speedContent := container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabel(ui.tr("label.sp_num")),
			ui.SpNumEntry,
		),
		ui.SpTestUploadCheck,
		ui.SpTestDownloadCheck,
	)

	pingContent := container.NewVBox(
		ui.PingTgdcCheck,
		ui.PingWebCheck,
	)

	if isMobilePlatform() {
		acc := widget.NewAccordion(
			widget.NewAccordionItem(ui.tr("config.general.title"), generalContent),
			widget.NewAccordionItem(ui.tr("config.china.title"), chinaContent),
			widget.NewAccordionItem(ui.tr("config.cpu.title"), cpuContent),
			widget.NewAccordionItem(ui.tr("config.mem.title"), memoryContent),
			widget.NewAccordionItem(ui.tr("config.disk.title"), diskContent),
			widget.NewAccordionItem(ui.tr("config.route.title"), routeContent),
			widget.NewAccordionItem(ui.tr("config.speed.title"), speedContent),
			widget.NewAccordionItem(ui.tr("config.ping.title"), pingContent),
		)
		acc.MultiOpen = false
		acc.Open(0)
		return widget.NewCard(ui.tr("config.card.title"), ui.tr("config.card.sub"), acc)
	}

	generalCard := ui.newIconCard(ui.tr("config.general.title"), ui.tr("config.general.sub"), theme.SettingsIcon(), generalContent)
	chinaCard := ui.newIconCard(ui.tr("config.china.title"), ui.tr("config.china.sub"), theme.InfoIcon(), chinaContent)
	cpuCard := ui.newIconCard(ui.tr("config.cpu.title"), ui.tr("config.cpu.sub"), theme.SettingsIcon(), cpuContent)
	memoryCard := ui.newIconCard(ui.tr("config.mem.title"), ui.tr("config.mem.sub"), theme.SettingsIcon(), memoryContent)
	diskCard := ui.newIconCard(ui.tr("config.disk.title"), ui.tr("config.disk.sub"), theme.StorageIcon(), diskContent)
	routeCard := ui.newIconCard(ui.tr("config.route.title"), ui.tr("config.route.sub"), theme.SearchIcon(), routeContent)
	speedCard := ui.newIconCard(ui.tr("config.speed.title"), ui.tr("config.speed.sub"), theme.DownloadIcon(), speedContent)
	pingCard := ui.newIconCard(ui.tr("config.ping.title"), ui.tr("config.ping.sub"), theme.InfoIcon(), pingContent)

	configGrid := container.NewAdaptiveGrid(optionGridColumns(),
		generalCard,
		chinaCard,
		cpuCard,
		memoryCard,
		diskCard,
		routeCard,
		speedCard,
		pingCard,
	)

	return widget.NewCard(ui.tr("config.card.title"), ui.tr("config.card.sub"), configGrid)
}
