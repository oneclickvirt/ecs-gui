package ui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func newTestUIForTest(t *testing.T) *TestUI {
	t.Helper()
	app := test.NewApp()
	ui := NewTestUI(app)
	t.Cleanup(func() {
		if ui.Terminal != nil {
			ui.Terminal.Destroy()
		}
		app.Quit()
	})
	return ui
}

func TestDesktopConfigSectionStaysWithinDefaultWindow(t *testing.T) {
	if isMobilePlatform() {
		t.Skip("desktop layout only")
	}

	ui := newTestUIForTest(t)
	for _, language := range []string{langZH, langEN} {
		ui.uiLang = language
		size := ui.createConfigSection().MinSize()
		if size.Width > 900 {
			t.Fatalf("%s config minimum width = %.1f, want <= 900", language, size.Width)
		}
		if size.Height > 1900 {
			t.Fatalf("%s config minimum height = %.1f, want <= 1900", language, size.Height)
		}
	}
}

func TestCollectExecutionConfigClampsNumericValues(t *testing.T) {
	ui := newTestUIForTest(t)

	ui.SpNumEntry.SetText("999")
	ui.OutputWidthEntry.SetText("999")
	ui.OutputFileEntry.SetText("")

	config := ui.collectExecutionConfig()
	if config.SpNum != 20 {
		t.Fatalf("expected sp num to clamp to 20, got %d", config.SpNum)
	}
	if config.OutputWidth != 120 {
		t.Fatalf("expected output width to clamp to 120, got %d", config.OutputWidth)
	}
	if config.FilePath != "goecs.md" {
		t.Fatalf("expected default file path, got %q", config.FilePath)
	}
}

func TestCollectExecutionConfigDeepDefaultsAreDisabledAndEmpty(t *testing.T) {
	ui := newTestUIForTest(t)
	if !ui.DeepDiskPathsEntry.Disabled() || !ui.DeepSMARTEntry.Disabled() || !ui.DeepBurnEntry.Disabled() || !ui.DeepGPUEntry.Disabled() {
		t.Fatal("deep target inputs must be disabled by default")
	}
	ui.DeepDiskPathsEntry.SetText("/must/not/run")
	ui.DeepSMARTEntry.SetText("/dev/must-not-run")
	ui.DeepBurnEntry.SetText("1m")
	ui.DeepGPUEntry.SetText("must-not-run")
	config := ui.collectExecutionConfig()
	if config.DeepMode || config.DeepDiskPaths != "" || config.DeepSMARTDevices != "" || config.DeepBurnDuration != 0 || config.DeepGPUDevice != "" {
		t.Fatalf("unexpected deep defaults: %#v", config)
	}
}

func TestCollectExecutionConfigParsesExplicitDeepInputs(t *testing.T) {
	ui := newTestUIForTest(t)
	ui.DeepModeCheck.SetChecked(true)
	if ui.DeepDiskPathsEntry.Disabled() || ui.DeepSMARTEntry.Disabled() || ui.DeepBurnEntry.Disabled() || ui.DeepGPUEntry.Disabled() {
		t.Fatal("deep target inputs must be enabled after explicit opt-in")
	}
	ui.DeepDiskPathsEntry.SetText("  /mnt/a,/mnt/b  ")
	ui.DeepSMARTEntry.SetText(" /dev/sda ")
	ui.DeepBurnEntry.SetText("45s")
	ui.DeepGPUEntry.SetText(" gpu0 ")

	config := ui.collectExecutionConfig()
	if !config.DeepMode || config.DeepDiskPaths != "/mnt/a,/mnt/b" || config.DeepSMARTDevices != "/dev/sda" || config.DeepBurnDuration != 45*time.Second || config.DeepGPUDevice != "gpu0" {
		t.Fatalf("unexpected explicit deep config: %#v", config)
	}
	ui.DeepBurnEntry.SetText("invalid")
	if got := ui.collectExecutionConfig().DeepBurnDuration; got != 0 {
		t.Fatalf("invalid burn duration = %s, want disabled", got)
	}
}

