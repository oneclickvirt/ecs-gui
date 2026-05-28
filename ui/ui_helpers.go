package ui

import (
	"runtime"
	"strconv"

	"fyne.io/fyne/v2"
)

type uiStateSnapshot struct {
	checks       map[string]bool
	selections   map[string]string
	entries      map[string]string
	presetKey    string
	terminalText string
	logContent   string
	logEnabled   bool
	activeTab    int
	statusText   string
	statusBadge  string
}

func (ui *TestUI) runOnUI(fn func()) {
	fyne.Do(fn)
}

func isMobilePlatform() bool {
	return runtime.GOOS == "android" || runtime.GOOS == "ios"
}

func optionGridColumns() int {
	if isMobilePlatform() {
		return 1
	}
	return 2
}

func (ui *TestUI) statusBadge(statusKey string) string {
	switch statusKey {
	case "status.running", "status.executing":
		return ui.tr("badge.running")
	case "status.stopping", "status.stopped":
		return ui.tr("badge.stopped")
	case "status.failed":
		return ui.tr("badge.failed")
	case "status.done":
		return ui.tr("badge.done")
	default:
		return ui.tr("badge.ready")
	}
}

func (ui *TestUI) setStatus(statusKey string) {
	if ui.StatusLabel != nil {
		ui.StatusLabel.SetText(ui.tr(statusKey))
	}
	if ui.StatusBadge != nil {
		ui.StatusBadge.SetText(ui.statusBadge(statusKey))
	}
}

func (ui *TestUI) tr(key string) string {
	lang := ui.uiLang
	if lang != langEN {
		lang = langZH
	}

	if table, ok := i18nText[key]; ok {
		if val, ok := table[lang]; ok && val != "" {
			return val
		}
		if val, ok := table[langZH]; ok {
			return val
		}
	}
	return key
}

func (ui *TestUI) selectedLanguageCode() string {
	if ui.LanguageSelect != nil && ui.LanguageSelect.Selected == "English" {
		return langEN
	}
	return langZH
}

func (ui *TestUI) rebuildPresetMappings() []string {
	labels := make([]string, 0, len(presetDefs))
	ui.presetLabelToKey = make(map[string]string, len(presetDefs))
	for _, def := range presetDefs {
		label := ui.tr(def.labelKey)
		labels = append(labels, label)
		ui.presetLabelToKey[label] = def.key
	}
	return labels
}

func (ui *TestUI) presetLabelByKey(key string) string {
	for _, def := range presetDefs {
		if def.key == key {
			return ui.tr(def.labelKey)
		}
	}
	return ui.tr("preset.custom")
}

func (ui *TestUI) snapshotUIState() uiStateSnapshot {
	state := uiStateSnapshot{
		checks: map[string]bool{
			"basic":     ui.BasicCheck.Checked,
			"cpu":       ui.CpuCheck.Checked,
			"memory":    ui.MemoryCheck.Checked,
			"disk":      ui.DiskCheck.Checked,
			"unlock":    ui.UnlockCheck.Checked,
			"security":  ui.SecurityCheck.Checked,
			"email":     ui.EmailCheck.Checked,
			"backtrace": ui.BacktraceCheck.Checked,
			"nt3":       ui.Nt3Check.Checked,
			"speed":     ui.SpeedCheck.Checked,
			"ping":      ui.PingCheck.Checked,
			"diskMulti": ui.DiskMultiCheck.Checked,
			"spUp":      ui.SpTestUploadCheck.Checked,
			"spDown":    ui.SpTestDownloadCheck.Checked,
			"chinaMode": ui.ChinaModeCheck.Checked,
			"pingTgdc":  ui.PingTgdcCheck.Checked,
			"pingWeb":   ui.PingWebCheck.Checked,
			"enableLog": ui.LogCheck.Checked,
		},
		selections: map[string]string{
			"language":     ui.LanguageSelect.Selected,
			"cpuMethod":    ui.CpuMethodSelect.Selected,
			"threadMode":   ui.ThreadModeSelect.Selected,
			"memMethod":    ui.MemoryMethodSelect.Selected,
			"diskMethod":   ui.DiskMethodSelect.Selected,
			"nt3Loc":       ui.Nt3LocationSelect.Selected,
			"nt3Type":      ui.Nt3TypeSelect.Selected,
			"unlockRegion": unlockRegionLabelToCode(ui.UnlockRegionSelect.Selected, ui.uiLang),
			"unlockIpVer":  ui.UnlockIpVersionSelect.Selected,
		},
		entries: map[string]string{
			"diskPath": ui.DiskPathEntry.Text,
			"spNum":    ui.SpNumEntry.Text,
		},
		presetKey:  ui.selectedPresetKey,
		logContent: ui.LogContent,
		logEnabled: ui.LogCheck.Checked,
		activeTab:  ui.MainTabs.SelectedIndex(),
		statusText: ui.StatusLabel.Text,
	}

	if ui.StatusBadge != nil {
		state.statusBadge = ui.StatusBadge.Text
	}

	if ui.Terminal != nil {
		state.terminalText = ui.Terminal.GetText()
	}

	return state
}

