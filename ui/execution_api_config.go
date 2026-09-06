package ui

import (
	"strings"

	ecsapi "github.com/oneclickvirt/ecs/api"
)

// apiConfigForExecution is shared by both GUI backends so presets retain the
// same upstream scheduling and speed-profile semantics regardless of build
// tags.
func apiConfigForExecution(config ExecutionConfig) *ecsapi.Config {
	selected := config.SelectedOptions
	apiConfig := ecsapi.NewConfig(ecsVersion)
	apiConfig.MenuMode = false
	apiConfig.Choice = upstreamChoiceForPreset(config.PresetKey)
	apiConfig.Language = config.Language
	apiConfig.CpuTestMethod = config.CpuMethod
	apiConfig.CpuTestThreadMode = config.ThreadMode
	apiConfig.MemoryTestMethod = config.MemoryMethod
	apiConfig.DiskTestMethod = config.DiskMethod
	apiConfig.DiskTestPath = config.DiskPath
	apiConfig.DiskMultiCheck = config.DiskMulti
	apiConfig.AutoChangeDiskMethod = config.AutoDiskMethod
	apiConfig.Nt3Location = config.Nt3Location
	apiConfig.Nt3CheckType = config.Nt3Type
	apiConfig.SpNum = config.SpNum
	apiConfig.Width = config.OutputWidth
	apiConfig.BasicStatus = selected["basic"]
	apiConfig.CpuTestStatus = selected["cpu"]
	apiConfig.MemoryTestStatus = selected["memory"]
	apiConfig.DiskTestStatus = selected["disk"]
	apiConfig.UtTestStatus = selected["unlock"] && !config.ChinaModeEnabled
	apiConfig.SecurityTestStatus = selected["security"]
	apiConfig.EmailTestStatus = selected["email"]
	apiConfig.BacktraceStatus = selected["backtrace"] && !config.ChinaModeEnabled
	apiConfig.Nt3Status = selected["nt3"] && !config.ChinaModeEnabled
	apiConfig.SpeedTestStatus = selected["speed"]
	apiConfig.PingTestStatus = selected["ping"] || config.ChinaModeEnabled
	apiConfig.PingSortOrder = config.PingSortOrder
	apiConfig.PingScope = config.PingScope
	apiConfig.TCPSortOrder = config.TCPSortOrder
	apiConfig.TgdcTestStatus = config.PingTgdc && !config.ChinaModeEnabled
	apiConfig.WebTestStatus = config.PingWeb && !config.ChinaModeEnabled
	apiConfig.OnlyChinaTest = config.ChinaModeEnabled
	apiConfig.UnlockTestRegion = config.UnlockRegion
	apiConfig.UnlockTestIPVersion = config.UnlockIpVersion
	apiConfig.UnlockTestShowIP = config.UnlockShowIP
	apiConfig.UnlockTestInterface = config.UnlockInterface
	apiConfig.UnlockTestDNSServers = config.UnlockDNS
	apiConfig.UnlockTestHTTPProxy = config.UnlockHTTPProxy
	apiConfig.UnlockTestSOCKSProxy = config.UnlockSOCKSProxy
	apiConfig.UnlockTestConcurrency = config.UnlockConcurrency
	apiConfig.EnableLogger = config.LogEnabled
	apiConfig.FilePath = config.FilePath
	apiConfig.EnableUpload = config.EnableUpload && !config.PrivacyMode
	apiConfig.AnalyzeResult = config.AnalyzeResult
	apiConfig.PrivacyMode = config.PrivacyMode
	apiConfig.DataOffline = config.DataOffline
	apiConfig.DNSMode = config.DNSMode
	apiConfig.DeepMode = config.DeepMode
	if config.DeepMode {
		apiConfig.DeepDiskPaths = config.DeepDiskPaths
		apiConfig.DeepSMARTDevices = config.DeepSMARTDevices
		apiConfig.DeepBurnDuration = config.DeepBurnDuration
		apiConfig.DeepGPUDevice = config.DeepGPUDevice
	}
	apiConfig.TCPProbeStatus = selected["tcp"]
	apiConfig.MaxDuration = config.MaxDuration
	if apiConfig.MaxDuration < 0 {
		apiConfig.MaxDuration = 0
	}
	apiConfig.HardwareBudget = config.HardwareBudget
	if apiConfig.HardwareBudget < 0 {
		apiConfig.HardwareBudget = 0
	}
	if apiConfig.MaxDuration > 0 && apiConfig.HardwareBudget > apiConfig.MaxDuration {
		apiConfig.HardwareBudget = apiConfig.MaxDuration
	}
	apiConfig.JSONPath = strings.TrimSpace(config.JSONPath)
	return apiConfig
}
