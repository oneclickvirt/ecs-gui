package ui

import (
	"context"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ExecutionConfig struct {
	SelectedOptions  map[string]bool
	Language         string
	TestUpload       bool
	TestDownload     bool
	ChinaModeEnabled bool
	AutoDiskMethod   bool
	CpuMethod        string
	ThreadMode       string
	MemoryMethod     string
	DiskMethod       string
	DiskPath         string
	DiskMulti        bool
	Nt3Location      string
	Nt3Type          string
	SpNum            int
	PingTgdc         bool
	PingWeb          bool
	UnlockRegion     string
	UnlockIpVersion  string
	UnlockShowIP     bool
	EnableUpload     bool
	AnalyzeResult    bool
	FilePath         string
	OutputWidth      int
	PresetKey        string
	LogEnabled       bool
}

type ProgressUpdate struct {
	ItemKey  string
	Current  int
	Total    int
	Fraction float64
}

// TestUI 测试界面结构体
type TestUI struct {
	App    fyne.App
	Window fyne.Window

	// 测试选项复选框 - 完整支持所有测试项
	BasicCheck    *widget.Check
	CpuCheck      *widget.Check
	MemoryCheck   *widget.Check
	DiskCheck     *widget.Check
	UnlockCheck   *widget.Check // 跨国流媒体解锁
	SecurityCheck *widget.Check // IP质量检测

	// 解锁配置
	UnlockRegionSelect    *widget.Select
	UnlockIpVersionSelect *widget.Select
	EmailCheck            *widget.Check // 邮件端口检测
	BacktraceCheck        *widget.Check // 上游及回程线路
	Nt3Check              *widget.Check // 三网回程路由
	SpeedCheck            *widget.Check // 网络测速
	PingCheck             *widget.Check // 三网PING值
	LogCheck              *widget.Check // 启用日志记录

	// 预设模式选择
	PresetSelect *widget.Select

	// 配置选项
	LanguageSelect      *widget.Select
	ThemeSelect         *widget.Select
	CpuMethodSelect     *widget.Select
	MemoryMethodSelect  *widget.Select
	DiskMethodSelect    *widget.Select
	DiskPathEntry       *widget.Entry
	ThreadModeSelect    *widget.Select
	Nt3LocationSelect   *widget.Select
	Nt3TypeSelect       *widget.Select
	DiskMultiCheck      *widget.Check
	AutoDiskMethodCheck *widget.Check
	SpNumEntry          *widget.Entry
	OutputWidthEntry    *widget.Entry
	OutputFileEntry     *widget.Entry
	ResultUploadCheck   *widget.Check
	AnalyzeResultCheck  *widget.Check
	// 速度测试配置
	SpTestUploadCheck   *widget.Check // 测试上传速度
	SpTestDownloadCheck *widget.Check // 测试下载速度
	// 中国模式
	ChinaModeCheck *widget.Check // 启用中国专项测试

	// PING测试配置
	// 注：PingCheck控制三网PING测试，以下两个单独控制TGDC和Web测试
	PingTgdcCheck     *widget.Check // 是否测试TGDC
	PingWebCheck      *widget.Check // 是否测试流行网站
	UnlockShowIPCheck *widget.Check // 是否显示解锁测试IP标签

	// 控制按钮
	StartButton *widget.Button
	StopButton  *widget.Button

	// 结果显示 - 使用终端输出组件
	Terminal    *TerminalOutput
	ProgressBar *widget.ProgressBar
	CurrentItem *widget.Label
	StatusLabel *widget.Label
	StatusBadge *widget.Label

	// 日志相关
	LogViewer  *widget.Entry      // 日志查看器
	LogTab     *container.TabItem // 日志标签页
	MainTabs   *container.AppTabs // 主标签页容器
	LogContent string             // 日志内容存储

	// 运行状态
	IsRunning bool
	CancelCtx context.Context
	CancelFn  context.CancelFunc
	Mu        sync.Mutex

	testChecks []*widget.Check

	uiLang               string
	themeMode            string
	presetLabelToKey     map[string]string
	selectedPresetKey    string
	suppressPresetChange bool
	inBackground         bool
}