func (ui *TestUI) restoreUIState(state uiStateSnapshot) {
	ui.suppressPresetChange = true
	ui.selectedPresetKey = state.presetKey
	ui.PresetSelect.SetSelected(ui.presetLabelByKey(ui.selectedPresetKey))
	ui.suppressPresetChange = false

	ui.BasicCheck.Checked = state.checks["basic"]
	ui.CpuCheck.Checked = state.checks["cpu"]
	ui.MemoryCheck.Checked = state.checks["memory"]
	ui.DiskCheck.Checked = state.checks["disk"]
	ui.UnlockCheck.Checked = state.checks["unlock"]
	ui.SecurityCheck.Checked = state.checks["security"]
	ui.EmailCheck.Checked = state.checks["email"]
	ui.BacktraceCheck.Checked = state.checks["backtrace"]
	ui.Nt3Check.Checked = state.checks["nt3"]
	ui.SpeedCheck.Checked = state.checks["speed"]
	ui.PingCheck.Checked = state.checks["ping"]
	ui.DiskMultiCheck.Checked = state.checks["diskMulti"]
	ui.SpTestUploadCheck.Checked = state.checks["spUp"]
	ui.SpTestDownloadCheck.Checked = state.checks["spDown"]
	ui.ChinaModeCheck.Checked = state.checks["chinaMode"]
	ui.PingTgdcCheck.Checked = state.checks["pingTgdc"]
	ui.PingWebCheck.Checked = state.checks["pingWeb"]
	ui.LogCheck.Checked = state.checks["enableLog"]

	ui.LanguageSelect.SetSelected(state.selections["language"])
	ui.CpuMethodSelect.SetSelected(state.selections["cpuMethod"])
	ui.ThreadModeSelect.SetSelected(state.selections["threadMode"])
	ui.MemoryMethodSelect.SetSelected(state.selections["memMethod"])
	ui.DiskMethodSelect.SetSelected(state.selections["diskMethod"])
	ui.Nt3LocationSelect.SetSelected(state.selections["nt3Loc"])
	ui.Nt3TypeSelect.SetSelected(state.selections["nt3Type"])
	if code := state.selections["unlockRegion"]; code != "" {
		ui.UnlockRegionSelect.SetSelected(unlockRegionCodeToLabel(code, ui.uiLang))
	}
	if ver := state.selections["unlockIpVer"]; ver != "" {
		ui.UnlockIpVersionSelect.SetSelected(ver)
	}

	ui.DiskPathEntry.SetText(state.entries["diskPath"])
	ui.SpNumEntry.SetText(state.entries["spNum"])

	ui.refreshAllChecks()
	ui.refreshSpeedTestChecks()

	ui.LogContent = state.logContent
	if state.logEnabled {
		ui.addLogTab()
		ui.refreshLogContent()
	}

	if ui.Terminal != nil {
		ui.Terminal.SetFullText(state.terminalText)
	}

	if state.activeTab >= 0 && state.activeTab < len(ui.MainTabs.Items) {
		ui.MainTabs.SelectIndex(state.activeTab)
	}

	if state.statusText != "" {
		ui.StatusLabel.SetText(state.statusText)
	}
	if state.statusBadge != "" && ui.StatusBadge != nil {
		ui.StatusBadge.SetText(state.statusBadge)
	}
}

// hasSelectedTests 检查是否有选中的测试项
func (ui *TestUI) hasSelectedTests() bool {
	for _, check := range ui.testChecks {
		if check != nil && check.Checked {
			return true
		}
	}
	return false
}

