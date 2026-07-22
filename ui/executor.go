package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	unlockexecutor "github.com/oneclickvirt/UnlockTests/executor"
	unlocktestmodel "github.com/oneclickvirt/UnlockTests/model"
	backtracemodel "github.com/oneclickvirt/backtrace/model"
	basicmodel "github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/utils"
	cputestmodel "github.com/oneclickvirt/cputest/model"
	disktestmodel "github.com/oneclickvirt/disktest/disk"
	"github.com/oneclickvirt/ecs-gui/internal/appmeta"
	gostunmodel "github.com/oneclickvirt/gostun/model"
	memorytestmodel "github.com/oneclickvirt/memorytest/memory"
	nt3model "github.com/oneclickvirt/nt3/model"
	ptmodel "github.com/oneclickvirt/pingtest/model"
	"github.com/oneclickvirt/pingtest/pt"
	"github.com/oneclickvirt/portchecker/email"
	speedtestmodel "github.com/oneclickvirt/speedtest/model"
)

var (
	ecsVersion        = appmeta.UpstreamECSVersion
	commandExecutorMu sync.Mutex
)

type CommandExecutor struct {
	outputCallback   func(string)
	progressCallback func(ProgressUpdate)
	core             CoreRunner
	ctx              context.Context
	cancel           context.CancelFunc
	structuredMu     sync.RWMutex
	structuredResult *StructuredRunResult
}

func NewCommandExecutor(outputCallback func(string)) *CommandExecutor {
	return &CommandExecutor{
		outputCallback: outputCallback,
		core:           ecsCoreRunner{},
	}
}

// SetContext 设置执行上下文
func (e *CommandExecutor) SetContext(ctx context.Context) {
	e.ctx = ctx
}

func (e *CommandExecutor) SetProgressCallback(callback func(ProgressUpdate)) {
	e.progressCallback = callback
}

func (e *CommandExecutor) SetCoreRunner(core CoreRunner) {
	if core != nil {
		e.core = core
	}
}

func (e *CommandExecutor) StructuredResult() (*StructuredRunResult, bool) {
	e.structuredMu.RLock()
	defer e.structuredMu.RUnlock()
	if e.structuredResult == nil {
		return nil, false
	}
	copy := *e.structuredResult
	return &copy, true
}

func (e *CommandExecutor) setStructuredResult(result StructuredRunResult) {
	e.structuredMu.Lock()
	e.structuredResult = &result
	e.structuredMu.Unlock()
}

type progressTracker struct {
	callback func(ProgressUpdate)
	steps    []string
	current  int
	started  map[string]bool
	finished map[string]bool
}

func newProgressTracker(callback func(ProgressUpdate), steps []string) *progressTracker {
	return &progressTracker{callback: callback, steps: steps, started: make(map[string]bool), finished: make(map[string]bool)}
}

func (t *progressTracker) start(itemKey string) {
	if t == nil {
		return
	}
	if t.started[itemKey] || t.finished[itemKey] {
		return
	}
	t.started[itemKey] = true
	if t.callback == nil {
		return
	}
	total := len(t.steps)
	if total == 0 {
		total = 1
	}
	t.callback(ProgressUpdate{
		ItemKey:  itemKey,
		Current:  t.current + 1,
		Total:    total,
		Fraction: float64(t.current) / float64(total),
	})
}

func (t *progressTracker) finish(itemKey string) {
	if t == nil {
		return
	}
	if t.finished[itemKey] {
		return
	}
	t.finished[itemKey] = true
	if t.callback == nil {
		return
	}
	total := len(t.steps)
	if total == 0 {
		total = 1
	}
	if t.current < total {
		t.current++
	}
	t.callback(ProgressUpdate{
		ItemKey:  itemKey,
		Current:  t.current,
		Total:    total,
		Fraction: float64(t.current) / float64(total),
	})
}

func (t *progressTracker) run(itemKey string, fn func() error) error {
	t.start(itemKey)
	if err := fn(); err != nil {
		return err
	}
	t.finish(itemKey)
	return nil
}

