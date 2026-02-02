package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createOptionsPanel 创建选项面板（测试项目 + 配置选项整合在一起）
func (ui *TestUI) createOptionsPanel() fyne.CanvasObject {
	// 预设模式选择
	ui.PresetSelect = widget.NewSelect(
		[]string{
			"自定义",
			"1. 融合怪完全体(能测全测)",
			"2. 极简版(系统信息+CPU+内存+磁盘+测速节点5个)",
			"3. 精简版(系统信息+CPU+内存+磁盘+常用流媒体+路由+测速节点5个)",
			"4. 精简网络版(系统信息+CPU+内存+磁盘+回程+路由+测速节点5个)",
			"5. 精简解锁版(系统信息+CPU+内存+磁盘IO+御三家+常用流媒体+测速节点5个)",
			"6. 网络单项(IP质量检测+上游及三网回程+广州三网回程详细路由+全国延迟+TGDC+网站延迟+测速节点11个)",
			"7. 解锁单项(御三家解锁+常用流媒体解锁)",
			"8. 硬件单项(系统信息+CPU+dd磁盘测试+fio磁盘测试)",
			"9. IP质量检测(15个数据库的IP质量检测+邮件端口检测)",
			"10. 三网回程线路检测+三网回程详细路由(北京上海广州成都)+全国延迟+TGDC+网站延迟",
		},
		ui.onPresetChanged,
	)
	ui.PresetSelect.Selected = "自定义"

	presetSection := widget.NewCard("预设模式", "快速选择测试组合", ui.PresetSelect)

	// === 测试项目复选框 ===
	ui.BasicCheck = widget.NewCheck("基础信息测试", nil)
	ui.BasicCheck.Checked = true

	ui.CpuCheck = widget.NewCheck("CPU 性能测试", nil)
	ui.CpuCheck.Checked = true

	ui.MemoryCheck = widget.NewCheck("内存性能测试", nil)
	ui.MemoryCheck.Checked = true

	ui.DiskCheck = widget.NewCheck("磁盘性能测试", nil)
	ui.DiskCheck.Checked = true

	ui.CommCheck = widget.NewCheck("御三家流媒体测试", nil)
	ui.CommCheck.Checked = false

	ui.UnlockCheck = widget.NewCheck("跨国流媒体解锁测试", nil)
	ui.UnlockCheck.Checked = false

	ui.SecurityCheck = widget.NewCheck("IP质量检测", nil)
	ui.SecurityCheck.Checked = false

	ui.EmailCheck = widget.NewCheck("邮件端口检测", nil)
	ui.EmailCheck.Checked = false

	ui.BacktraceCheck = widget.NewCheck("上游及回程线路检测", nil)
	ui.BacktraceCheck.Checked = false

	ui.Nt3Check = widget.NewCheck("三网回程路由检测", nil)
	ui.Nt3Check.Checked = false

	ui.SpeedCheck = widget.NewCheck("网络测速", nil)
	ui.SpeedCheck.Checked = false

	ui.PingCheck = widget.NewCheck("三网PING值检测", nil)
	ui.PingCheck.Checked = false

	ui.LogCheck = widget.NewCheck("启用日志记录", ui.onLogCheckChanged)
	ui.LogCheck.Checked = false

	// 全选/取消全选按钮
	selectAllBtn := widget.NewButton("全选", func() {
		ui.setAllChecks(true)
	})

	deselectAllBtn := widget.NewButton("取消全选", func() {
		ui.setAllChecks(false)
	})

	buttonRow := container.NewHBox(selectAllBtn, deselectAllBtn)

	// 测试项目分组 - 使用网格布局，每行2个
	basicTests := container.NewVBox(
		ui.BasicCheck,
		ui.CpuCheck,
		ui.MemoryCheck,
		ui.DiskCheck,
	)

	networkTests := container.NewVBox(
		ui.SpeedCheck,
		ui.SecurityCheck,
		ui.EmailCheck,
		ui.BacktraceCheck,
	)

	advancedTests := container.NewVBox(
		ui.Nt3Check,
		ui.PingCheck,
		ui.CommCheck,
		ui.UnlockCheck,
	)

	testsGrid := container.NewGridWithColumns(3,
		basicTests,
		networkTests,
		advancedTests,
	)

	testsSection := widget.NewCard("测试项目", "", container.NewVBox(
		buttonRow,
		testsGrid,
	))

	// === 配置选项 ===
	configSection := ui.createConfigSection()

	// 整合所有内容
	allContent := container.NewVBox(
		presetSection,
		testsSection,
		configSection,
	)

	return allContent
}