func TestCollectExecutionConfigParsesUnlockNetworkInputs(t *testing.T) {
	ui := newTestUIForTest(t)
	ui.UnlockInterfaceEntry.SetText("eth0")
	ui.UnlockDNSEntry.SetText("1.1.1.1,2606:4700:4700::1111")
	ui.UnlockHTTPProxyEntry.SetText("http://127.0.0.1:8080")
	ui.UnlockSOCKSProxyEntry.SetText("socks5://127.0.0.1:1080")
	ui.UnlockConcurrencyEntry.SetText("150")
	config := ui.collectExecutionConfig()
	if config.UnlockInterface != "eth0" || config.UnlockDNS == "" || config.UnlockHTTPProxy == "" || config.UnlockSOCKSProxy == "" {
		t.Fatalf("unlock network inputs were not preserved: %#v", config)
	}
	if config.UnlockConcurrency != 100 {
		t.Fatalf("unlock concurrency = %d, want clamp to 100", config.UnlockConcurrency)
	}
}

func TestCollectExecutionConfigParsesRuntimeDataAndPrivacyInputs(t *testing.T) {
	ui := newTestUIForTest(t)
	ui.MaxDurationEntry.SetText("9m")
	ui.HardwareBudgetEntry.SetText("90s")
	ui.JSONPathEntry.SetText(" result.json ")
	ui.DataOfflineCheck.SetChecked(true)
	ui.ResultUploadCheck.SetChecked(true)
	ui.PrivacyModeCheck.SetChecked(true)

	config := ui.collectExecutionConfig()
	if config.MaxDuration != 9*time.Minute || config.HardwareBudget != 90*time.Second {
		t.Fatalf("runtime budgets were not parsed: %#v", config)
	}
	if !config.DataOffline || config.JSONPath != "result.json" {
		t.Fatalf("data/JSON inputs were not preserved: %#v", config)
	}
	if !config.PrivacyMode || config.EnableUpload {
		t.Fatalf("privacy mode must fail closed for uploads: %#v", config)
	}
}

func TestStandardPresetSelectsExpectedOptions(t *testing.T) {
	ui := newTestUIForTest(t)

	ui.onPresetChanged(ui.presetLabelByKey("standard"))

	if !ui.BasicCheck.Checked || !ui.CpuCheck.Checked || !ui.MemoryCheck.Checked || !ui.DiskCheck.Checked {
		t.Fatal("standard preset should enable core performance checks")
	}
	if !ui.UnlockCheck.Checked || !ui.Nt3Check.Checked || !ui.SpeedCheck.Checked {
		t.Fatal("standard preset should enable unlock, route, and speed checks")
	}
	if ui.PingCheck.Checked || ui.PingTgdcCheck.Checked || ui.PingWebCheck.Checked {
		t.Fatal("standard preset should not enable ping extensions")
	}
	if ui.SpNumEntry.Text != "5" {
		t.Fatalf("standard preset should set speed node count to 5, got %q", ui.SpNumEntry.Text)
	}
}

func TestFullPresetMatchesGoECSOptionOneEnhancements(t *testing.T) {
	ui := newTestUIForTest(t)
	ui.onPresetChanged(ui.presetLabelByKey("full"))

	config := ui.collectExecutionConfig()
	if !config.SelectedOptions["ping"] || !config.DiskMulti || !config.DeepMode || config.DeepBurnDuration != 20*time.Second {
		t.Fatalf("full preset did not enable Ping/deep hardware defaults: %#v", config)
	}
	if !config.PingTgdc || !config.PingWeb || !config.UnlockShowIP {
		t.Fatalf("full preset did not enable network enhancements: %#v", config)
	}
	if config.PingSortOrder != "latency" || config.PingScope != "auto" || config.TCPSortOrder != "name" {
		t.Fatalf("full preset ordering defaults are inconsistent: %#v", config)
	}

	ui.onPresetChanged(ui.presetLabelByKey("standard"))
	standard := ui.collectExecutionConfig()
	if standard.DeepMode || standard.DiskMulti || standard.DeepBurnDuration != 0 {
		t.Fatalf("standard preset retained full-only enhancements: %#v", standard)
	}
}

func TestCollectExecutionConfigKeepsNetworkOrderingSelections(t *testing.T) {
	ui := newTestUIForTest(t)
	ui.PingSortSelect.SetSelected("name")
	ui.PingScopeSelect.SetSelected("international")
	ui.TCPSortSelect.SetSelected("latency")
	config := ui.collectExecutionConfig()
	if config.PingSortOrder != "name" || config.PingScope != "international" || config.TCPSortOrder != "latency" {
		t.Fatalf("network ordering selections were not retained: %#v", config)
	}
}