func buildProgressSteps(config ExecutionConfig, connected bool) []string {
	selected := config.SelectedOptions
	pingEnabled := selected["ping"]
	pingTgdc := config.PingTgdc
	pingWeb := config.PingWeb
	unlockEnabled := selected["unlock"]

	if config.ChinaModeEnabled {
		unlockEnabled = false
		pingEnabled = true
		pingTgdc = false
		pingWeb = false
	}

	steps := []string{"progress.precheck"}
	if selected["basic"] || selected["security"] {
		steps = append(steps, "progress.basic_security")
	}
	if selected["cpu"] {
		steps = append(steps, "progress.cpu")
	}
	if selected["memory"] {
		steps = append(steps, "progress.memory")
	}
	if selected["disk"] {
		steps = append(steps, "progress.disk")
	}
	if config.DeepMode {
		steps = append(steps, "progress.deep_hardware")
	}
	if connected && unlockEnabled {
		steps = append(steps, "progress.unlock")
	}
	if connected && selected["security"] {
		steps = append(steps, "progress.ip_quality")
	}
	if connected && selected["email"] {
		steps = append(steps, "progress.email")
	}
	if connected && selected["backtrace"] {
		steps = append(steps, "progress.backtrace")
	}
	if connected && selected["nt3"] {
		steps = append(steps, "progress.nt3")
	}
	if connected && pingEnabled {
		steps = append(steps, "progress.ping")
	}
	if connected && pingTgdc {
		steps = append(steps, "progress.tgdc")
	}
	if connected && pingWeb {
		steps = append(steps, "progress.web")
	}
	if connected && selected["speed"] {
		steps = append(steps, "progress.speed")
	}
	if config.AnalyzeResult {
		steps = append(steps, "progress.summary")
	}
	if connected && config.EnableUpload {
		steps = append(steps, "progress.upload")
	}
	if connected {
		steps = append(steps, "progress.nat", "progress.tcp")
	}
	steps = append(steps, "progress.finish")
	return steps
}

