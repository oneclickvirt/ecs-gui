package ui

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ExecutionConfig struct {
	SelectedOptions   map[string]bool
	Language          string
	ChinaModeEnabled  bool
	DeepMode          bool
	DeepDiskPaths     string
	DeepSMARTDevices  string
	DeepBurnDuration  time.Duration
	DeepGPUDevice     string
	AutoDiskMethod    bool
	CpuMethod         string
	ThreadMode        string
	MemoryMethod      string
	DiskMethod        string
	DiskPath          string
	DiskMulti         bool
	Nt3Location       string
	Nt3Type           string
	SpNum             int
	PingTgdc          bool
	PingWeb           bool
	UnlockRegion      string
	UnlockIpVersion   string
	UnlockShowIP      bool
	UnlockInterface   string
	UnlockDNS         string
	UnlockHTTPProxy   string
	UnlockSOCKSProxy  string
	UnlockConcurrency int
	EnableUpload      bool
	AnalyzeResult     bool
	FilePath          string
	JSONPath          string
	OutputWidth       int
	MaxDuration       time.Duration
	HardwareBudget    time.Duration
	DataOffline       bool
	PrivacyMode       bool
	PresetKey         string
	LogEnabled        bool
}

type ProgressUpdate struct {
	ItemKey  string
	Current  int
	Total    int
	Fraction float64
}

type StructuredSection struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
}

type StructuredDataVersion struct {
	Schema      string    `json:"schema"`
	GeneratedAt time.Time `json:"generated_at"`
	Source      string    `json:"source"`
	Fallback    string    `json:"fallback,omitempty"`
	File        string    `json:"file,omitempty"`
	Count       int       `json:"count,omitempty"`
}

type StructuredDataFile struct {
	File        string    `json:"file"`
	Schema      string    `json:"schema"`
	GeneratedAt time.Time `json:"generated_at"`
	Source      string    `json:"source"`
	Fallback    string    `json:"fallback,omitempty"`
	Count       int       `json:"count"`
	Status      string    `json:"status"`
	Reason      string    `json:"reason,omitempty"`
}

type StructuredComponent struct {
	Name          string          `json:"name"`
	SchemaVersion string          `json:"schema_version"`
	Status        string          `json:"status"`
	Reason        string          `json:"reason,omitempty"`
	DurationMS    int64           `json:"duration_ms,omitempty"`
	Payload       json.RawMessage `json:"payload,omitempty"`
}

type StructuredRunResult struct {
	SchemaVersion string                 `json:"schema_version"`
	ECSVersion    string                 `json:"ecs_version,omitempty"`
	Status        string                 `json:"status"`
	StartedAt     time.Time              `json:"started_at"`
	FinishedAt    time.Time              `json:"finished_at"`
	DurationMS    int64                  `json:"duration_ms"`
	DeepMode      bool                   `json:"deep_mode"`
	PrivacyMode   bool                   `json:"privacy_mode"`
	Data          *StructuredDataVersion `json:"data,omitempty"`
	DataFiles     []StructuredDataFile   `json:"data_files,omitempty"`
	Sections      []StructuredSection    `json:"sections"`
	Components    []StructuredComponent  `json:"components,omitempty"`
	Text          string                 `json:"text,omitempty"`
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
	UnlockRegionSelect     *widget.Select
	UnlockIpVersionSelect  *widget.Select
	UnlockInterfaceEntry   *widget.Entry
	UnlockDNSEntry         *widget.Entry
	UnlockHTTPProxyEntry   *widget.Entry
	UnlockSOCKSProxyEntry  *widget.Entry
	UnlockConcurrencyEntry *widget.Entry
	EmailCheck             *widget.Check // 邮件端口检测
	BacktraceCheck         *widget.Check // 上游及回程线路
	Nt3Check               *widget.Check // 三网回程路由
	SpeedCheck             *widget.Check // 网络测速
	PingCheck              *widget.Check // 三网PING值
	LogCheck               *widget.Check // 启用日志记录

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
	DeepModeCheck       *widget.Check
	DeepDiskPathsEntry  *widget.Entry
	DeepSMARTEntry      *widget.Entry
	DeepBurnEntry       *widget.Entry
	DeepGPUEntry        *widget.Entry
	SpNumEntry          *widget.Entry
	OutputWidthEntry    *widget.Entry
	OutputFileEntry     *widget.Entry
	JSONPathEntry       *widget.Entry
	MaxDurationEntry    *widget.Entry
	HardwareBudgetEntry *widget.Entry
	DataOfflineCheck    *widget.Check
	PrivacyModeCheck    *widget.Check
	ResultUploadCheck   *widget.Check
	AnalyzeResultCheck  *widget.Check
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
	Terminal              *TerminalOutput
	ProgressBar           *widget.ProgressBar
	CurrentItem           *widget.Label
	StatusLabel           *widget.Label
	StatusBadge           *widget.Label
	DataStatusLabel       *widget.Label
	PartialReasonLabel    *widget.Label
	StructuredDetailsView *widget.Entry

	// 日志相关
	LogViewer  *widget.Entry      // 日志查看器
	LogTab     *container.TabItem // 日志标签页
	MainTabs   *container.AppTabs // 主标签页容器
	LogContent string             // 日志内容存储

	// 运行状态
	IsRunning        bool
	CancelCtx        context.Context
	CancelFn         context.CancelFunc
	Mu               sync.Mutex
	StructuredResult *StructuredRunResult

	testChecks []*widget.Check

	uiLang               string
	themeMode            string
	presetLabelToKey     map[string]string
	selectedPresetKey    string
	suppressPresetChange bool
	inBackground         bool
}