func TestBuildProgressStepsHonorsChinaMode(t *testing.T) {
	steps := buildProgressSteps(ExecutionConfig{
		SelectedOptions: map[string]bool{
			"unlock": true,
			"ping":   false,
		},
		ChinaModeEnabled: true,
	}, true)

	joined := strings.Join(steps, ",")
	if strings.Contains(joined, "progress.unlock") {
		t.Fatal("china mode should remove streaming unlock from progress steps")
	}
	if !strings.Contains(joined, "progress.ping") {
		t.Fatal("china mode should force ping progress step")
	}
}

func TestBuildProgressStepsKeepsTelegramAndWebsiteSeparate(t *testing.T) {
	steps := buildProgressSteps(ExecutionConfig{
		SelectedOptions: map[string]bool{"ping": true},
		PingTgdc:        true,
		PingWeb:         true,
	}, true)
	joined := strings.Join(steps, ",")
	for _, key := range []string{"progress.ping", "progress.tgdc", "progress.web"} {
		if !strings.Contains(joined, key) {
			t.Fatalf("structured progress is missing %s: %v", key, steps)
		}
	}
}

func TestEffectiveNT3TypeForStack(t *testing.T) {
	cases := []struct {
		requested string
		stack     string
		want      string
	}{
		{"both", "IPv4", "ipv4"},
		{"both", "IPv6", "ipv6"},
		{"ipv6", "IPv4", "ipv4"},
		{"ipv4", "IPv6", "ipv6"},
		{"both", "DualStack", "both"},
		{"unexpected", "DualStack", "both"},
	}

	for _, tt := range cases {
		if got := effectiveNT3TypeForStack(tt.requested, tt.stack); got != tt.want {
			t.Fatalf("effectiveNT3TypeForStack(%q, %q) = %q, want %q", tt.requested, tt.stack, got, tt.want)
		}
	}
}

func TestCollectExecutionConfigDefaultsNT3ToBoth(t *testing.T) {
	ui := newTestUIForTest(t)
	config := ui.collectExecutionConfig()
	if config.Nt3Type != "both" {
		t.Fatalf("default NT3 type = %q, want both", config.Nt3Type)
	}
}

func TestBuildProgressStepsSkipsNetworkStagesWhenOffline(t *testing.T) {
	steps := buildProgressSteps(ExecutionConfig{
		SelectedOptions: map[string]bool{
			"basic":     true,
			"unlock":    true,
			"security":  true,
			"email":     true,
			"backtrace": true,
			"nt3":       true,
			"speed":     true,
		},
		EnableUpload: true,
	}, false)

	joined := strings.Join(steps, ",")
	for _, key := range []string{"progress.unlock", "progress.email", "progress.backtrace", "progress.nt3", "progress.speed", "progress.upload"} {
		if strings.Contains(joined, key) {
			t.Fatalf("offline progress steps should not include %s: %v", key, steps)
		}
	}
	if !strings.Contains(joined, "progress.basic_security") {
		t.Fatalf("offline progress should still include local basic/security step: %v", steps)
	}
}