func (e *CommandExecutor) Execute(config ExecutionConfig) (runErr error) {
	commandExecutorMu.Lock()
	defer commandExecutorMu.Unlock()

	e.structuredMu.Lock()
	e.structuredResult = nil
	e.structuredMu.Unlock()
	if e.core == nil {
		e.core = ecsCoreRunner{}
	}
	e.resetDetectedNetworkState()

	// 设置测试选项
	selectedOptions := config.SelectedOptions
	language := config.Language
	chinaModeEnabled := config.ChinaModeEnabled
	width := config.OutputWidth
	if width <= 0 {
		width = 82
	}
	setComponentLogging(config.LogEnabled)

	basicStatus := selectedOptions["basic"]
	cpuTestStatus := selectedOptions["cpu"]
	memoryTestStatus := selectedOptions["memory"]
	diskTestStatus := selectedOptions["disk"]
	utTestStatus := selectedOptions["unlock"]
	securityTestStatus := selectedOptions["security"]
	emailTestStatus := selectedOptions["email"]
	backtraceStatus := selectedOptions["backtrace"]
	nt3Status := selectedOptions["nt3"]
	speedTestStatus := selectedOptions["speed"]
	pingTestStatus := selectedOptions["ping"]
	pingTgdc := config.PingTgdc
	pingWeb := config.PingWeb
	effectiveNt3Type := normalizeNT3Type(config.Nt3Type)

	// 中国模式逻辑：禁用流媒体测试，启用PING测试（只测三网PING）
	// 对齐主仓库逻辑：中国模式下强制启用ping，但不测TGDC和Web
	if chinaModeEnabled {
		utTestStatus = false
		pingTestStatus = true
		// 中国模式下强制禁用TGDC和Web测试
		pingTgdc = false
		pingWeb = false
	}

	if e.progressCallback != nil {
		e.progressCallback(ProgressUpdate{
			ItemKey:  "progress.precheck",
			Current:  1,
			Total:    1,
			Fraction: 0.02,
		})
	}

	// 检查网络连接
	preCheck := utils.CheckPublicAccess(3 * time.Second)
	effectiveNt3Type = effectiveNT3TypeForStack(effectiveNt3Type, preCheck.StackType)
	tracker := newProgressTracker(e.progressCallback, buildProgressSteps(config, preCheck.Connected))
	tracker.finish("progress.precheck")

	// 初始化变量
	var (
		wg1, wg2                                       sync.WaitGroup
		ipv4, ipv6, basicInfo, securityInfo, emailInfo string
		mediaInfo                                      string
		outputMutex                                    sync.Mutex
		captureMutex                                   sync.Mutex
		captured                                       strings.Builder
		captureTruncated                               bool
	)
	startTime := time.Now()
	defer func() {
		ctx := e.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		e.setStructuredResult(buildGUIStructuredReport(config, preCheck.Connected, tracker, runErr, ctx, startTime, time.Now()))
	}()
	captureLimit := resultCaptureLimit()
	appendCaptured := func(text string) {
		cleanText := ansiRegex.ReplaceAllString(text, "")
		captureMutex.Lock()
		defer captureMutex.Unlock()
		if captureLimit <= 0 || len(cleanText) == 0 {
			return
		}
		remaining := captureLimit - captured.Len()
		if remaining <= 0 {
			captureTruncated = true
			return
		}
		if len(cleanText) > remaining {
			captured.WriteString(cleanText[:remaining])
			captureTruncated = true
			return
		}
		captured.WriteString(cleanText)
	}
	capturedOutput := func() string {
		captureMutex.Lock()
		defer captureMutex.Unlock()
		out := captured.String()
		if captureTruncated {
			out += "\n[结果过长，GUI 已截断用于上传/分析的历史输出]\n"
		}
		return out
	}

	// 确保有上下文
	if e.ctx == nil {
		e.ctx = context.Background()
	}

	if needsNetworkIdentityProbe(preCheck.Connected, basicStatus, securityTestStatus, utTestStatus, backtraceStatus) {
		if checkCancelledWithContext(e.ctx) {
			return fmt.Errorf("测试已取消")
		}
		ipv4, ipv6, _ = OnlyBasicsIpInfo(language)
		e.syncDetectedIPs(ipv4, ipv6)
	}

	// 重定向输出到回调
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("创建管道失败: %v", err)
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		defer r.Close()

		buf := make([]byte, 8192) // 增加缓冲区大小
		var partial string        // 用于保存不完整的行
		for {
			n, err := r.Read(buf)
			if n > 0 {
				text := partial + string(buf[:n])
				// 找到最后一个换行符
				lastNewline := strings.LastIndex(text, "\n")
				if lastNewline >= 0 {
					// 输出完整的行
					complete := text[:lastNewline+1]
					if e.outputCallback != nil {
						e.outputCallback(complete)
					}
					appendCaptured(complete)
					// 保存不完整的部分
					partial = text[lastNewline+1:]
				} else {
					// 没有换行符，全部保存为不完整部分
					partial = text
				}
			}
			if err != nil {
				if err == io.EOF {
					// 输出剩余的不完整部分
					if partial != "" {
						if e.outputCallback != nil {
							e.outputCallback(partial)
						}
						appendCaptured(partial)
					}
					return
				}
				if e.outputCallback != nil {
					msg := fmt.Sprintf("\n读取输出失败: %v\n", err)
					e.outputCallback(msg)
					appendCaptured(msg)
				}
				return
			}
		}
	}()

	// 延迟恢复 stdout 和清理资源
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		w.Close()

		// 等待reader goroutine完成，但不要无限等待
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			_ = r.Close()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
			}
		}

	}()

	os.Stdout = w
	os.Stderr = w

	// 检查取消的辅助函数
	checkCancelled := func() bool {
		select {
		case <-e.ctx.Done():
			return true
		default:
			return false
		}
	}

	// 执行测试（参考原goecs.go的runChineseTests和runEnglishTests顺序）
	// 1. 打印头部和基本信息
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if basicStatus || securityTestStatus {
		tracker.start("progress.basic_security")
		outputMutex.Lock()
		PrintHead(language, width, ecsVersion)
		if basicStatus {
			if language == "zh" {
				PrintCenteredTitle("系统基础信息", width)
			} else {
				PrintCenteredTitle("System-Basic-Information", width)
			}
		}
		// 根据网络连接状态选择检测类型
		checkType := effectiveNt3Type
		var adjustedNt3Type string
		if preCheck.Connected && preCheck.StackType == "DualStack" {
			ipv4, ipv6, basicInfo, securityInfo, adjustedNt3Type = BasicsAndSecurityCheck(language, checkType, securityTestStatus)
		} else if preCheck.Connected && preCheck.StackType == "IPv4" {
			ipv4, ipv6, basicInfo, securityInfo, adjustedNt3Type = BasicsAndSecurityCheck(language, "ipv4", securityTestStatus)
		} else if preCheck.Connected && preCheck.StackType == "IPv6" {
			ipv4, ipv6, basicInfo, securityInfo, adjustedNt3Type = BasicsAndSecurityCheck(language, "ipv6", securityTestStatus)
		} else {
			ipv4, ipv6, basicInfo, securityInfo, adjustedNt3Type = BasicsAndSecurityCheck(language, "", false)
			securityTestStatus = false
		}
		e.syncDetectedIPs(ipv4, ipv6)
		if adjustedNt3Type != "" {
			effectiveNt3Type = normalizeNT3Type(adjustedNt3Type)
		}
		if basicStatus {
			fmt.Printf("%s", basicInfo)
		}
		outputMutex.Unlock()
		tracker.finish("progress.basic_security")
	}

	// 2. CPU测试
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if cpuTestStatus {
		tracker.start("progress.cpu")
		outputMutex.Lock()
		realTestMethod, res := e.core.CpuTest(language, config.CpuMethod, config.ThreadMode)
		if language == "zh" {
			PrintCenteredTitle(fmt.Sprintf("CPU测试-通过%s测试", realTestMethod), width)
		} else {
			PrintCenteredTitle(fmt.Sprintf("CPU-Test--%s-Method", realTestMethod), width)
		}
		fmt.Print(res)
		outputMutex.Unlock()
		tracker.finish("progress.cpu")
	}

	// 3. 内存测试
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if memoryTestStatus {
		tracker.start("progress.memory")
		outputMutex.Lock()
		realTestMethod, res := e.core.MemoryTest(language, config.MemoryMethod)
		if language == "zh" {
			PrintCenteredTitle(fmt.Sprintf("内存测试-通过%s测试", realTestMethod), width)
		} else {
			PrintCenteredTitle(fmt.Sprintf("Memory-Test--%s-Method", realTestMethod), width)
		}
		fmt.Print(res)
		outputMutex.Unlock()
		tracker.finish("progress.memory")
	}

	// 4. 磁盘测试
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if diskTestStatus {
		tracker.start("progress.disk")
		outputMutex.Lock()
		if config.AutoDiskMethod {
			realTestMethod, res := e.core.DiskTest(language, config.DiskMethod, config.DiskPath, config.DiskMulti, true)
			if language == "zh" {
				PrintCenteredTitle(fmt.Sprintf("硬盘测试-通过%s测试", realTestMethod), width)
			} else {
				PrintCenteredTitle(fmt.Sprintf("Disk-Test--%s-Method", realTestMethod), width)
			}
			fmt.Print(res)
		} else {
			if language == "zh" {
				PrintCenteredTitle("硬盘测试-通过dd测试", width)
			} else {
				PrintCenteredTitle("Disk-Test--dd-Method", width)
			}
			_, res := e.core.DiskTest(language, "dd", config.DiskPath, config.DiskMulti, false)
			fmt.Print(res)
			if language == "zh" {
				PrintCenteredTitle("硬盘测试-通过fio测试", width)
			} else {
				PrintCenteredTitle("Disk-Test--fio-Method", width)
			}
			_, res = e.core.DiskTest(language, "fio", config.DiskPath, config.DiskMulti, false)
			fmt.Print(res)
		}
		outputMutex.Unlock()
		tracker.finish("progress.disk")
	}

	// 5. 启动异步测试（流媒体解锁和邮件端口）
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if utTestStatus && preCheck.Connected {
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			// 检查取消
			if !checkCancelled() {
				mediaInfo = e.core.MediaTest(language, config.UnlockRegion, config.UnlockIpVersion, config.UnlockShowIP)
			}
		}()
	}

	if emailTestStatus && preCheck.Connected {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			// 检查取消
			if !checkCancelled() {
				emailInfo = email.EmailCheck()
			}
		}()
	}

	// 6. 显示跨国流媒体解锁结果
	if utTestStatus && preCheck.Connected {
		tracker.start("progress.unlock")
		// 使用带超时的等待
		waitDone := make(chan struct{})
		go func() {
			wg1.Wait()
			close(waitDone)
		}()

		select {
		case <-waitDone:
			// 正常完成
		case <-e.ctx.Done():
			// 被取消
			return fmt.Errorf("测试已取消")
		case <-time.After(5 * time.Minute):
			// 超时
			mediaInfo = "\n流媒体测试超时\n"
		}
		outputMutex.Lock()
		if language == "zh" {
			PrintCenteredTitle("跨国流媒体解锁", width)
		} else {
			PrintCenteredTitle("Cross-Border-Streaming-Media-Unlock", width)
		}
		fmt.Printf("%s", mediaInfo)
		outputMutex.Unlock()
		tracker.finish("progress.unlock")
	}

	// 8. 显示IP质量检测结果
	if securityTestStatus && preCheck.Connected {
		tracker.start("progress.ip_quality")
		outputMutex.Lock()
		if language == "zh" {
			PrintCenteredTitle("IP质量检测", width)
		} else {
			PrintCenteredTitle("IP-Quality-Check", width)
		}
		fmt.Printf("%s", securityInfo)
		outputMutex.Unlock()
		tracker.finish("progress.ip_quality")
	}

	// 9. 显示邮件端口测试结果
	if emailTestStatus && preCheck.Connected {
		tracker.start("progress.email")
		// 使用带超时的等待
		waitDone := make(chan struct{})
		go func() {
			wg2.Wait()
			close(waitDone)
		}()

		select {
		case <-waitDone:
			// 正常完成
		case <-e.ctx.Done():
			// 被取消
			return fmt.Errorf("测试已取消")
		case <-time.After(3 * time.Minute):
			// 超时
			emailInfo = "\n邮件端口测试超时\n"
		}
		outputMutex.Lock()
		if language == "zh" {
			PrintCenteredTitle("邮件端口检测", width)
		} else {
			PrintCenteredTitle("Email-Port-Check", width)
		}
		fmt.Println(emailInfo)
		outputMutex.Unlock()
		tracker.finish("progress.email")
	}

	// 10. 上游及回程线路检测
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if backtraceStatus && preCheck.Connected {
		tracker.start("progress.backtrace")
		outputMutex.Lock()
		if language == "zh" {
			PrintCenteredTitle("上游及回程线路检测", width)
		} else {
			PrintCenteredTitle("Upstreams-Backtrace-Check", width)
		}
		e.core.UpstreamsCheck(language)
		outputMutex.Unlock()
		tracker.finish("progress.backtrace")
	}

	// 11. 三网回程路由检测
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if nt3Status && preCheck.Connected {
		tracker.start("progress.nt3")
		outputMutex.Lock()
		if language == "zh" {
			PrintCenteredTitle("三网回程路由检测", width)
		} else {
			PrintCenteredTitle("NextTrace-3Networks-Check", width)
		}
		e.core.NextTrace3Check(language, config.Nt3Location, effectiveNt3Type)
		outputMutex.Unlock()
		tracker.finish("progress.nt3")
	}

	// 12. PING值测试
	// 对齐主仓库逻辑：
	// - 中国模式(chinaModeEnabled)下：只测三网PING，不测TGDC和Web
	// - 非中国模式且pingTestStatus=true：根据用户配置决定
	// - 单独的pingTgdc/pingWeb可以在没有pingTestStatus的情况下也显示
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if pingTestStatus && preCheck.Connected {
		tracker.start("progress.ping")
		outputMutex.Lock()

		// 判断是否为中国模式
		if chinaModeEnabled {
			// 中国模式：只测三网PING
			if language == "zh" {
				PrintCenteredTitle("PING值检测", width)
			} else {
				PrintCenteredTitle("PING-Test", width)
			}
			pingResult := runPingProfile(config, language)
			fmt.Println(pingResult)
		} else {
			// 非中国模式：根据配置测试
			if language == "zh" {
				PrintCenteredTitle("PING值检测", width)
			} else {
				PrintCenteredTitle("PING-Test", width)
			}
			pingResult := runPingProfile(config, language)
			fmt.Println(pingResult)

			// 根据用户配置决定是否测试TGDC和Web
			if pingTgdc {
				fmt.Println(pt.TelegramDCTest())
			}
			if pingWeb {
				fmt.Println(pt.WebsiteTest())
			}
		}

		outputMutex.Unlock()
		tracker.finish("progress.ping")
	}

	// 单独的TGDC和Web测试（当pingTestStatus=false但用户单独启用时）
	if !pingTestStatus && preCheck.Connected && (pingTgdc || pingWeb) {
		tracker.start("progress.ping")
		outputMutex.Lock()
		if language == "zh" {
			PrintCenteredTitle("PING值检测", width)
		} else {
			PrintCenteredTitle("PING-Test", width)
		}

		if pingTgdc {
			fmt.Println(pt.TelegramDCTest())
		}
		if pingWeb {
			fmt.Println(pt.WebsiteTest())
		}

		outputMutex.Unlock()
		tracker.finish("progress.ping")
	}

	// 13. 速度测试
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if speedTestStatus && preCheck.Connected {
		tracker.start("progress.speed")
		outputMutex.Lock()
		if language == "zh" {
			PrintCenteredTitle("就近节点测速", width)
		} else {
			PrintCenteredTitle("Speed-Test", width)
		}
		e.core.SpeedTestShowHead(language)
		runSpeedProfile(e.core, config, language)
		outputMutex.Unlock()
		tracker.finish("progress.speed")
	}

	// 打印时间信息
	outputMutex.Lock()
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60
	currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
	PrintCenteredTitle("", width)
	if language == "zh" {
		fmt.Printf("花费          : %d 分 %d 秒\n", minutes, seconds)
		fmt.Printf("时间          : %s\n", currentTime)
	} else {
		fmt.Printf("Cost    Time          : %d min %d sec\n", minutes, seconds)
		fmt.Printf("Current Time          : %s\n", currentTime)
	}
	PrintCenteredTitle("", width)
	outputMutex.Unlock()

	time.Sleep(200 * time.Millisecond)
	if config.AnalyzeResult {
		tracker.start("progress.summary")
		if summary := BuildResultSummary(language, capturedOutput()); strings.TrimSpace(summary) != "" {
			outputMutex.Lock()
			fmt.Print(summary)
			if !strings.HasSuffix(summary, "\n") {
				fmt.Println()
			}
			outputMutex.Unlock()
			time.Sleep(100 * time.Millisecond)
		}
		tracker.finish("progress.summary")
	}

	if config.EnableUpload && preCheck.Connected {
		tracker.start("progress.upload")
		outputMutex.Lock()
		uploadPath := safeUploadFilePath(config.FilePath)
		uploadConfig := e.core.NewConfig(ecsVersion)
		uploadConfig.Language = language
		uploadConfig.FilePath = uploadPath
		uploadConfig.EnableUpload = true
		e.core.HandleUploadResults(uploadConfig, capturedOutput())
		_ = os.Remove(uploadPath)
		outputMutex.Unlock()
		tracker.finish("progress.upload")
	}

	tracker.start("progress.finish")
	tracker.finish("progress.finish")

	return nil
}