// createConfigSection 创建配置选项区域
func (ui *TestUI) createConfigSection() fyne.CanvasObject {
	// 语言选择
	ui.LanguageSelect = widget.NewSelect(
		[]string{"中文", "English"},
		func(value string) {},
	)
	ui.LanguageSelect.Selected = "中文"

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
	ui.DiskPathEntry.SetPlaceHolder("/tmp 或留空自动检测")

	ui.DiskMultiCheck = widget.NewCheck("启用多磁盘检测", nil)
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
	ui.SpNumEntry.SetPlaceHolder("每运营商测速节点数")

	// 速度测试上传下载控制
	ui.SpTestUploadCheck = widget.NewCheck("测试上传速度", nil)
	ui.SpTestUploadCheck.Checked = true

	ui.SpTestDownloadCheck = widget.NewCheck("测试下载速度", nil)
	ui.SpTestDownloadCheck.Checked = true

	// 中国模式
	ui.ChinaModeCheck = widget.NewCheck("启用中国专项测试", nil)
	ui.ChinaModeCheck.Checked = false

	// PING测试配置
	ui.PingTgdcCheck = widget.NewCheck("测试Telegram DC", nil)
	ui.PingTgdcCheck.Checked = false

	ui.PingWebCheck = widget.NewCheck("测试流行网站", nil)
	ui.PingWebCheck.Checked = false

	// 使用表单布局更紧凑
	configForm := container.NewVBox(
		widget.NewLabel("通用配置:"),
		container.NewGridWithColumns(2,
			widget.NewLabel("语言:"),
			ui.LanguageSelect,
		),
		ui.LogCheck, // 日志选项
		widget.NewSeparator(),
		widget.NewLabel("中国专项测试:"),
		ui.ChinaModeCheck,
		widget.NewLabel("(启用后将禁用流媒体测试，启用PING测试)"),
		widget.NewSeparator(),
		widget.NewLabel("CPU配置:"),
		container.NewGridWithColumns(2,
			widget.NewLabel("测试方法:"),
			ui.CpuMethodSelect,
			widget.NewLabel("线程模式:"),
			ui.ThreadModeSelect,
		),
		widget.NewSeparator(),
		widget.NewLabel("内存配置:"),
		container.NewGridWithColumns(2,
			widget.NewLabel("测试方法:"),
			ui.MemoryMethodSelect,
		),
		widget.NewSeparator(),
		widget.NewLabel("磁盘配置:"),
		container.NewGridWithColumns(2,
			widget.NewLabel("测试方法:"),
			ui.DiskMethodSelect,
			widget.NewLabel("测试路径:"),
			ui.DiskPathEntry,
		),
		ui.DiskMultiCheck,
		widget.NewSeparator(),
		widget.NewLabel("三网回程配置:"),
		container.NewGridWithColumns(2,
			widget.NewLabel("测试地点:"),
			ui.Nt3LocationSelect,
			widget.NewLabel("测试类型:"),
			ui.Nt3TypeSelect,
		),
		widget.NewSeparator(),
		widget.NewLabel("测速配置:"),
		container.NewGridWithColumns(2,
			widget.NewLabel("节点数/运营商:"),
			ui.SpNumEntry,
		),
		ui.SpTestUploadCheck,
		ui.SpTestDownloadCheck,
		widget.NewSeparator(),
		widget.NewLabel("PING测试配置:"),
		widget.NewLabel("(三网PING值测试不包含以下内容，勾选才测试)"),
		ui.PingTgdcCheck,
		ui.PingWebCheck,
	)

	return widget.NewCard("详细配置", "调整测试参数", configForm)
}