func TestNeedsNetworkIdentityProbe(t *testing.T) {
	cases := []struct {
		name            string
		connected       bool
		basicStatus     bool
		securityStatus  bool
		unlockStatus    bool
		backtraceStatus bool
		want            bool
	}{
		{name: "single unlock needs hidden ip probe", connected: true, unlockStatus: true, want: true},
		{name: "single backtrace needs hidden ip probe", connected: true, backtraceStatus: true, want: true},
		{name: "basic already probes ip", connected: true, basicStatus: true, unlockStatus: true, want: false},
		{name: "security already probes ip", connected: true, securityStatus: true, backtraceStatus: true, want: false},
		{name: "offline skips probe", connected: false, unlockStatus: true, backtraceStatus: true, want: false},
		{name: "speed does not need ip identity", connected: true, want: false},
	}

	for _, tt := range cases {
		got := needsNetworkIdentityProbe(tt.connected, tt.basicStatus, tt.securityStatus, tt.unlockStatus, tt.backtraceStatus)
		if got != tt.want {
			t.Fatalf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestCollectExecutionConfigKeepsManualSpeedSelection(t *testing.T) {
	ui := newTestUIForTest(t)

	ui.onPresetChanged(ui.presetLabelByKey("hardware_only"))
	ui.SpeedCheck.Checked = true

	config := ui.collectExecutionConfig()
	if !config.SelectedOptions["speed"] {
		t.Fatal("manually selected speed test should remain enabled in execution config")
	}
}

func TestSetProgressClampsFractionAndShowsCurrentItem(t *testing.T) {
	ui := newTestUIForTest(t)

	ui.setProgress(ProgressUpdate{
		ItemKey:  "progress.cpu",
		Current:  5,
		Total:    3,
		Fraction: 2,
	})

	if ui.ProgressBar.Value != 1 {
		t.Fatalf("expected progress to clamp to 1, got %f", ui.ProgressBar.Value)
	}
	if !strings.Contains(ui.CurrentItem.Text, "CPU") || !strings.Contains(ui.CurrentItem.Text, "5/3") {
		t.Fatalf("current item did not render progress details: %q", ui.CurrentItem.Text)
	}

	ui.setProgress(ProgressUpdate{ItemKey: "progress.memory", Fraction: -1})
	if ui.ProgressBar.Value != 0 {
		t.Fatalf("expected progress to clamp to 0, got %f", ui.ProgressBar.Value)
	}
}

func TestSafeUploadFilePathUsesTempDirAndSanitizesName(t *testing.T) {
	path := safeUploadFilePath("/etc/passwd")
	if filepath.Clean(filepath.Dir(path)) != filepath.Clean(os.TempDir()) {
		t.Fatalf("upload path should be in temp dir, got %q", path)
	}
	if filepath.Base(path) == "passwd" || !strings.HasSuffix(filepath.Base(path), ".md") {
		t.Fatalf("upload filename should be sanitized and markdown-like, got %q", filepath.Base(path))
	}
	if strings.Contains(path, "/etc/") {
		t.Fatalf("upload path should not preserve unsafe source directories: %q", path)
	}

	if got := sanitizeUploadFileName(".env"); got != "goecs.md" {
		t.Fatalf("hidden/sensitive filenames should use fallback, got %q", got)
	}
	if got := sanitizeUploadFileName("my-token.txt"); got != "goecs.md" {
		t.Fatalf("sensitive filenames should use fallback, got %q", got)
	}
	if got := sanitizeUploadFileName("report?.html"); got != "report-.md" {
		t.Fatalf("unexpected sanitized name: %q", got)
	}
}

func TestFriendlyErrorMessageBuckets(t *testing.T) {
	ui := newTestUIForTest(t)
	ui.uiLang = langEN

	cases := []struct {
		err  error
		want string
	}{
		{errors.New("permission denied"), ui.tr("error.permission")},
		{errors.New("network connection reset"), ui.tr("error.network")},
		{errors.New("context deadline timeout"), ui.tr("error.timeout")},
		{errors.New("cancelled by user"), ui.tr("error.cancelled")},
		{errors.New("unexpected failure"), ui.tr("error.generic")},
	}
	for _, tt := range cases {
		if got := ui.friendlyErrorMessage(tt.err); got != tt.want {
			t.Fatalf("friendlyErrorMessage(%q) = %q, want %q", tt.err, got, tt.want)
		}
	}
}

func TestFormatHumanDurationLanguages(t *testing.T) {
	if got := formatHumanDuration(75*time.Second, langEN); got != "1 min 15 sec" {
		t.Fatalf("unexpected english duration: %q", got)
	}
	if got := formatHumanDuration(9*time.Second, langZH); got != "9 秒" {
		t.Fatalf("unexpected chinese duration: %q", got)
	}
	if got := formatHumanDuration(-time.Second, langEN); got != "0 sec" {
		t.Fatalf("negative duration should clamp to zero, got %q", got)
	}
}

func TestFormatResultExportWrapsPlainTextOnce(t *testing.T) {
	exported := formatResultExport("hello\n")
	if !strings.HasPrefix(exported, "# GOECS Result\n\n```text\nhello\n```") {
		t.Fatalf("unexpected export format: %q", exported)
	}

	again := formatResultExport(exported)
	if strings.Count(again, "# GOECS Result") != 1 {
		t.Fatalf("export header duplicated: %q", again)
	}
}

func TestBuildResultSummaryDetectsSections(t *testing.T) {
	output := "CPU-Test\nSpeed-Test\n10.00 Mbps\nError: timeout\n"
	summary := BuildResultSummary("en", output)

	if !strings.Contains(summary, "CPU") || !strings.Contains(summary, "Speed") {
		t.Fatalf("summary did not detect sections: %q", summary)
	}
	if !strings.Contains(summary, "Speed Rows         : 1") {
		t.Fatalf("summary did not count speed rows: %q", summary)
	}
}