func resultCaptureLimit() int {
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		return 2 * 1024 * 1024
	}
	return 8 * 1024 * 1024
}

func normalizeNT3Type(checkType string) string {
	switch strings.ToLower(strings.TrimSpace(checkType)) {
	case "ipv6":
		return "ipv6"
	case "both":
		return "both"
	default:
		return "both"
	}
}

func effectiveNT3TypeForStack(requested, stack string) string {
	requested = normalizeNT3Type(requested)
	switch stack {
	case "IPv4":
		return "ipv4"
	case "IPv6":
		return "ipv6"
	default:
		return requested
	}
}

func needsNetworkIdentityProbe(connected, basicStatus, securityStatus, unlockStatus, backtraceStatus bool) bool {
	if !connected || basicStatus || securityStatus {
		return false
	}
	return unlockStatus || backtraceStatus
}

func checkCancelledWithContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func (e *CommandExecutor) resetDetectedNetworkState() {
	e.syncDetectedIPs("", "")
}

func (e *CommandExecutor) syncDetectedIPs(ipv4, ipv6 string) {
	ipv4 = strings.TrimSpace(ipv4)
	ipv6 = strings.TrimSpace(ipv6)
	if e.core != nil {
		e.core.SetIPv4Address(ipv4)
		e.core.SetIPv6Address(ipv6)
	}
	unlockexecutor.IPV4 = ipv4 != ""
	unlockexecutor.IPV6 = ipv6 != ""
}

