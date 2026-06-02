package ui

import ecsapi "github.com/oneclickvirt/ecs/api"

type CoreRunner interface {
	CpuTest(language, method, threadMode string) (string, string)
	MemoryTest(language, method string) (string, string)
	DiskTest(language, method, path string, multi, auto bool) (string, string)
	MediaTest(language, region, ipVersion string, showIP bool) string
	UpstreamsCheck(language string)
	NextTrace3Check(language, location, checkType string)
	SpeedTestShowHead(language string)
	SpeedTestNearby()
	SpeedTestCustom(platform, operator string, num int, language string)
	NewConfig(version string) *ecsapi.Config
	HandleUploadResults(config *ecsapi.Config, output string)
}

type ecsCoreRunner struct{}

func (ecsCoreRunner) CpuTest(language, method, threadMode string) (string, string) {
	return ecsapi.CpuTest(language, method, threadMode)
}

func (ecsCoreRunner) MemoryTest(language, method string) (string, string) {
	return ecsapi.MemoryTest(language, method)
}

func (ecsCoreRunner) DiskTest(language, method, path string, multi, auto bool) (string, string) {
	return ecsapi.DiskTest(language, method, path, multi, auto)
}

func (ecsCoreRunner) MediaTest(language, region, ipVersion string, showIP bool) string {
	return ecsapi.MediaTest(language, region, ipVersion, showIP)
}

func (ecsCoreRunner) UpstreamsCheck(language string) {
	ecsapi.UpstreamsCheck(language)
}

func (ecsCoreRunner) NextTrace3Check(language, location, checkType string) {
	ecsapi.NextTrace3Check(language, location, checkType)
}

func (ecsCoreRunner) SpeedTestShowHead(language string) {
	ecsapi.SpeedTestShowHead(language)
}

func (ecsCoreRunner) SpeedTestNearby() {
	ecsapi.SpeedTestNearby()
}

func (ecsCoreRunner) SpeedTestCustom(platform, operator string, num int, language string) {
	ecsapi.SpeedTestCustom(platform, operator, num, language)
}

func (ecsCoreRunner) NewConfig(version string) *ecsapi.Config {
	return ecsapi.NewConfig(version)
}

func (ecsCoreRunner) HandleUploadResults(config *ecsapi.Config, output string) {
	ecsapi.HandleUploadResults(config, output)
}