// isCancelled 检查测试是否被取消
func (ui *TestUI) isCancelled() bool {
	if ui.CancelCtx == nil {
		return false
	}
	select {
	case <-ui.CancelCtx.Done():
		return true
	default:
		return false
	}
}

func (ui *TestUI) isRunning() bool {
	ui.Mu.Lock()
	defer ui.Mu.Unlock()
	return ui.IsRunning
}

// resetUIState 重置UI状态（线程安全）
func (ui *TestUI) resetUIState() {
	ui.Mu.Lock()
	ui.IsRunning = false
	ui.Mu.Unlock()

	ui.runOnUI(func() {
		ui.StartButton.Enable()
		ui.StopButton.Disable()
		ui.ProgressBar.Hide()
		ui.ProgressBar.SetValue(0)

		if ui.StatusLabel.Text == ui.tr("status.stopping") {
			ui.setStatus("status.stopped")
		}
	})
}

// GetSelectedOptions 获取所有选中的测试选项
func (ui *TestUI) GetSelectedOptions() map[string]bool {
	return map[string]bool{
		"basic":     ui.BasicCheck.Checked,
		"cpu":       ui.CpuCheck.Checked,
		"memory":    ui.MemoryCheck.Checked,
		"disk":      ui.DiskCheck.Checked,
		"unlock":    ui.UnlockCheck.Checked,
		"security":  ui.SecurityCheck.Checked,
		"email":     ui.EmailCheck.Checked,
		"backtrace": ui.BacktraceCheck.Checked,
		"nt3":       ui.Nt3Check.Checked,
		"speed":     ui.SpeedCheck.Checked,
		"ping":      ui.PingCheck.Checked,
	}
}

func (ui *TestUI) collectExecutionConfig() ExecutionConfig {
	language := "zh"
	if ui.selectedLanguageCode() == langEN {
		language = "en"
	}

	cpuMethod := ui.CpuMethodSelect.Selected
	if cpuMethod == "" {
		cpuMethod = "sysbench"
	}

	threadMode := ui.ThreadModeSelect.Selected
	if threadMode == "" {
		threadMode = "multi"
	}

	memoryMethod := ui.MemoryMethodSelect.Selected
	if memoryMethod == "" {
		memoryMethod = "auto"
	}

	diskMethod := ui.DiskMethodSelect.Selected
	if diskMethod == "" {
		diskMethod = "auto"
	}

	nt3Location := ui.Nt3LocationSelect.Selected
	if nt3Location == "" {
		nt3Location = "GZ"
	}

	nt3Type := ui.Nt3TypeSelect.Selected
	if nt3Type == "" {
		nt3Type = "ipv4"
	}

	spNum := 2
	if ui.SpNumEntry.Text != "" {
		if parsed, err := strconv.Atoi(ui.SpNumEntry.Text); err == nil && parsed > 0 {
			spNum = parsed
		}
	}
	if spNum > 20 {
		spNum = 20
	}

	logEnabled := false
	if ui.LogCheck != nil {
		logEnabled = ui.LogCheck.Checked
	}

	pingTgdc := ui.PingTgdcCheck.Checked
	pingWeb := ui.PingWebCheck.Checked

	unlockRegion := unlockRegionLabelToCode(ui.UnlockRegionSelect.Selected, language)
	if unlockRegion == "" {
		unlockRegion = "0"
	}
	unlockIpVersion := ui.UnlockIpVersionSelect.Selected
	if unlockIpVersion == "" {
		unlockIpVersion = "auto"
	}

	return ExecutionConfig{
		SelectedOptions:  ui.GetSelectedOptions(),
		Language:         language,
		TestUpload:       ui.SpTestUploadCheck.Checked,
		TestDownload:     ui.SpTestDownloadCheck.Checked,
		ChinaModeEnabled: ui.ChinaModeCheck.Checked,
		CpuMethod:        cpuMethod,
		ThreadMode:       threadMode,
		MemoryMethod:     memoryMethod,
		DiskMethod:       diskMethod,
		DiskPath:         ui.DiskPathEntry.Text,
		DiskMulti:        ui.DiskMultiCheck.Checked,
		Nt3Location:      nt3Location,
		Nt3Type:          nt3Type,
		SpNum:            spNum,
		PingTgdc:         pingTgdc,
		PingWeb:          pingWeb,
		UnlockRegion:     unlockRegion,
		UnlockIpVersion:  unlockIpVersion,
		LogEnabled:       logEnabled,
	}
}