func safeUploadFilePath(input string) string {
	base := sanitizeUploadFileName(input)
	return filepath.Join(os.TempDir(), fmt.Sprintf("goecs-gui-%d-%s", os.Getpid(), base))
}

func sanitizeUploadFileName(input string) string {
	const fallback = "goecs.md"

	name := strings.TrimSpace(strings.ReplaceAll(input, "\\", "/"))
	name = filepath.Base(name)
	if name == "" || name == "." || name == ".." || strings.HasPrefix(name, ".") || isSensitiveUploadFileName(name) {
		return fallback
	}

	var b strings.Builder
	lastDash := false
	for _, r := range name {
		allowed := (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '.' || r == '_' || r == '-' || r == ' '
		if allowed {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}

	name = strings.Trim(b.String(), " .-_")
	if name == "" {
		return fallback
	}

	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".md", ".txt":
	case "":
		name += ".md"
	default:
		name = strings.TrimSuffix(name, filepath.Ext(name)) + ".md"
	}

	if len(name) > 96 {
		ext = filepath.Ext(name)
		stem := strings.TrimSuffix(name, ext)
		maxStem := 96 - len(ext)
		if maxStem < 1 {
			return fallback
		}
		name = stem[:maxStem] + ext
	}

	return name
}

func isSensitiveUploadFileName(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	stem := strings.TrimSuffix(lower, filepath.Ext(lower))
	switch stem {
	case "passwd", "shadow", "id_rsa", "id_ed25519":
		return true
	}
	for _, marker := range []string{"password", "passwd", "secret", "token", "credential", "keystore"} {
		if strings.Contains(stem, marker) {
			return true
		}
	}
	return false
}

func setComponentLogging(enabled bool) {
	gostunmodel.EnableLoger = enabled
	basicmodel.EnableLoger = enabled
	cputestmodel.EnableLoger = enabled
	memorytestmodel.EnableLoger = enabled
	disktestmodel.EnableLoger = enabled
	unlocktestmodel.EnableLoger = enabled
	ptmodel.EnableLoger = enabled
	backtracemodel.EnableLoger = enabled
	nt3model.EnableLoger = enabled
	speedtestmodel.EnableLoger = enabled
}

func runSpeedProfile(core CoreRunner, config ExecutionConfig, language string) {
	spNum := config.SpNum
	if spNum <= 0 {
		spNum = 2
	}
	switch config.PresetKey {
	case "full":
		if language == "zh" {
			core.SpeedTestNearby()
			core.SpeedTestCustom("net", "global", 2, language)
			core.SpeedTestCustom("net", "cu", spNum, language)
			core.SpeedTestCustom("net", "ct", spNum, language)
			core.SpeedTestCustom("net", "cmcc", spNum, language)
		} else {
			core.SpeedTestCustom("net", "global", 4, language)
		}
	case "minimal", "standard", "network_focus", "unlock_focus":
		if language == "zh" {
			core.SpeedTestNearby()
			core.SpeedTestCustom("net", "other", 1, language)
			core.SpeedTestCustom("net", "cu", 1, language)
			core.SpeedTestCustom("net", "ct", 1, language)
			core.SpeedTestCustom("net", "cmcc", 1, language)
		} else {
			core.SpeedTestCustom("net", "global", 4, language)
		}
	case "network_only":
		core.SpeedTestCustom("net", "global", 11, language)
	default:
		if language == "zh" {
			core.SpeedTestNearby()
			core.SpeedTestCustom("net", "cu", spNum, language)
			core.SpeedTestCustom("net", "ct", spNum, language)
			core.SpeedTestCustom("net", "cmcc", spNum, language)
		} else {
			core.SpeedTestCustom("net", "global", 4, language)
		}
	}
}

func runPingProfile(config ExecutionConfig, language string) string {
	scope := ptmodel.PingScope(config.PingScope)
	if strings.EqualFold(language, "en") {
		scope = ptmodel.PingScopeInternational
	}
	return pt.PingTestWithOptions(pt.PingOptions{
		Language: language,
		Scope:    scope,
		Sort:     ptmodel.PingSort(config.PingSortOrder),
	})
}
