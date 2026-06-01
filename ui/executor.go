package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	unlocktestmodel "github.com/oneclickvirt/UnlockTests/model"
	backtracemodel "github.com/oneclickvirt/backtrace/model"
	basicmodel "github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/utils"
	cputestmodel "github.com/oneclickvirt/cputest/model"
	disktestmodel "github.com/oneclickvirt/disktest/disk"
	ecsapi "github.com/oneclickvirt/ecs/api"
	gostunmodel "github.com/oneclickvirt/gostun/model"
	memorytestmodel "github.com/oneclickvirt/memorytest/memory"
	nt3model "github.com/oneclickvirt/nt3/model"
	ptmodel "github.com/oneclickvirt/pingtest/model"
	"github.com/oneclickvirt/pingtest/pt"
	"github.com/oneclickvirt/portchecker/email"
	speedtestmodel "github.com/oneclickvirt/speedtest/model"
)

const ecsVersion = "v0.1.139"

type CommandExecutor struct {
	outputCallback func(string)
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewCommandExecutor(outputCallback func(string)) *CommandExecutor {
	return &CommandExecutor{
		outputCallback: outputCallback,
	}
}

// SetContext 设置执行上下文
func (e *CommandExecutor) SetContext(ctx context.Context) {
	e.ctx = ctx
}

func (e *CommandExecutor) Execute(config ExecutionConfig) error {
	// 设置测试选项
	selectedOptions := config.SelectedOptions
	language := config.Language
	testUpload := config.TestUpload
	testDownload := config.TestDownload
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

	// 中国模式逻辑：禁用流媒体测试，启用PING测试（只测三网PING）
	// 对齐主仓库逻辑：中国模式下强制启用ping，但不测TGDC和Web
	if chinaModeEnabled {
		utTestStatus = false
		pingTestStatus = true
		// 中国模式下强制禁用TGDC和Web测试
		pingTgdc = false
		pingWeb = false
	}

	// 检查网络连接
	preCheck := utils.CheckPublicAccess(3 * time.Second)

	// 初始化变量
	var (
		wg1, wg2                                      sync.WaitGroup
		basicInfo, securityInfo, emailInfo, mediaInfo string
		outputMutex                                   sync.Mutex
		captureMutex                                  sync.Mutex
		captured                                      strings.Builder
		captureTruncated                              bool
	)
	startTime := time.Now()
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

	// 重定向输出到回调
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("创建管道失败: %v", err)
	}

	done := make(chan struct{})
	readerCtx, readerCancel := context.WithCancel(e.ctx)

	go func() {
		defer close(done)
		defer readerCancel()
		defer r.Close()

		buf := make([]byte, 8192) // 增加缓冲区大小
		var partial string        // 用于保存不完整的行
		for {
			// 检查上下文是否取消
			select {
			case <-readerCtx.Done():
				// 输出剩余内容
				if partial != "" && e.outputCallback != nil {
					e.outputCallback(partial)
					appendCaptured(partial)
				}
				return
			default:
			}

			n, err := r.Read(buf)
			if n > 0 && e.outputCallback != nil {
				text := partial + string(buf[:n])
				// 找到最后一个换行符
				lastNewline := strings.LastIndex(text, "\n")
				if lastNewline >= 0 {
					// 输出完整的行
					complete := text[:lastNewline+1]
					e.outputCallback(complete)
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
					if partial != "" && e.outputCallback != nil {
						e.outputCallback(partial)
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
		w.Close()

		// 等待reader goroutine完成，但不要无限等待
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			readerCancel()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
			}
		}

		os.Stdout = oldStdout
	}()

	os.Stdout = w

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
		checkType := config.Nt3Type
		if preCheck.Connected && preCheck.StackType == "DualStack" {
			_, _, basicInfo, securityInfo, _ = BasicsAndSecurityCheck(language, checkType, securityTestStatus)
		} else if preCheck.Connected && preCheck.StackType == "IPv4" {
			_, _, basicInfo, securityInfo, _ = BasicsAndSecurityCheck(language, "ipv4", securityTestStatus)
		} else if preCheck.Connected && preCheck.StackType == "IPv6" {
			_, _, basicInfo, securityInfo, _ = BasicsAndSecurityCheck(language, "ipv6", securityTestStatus)
		} else {
			_, _, basicInfo, securityInfo, _ = BasicsAndSecurityCheck(language, "", false)
			securityTestStatus = false
		}
		if basicStatus {
			fmt.Printf("%s", basicInfo)
		}
		outputMutex.Unlock()
	}

	// 2. CPU测试
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if cpuTestStatus {
		outputMutex.Lock()
		realTestMethod, res := ecsapi.CpuTest(language, config.CpuMethod, config.ThreadMode)
		if language == "zh" {
			PrintCenteredTitle(fmt.Sprintf("CPU测试-通过%s测试", realTestMethod), width)
		} else {
			PrintCenteredTitle(fmt.Sprintf("CPU-Test--%s-Method", realTestMethod), width)
		}
		fmt.Print(res)
		outputMutex.Unlock()
	}

	// 3. 内存测试
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if memoryTestStatus {
		outputMutex.Lock()
		realTestMethod, res := ecsapi.MemoryTest(language, config.MemoryMethod)
		if language == "zh" {
			PrintCenteredTitle(fmt.Sprintf("内存测试-通过%s测试", realTestMethod), width)
		} else {
			PrintCenteredTitle(fmt.Sprintf("Memory-Test--%s-Method", realTestMethod), width)
		}
		fmt.Print(res)
		outputMutex.Unlock()
	}

	// 4. 磁盘测试
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if diskTestStatus {
		outputMutex.Lock()
		if config.AutoDiskMethod {
			realTestMethod, res := ecsapi.DiskTest(language, config.DiskMethod, config.DiskPath, config.DiskMulti, true)
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
			_, res := ecsapi.DiskTest(language, "dd", config.DiskPath, config.DiskMulti, false)
			fmt.Print(res)
			if language == "zh" {
				PrintCenteredTitle("硬盘测试-通过fio测试", width)
			} else {
				PrintCenteredTitle("Disk-Test--fio-Method", width)
			}
			_, res = ecsapi.DiskTest(language, "fio", config.DiskPath, config.DiskMulti, false)
			fmt.Print(res)
		}
		outputMutex.Unlock()
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
				mediaInfo = ecsapi.MediaTest(language, config.UnlockRegion, config.UnlockIpVersion, config.UnlockShowIP)
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
	}

	// 8. 显示IP质量检测结果
	if securityTestStatus && preCheck.Connected {
		outputMutex.Lock()
		if language == "zh" {
			PrintCenteredTitle("IP质量检测", width)
		} else {
			PrintCenteredTitle("IP-Quality-Check", width)
		}
		fmt.Printf("%s", securityInfo)
		outputMutex.Unlock()
	}

	// 9. 显示邮件端口测试结果
	if emailTestStatus && preCheck.Connected {
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
	}

	// 10. 上游及回程线路检测
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if backtraceStatus && preCheck.Connected {
		outputMutex.Lock()
		if language == "zh" {
			PrintCenteredTitle("上游及回程线路检测", width)
		} else {
			PrintCenteredTitle("Upstreams-Backtrace-Check", width)
		}
		ecsapi.UpstreamsCheck(language)
		outputMutex.Unlock()
	}

	// 11. 三网回程路由检测
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if nt3Status && preCheck.Connected {
		outputMutex.Lock()
		if language == "zh" {
			PrintCenteredTitle("三网回程路由检测", width)
		} else {
			PrintCenteredTitle("NextTrace-3Networks-Check", width)
		}
		ecsapi.NextTrace3Check(language, config.Nt3Location, config.Nt3Type)
		outputMutex.Unlock()
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
		outputMutex.Lock()

		// 判断是否为中国模式
		if chinaModeEnabled {
			// 中国模式：只测三网PING
			if language == "zh" {
				PrintCenteredTitle("PING值检测", width)
			} else {
				PrintCenteredTitle("PING-Test", width)
			}
			pingResult := pt.PingTest()
			fmt.Println(pingResult)
		} else {
			// 非中国模式：根据配置测试
			if language == "zh" {
				PrintCenteredTitle("PING值检测", width)
			} else {
				PrintCenteredTitle("PING-Test", width)
			}
			pingResult := pt.PingTest()
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
	}

	// 单独的TGDC和Web测试（当pingTestStatus=false但用户单独启用时）
	if !pingTestStatus && preCheck.Connected && (pingTgdc || pingWeb) {
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
	}

	// 13. 速度测试
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if speedTestStatus && preCheck.Connected {
		outputMutex.Lock()
		if language == "zh" {
			PrintCenteredTitle("就近节点测速", width)
		} else {
			PrintCenteredTitle("Speed-Test", width)
		}
		ecsapi.SpeedTestShowHead(language)

		// 根据上传/下载配置进行测试
		if testUpload || testDownload {
			runSpeedProfile(config, language)
		}
		outputMutex.Unlock()
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
		if summary := BuildResultSummary(language, capturedOutput()); strings.TrimSpace(summary) != "" {
			outputMutex.Lock()
			fmt.Print(summary)
			if !strings.HasSuffix(summary, "\n") {
				fmt.Println()
			}
			outputMutex.Unlock()
			time.Sleep(100 * time.Millisecond)
		}
	}

	if config.EnableUpload && preCheck.Connected {
		outputMutex.Lock()
		uploadConfig := ecsapi.NewConfig(ecsVersion)
		uploadConfig.Language = language
		uploadConfig.FilePath = config.FilePath
		uploadConfig.EnableUpload = true
		ecsapi.HandleUploadResults(uploadConfig, capturedOutput())
		outputMutex.Unlock()
	}

	return nil
}

func resultCaptureLimit() int {
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		return 2 * 1024 * 1024
	}
	return 8 * 1024 * 1024
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

func runSpeedProfile(config ExecutionConfig, language string) {
	spNum := config.SpNum
	if spNum <= 0 {
		spNum = 2
	}
	switch config.PresetKey {
	case "full":
		ecsapi.SpeedTestNearby()
		ecsapi.SpeedTestCustom("net", "global", 2, language)
		ecsapi.SpeedTestCustom("net", "cu", spNum, language)
		ecsapi.SpeedTestCustom("net", "ct", spNum, language)
		ecsapi.SpeedTestCustom("net", "cmcc", spNum, language)
	case "minimal", "standard", "network_focus", "unlock_focus":
		if language == "zh" {
			ecsapi.SpeedTestNearby()
			ecsapi.SpeedTestCustom("net", "other", 1, language)
			ecsapi.SpeedTestCustom("net", "cu", 1, language)
			ecsapi.SpeedTestCustom("net", "ct", 1, language)
			ecsapi.SpeedTestCustom("net", "cmcc", 1, language)
		} else {
			ecsapi.SpeedTestCustom("net", "global", 4, language)
		}
	case "network_only":
		ecsapi.SpeedTestCustom("net", "global", 11, language)
	default:
		ecsapi.SpeedTestNearby()
		ecsapi.SpeedTestCustom("net", "cu", spNum, language)
		ecsapi.SpeedTestCustom("net", "ct", spNum, language)
		ecsapi.SpeedTestCustom("net", "cmcc", spNum, language)
	}
}
