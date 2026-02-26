package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/oneclickvirt/CommonMediaTests/commediatests"
	"github.com/oneclickvirt/basics/utils"
	ecsapi "github.com/oneclickvirt/ecs/api"
	"github.com/oneclickvirt/pingtest/pt"
	"github.com/oneclickvirt/portchecker/email"
)

const ecsVersion = "v0.1.98"

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

func (e *CommandExecutor) Execute(selectedOptions map[string]bool, language string, testUpload bool, testDownload bool, chinaModeEnabled bool,
	cpuMethod, threadMode, memoryMethod, diskMethod, diskPath string, diskMulti bool,
	nt3Location, nt3Type string, spNum int, pingTgdc, pingWeb bool) error {
	// 设置测试选项
	basicStatus := selectedOptions["basic"]
	cpuTestStatus := selectedOptions["cpu"]
	memoryTestStatus := selectedOptions["memory"]
	diskTestStatus := selectedOptions["disk"]
	commTestStatus := selectedOptions["comm"]
	utTestStatus := selectedOptions["unlock"]
	securityTestStatus := selectedOptions["security"]
	emailTestStatus := selectedOptions["email"]
	backtraceStatus := selectedOptions["backtrace"]
	nt3Status := selectedOptions["nt3"]
	speedTestStatus := selectedOptions["speed"]
	pingTestStatus := selectedOptions["ping"]

	// 中国模式逻辑：禁用流媒体测试，启用PING测试（只测三网PING）
	// 对齐主仓库逻辑：中国模式下强制启用ping，但不测TGDC和Web
	if chinaModeEnabled {
		commTestStatus = false
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
		output, tempOutput                            string
		outputMutex                                   sync.Mutex
	)
	startTime := time.Now()
	uploadDone := make(chan bool, 1)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

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

	done := make(chan bool)
	readerCtx, readerCancel := context.WithCancel(e.ctx)

	go func() {
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
				}
				done <- true
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
					e.outputCallback(text[:lastNewline+1])
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
					}
					done <- true
					return
				}
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
			<-done
		}

		os.Stdout = oldStdout
		signal.Stop(sig)
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
		PrintHead(language, 82, ecsVersion)
		if basicStatus {
			if language == "zh" {
				PrintCenteredTitle("系统基础信息", 82)
			} else {
				PrintCenteredTitle("System-Basic-Information", 82)
			}
		}
		// 根据网络连接状态选择检测类型
		checkType := nt3Type
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
		realTestMethod, res := ecsapi.CpuTest(language, cpuMethod, threadMode)
		if language == "zh" {
			PrintCenteredTitle(fmt.Sprintf("CPU测试-通过%s测试", realTestMethod), 82)
		} else {
			PrintCenteredTitle(fmt.Sprintf("CPU-Test--%s-Method", realTestMethod), 82)
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
		realTestMethod, res := ecsapi.MemoryTest(language, memoryMethod)
		if language == "zh" {
			PrintCenteredTitle(fmt.Sprintf("内存测试-通过%s测试", realTestMethod), 82)
		} else {
			PrintCenteredTitle(fmt.Sprintf("Memory-Test--%s-Method", realTestMethod), 82)
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
		realTestMethod, res := ecsapi.DiskTest(language, diskMethod, diskPath, diskMulti, true)
		if language == "zh" {
			PrintCenteredTitle(fmt.Sprintf("硬盘测试-通过%s测试", realTestMethod), 82)
		} else {
			PrintCenteredTitle(fmt.Sprintf("Disk-Test--%s-Method", realTestMethod), 82)
		}
		fmt.Print(res)
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
				mediaInfo = ecsapi.MediaTest(language)
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

	// 6. 御三家流媒体测试（仅中文）
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if commTestStatus && preCheck.Connected && language == "zh" {
		outputMutex.Lock()
		PrintCenteredTitle("御三家流媒体测试", 82)
		commInfo := commediatests.MediaTests(language)
		fmt.Print(commInfo)
		outputMutex.Unlock()
	}

	// 7. 显示跨国流媒体解锁结果
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
			PrintCenteredTitle("跨国流媒体解锁", 82)
		} else {
			PrintCenteredTitle("Cross-Border-Streaming-Media-Unlock", 82)
		}
		fmt.Printf("%s", mediaInfo)
		outputMutex.Unlock()
	}

	// 8. 显示IP质量检测结果
	if securityTestStatus && preCheck.Connected {
		outputMutex.Lock()
		if language == "zh" {
			PrintCenteredTitle("IP质量检测", 82)
		} else {
			PrintCenteredTitle("IP-Quality-Check", 82)
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
			PrintCenteredTitle("邮件端口检测", 82)
		} else {
			PrintCenteredTitle("Email-Port-Check", 82)
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
			PrintCenteredTitle("上游及回程线路检测", 82)
		} else {
			PrintCenteredTitle("Upstreams-Backtrace-Check", 82)
		}
		ecsapi.UpstreamsCheck()
		outputMutex.Unlock()
	}

	// 11. 三网回程路由检测
	if checkCancelled() {
		return fmt.Errorf("测试已取消")
	}

	if nt3Status && preCheck.Connected {
		outputMutex.Lock()
		if language == "zh" {
			PrintCenteredTitle("三网回程路由检测", 82)
		} else {
			PrintCenteredTitle("NextTrace-3Networks-Check", 82)
		}
		ecsapi.NextTrace3Check(language, nt3Location, nt3Type)
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
				PrintCenteredTitle("PING值检测", 82)
			} else {
				PrintCenteredTitle("PING-Test", 82)
			}
			pingResult := pt.PingTest()
			fmt.Println(pingResult)
		} else {
			// 非中国模式：根据配置测试
			if language == "zh" {
				PrintCenteredTitle("PING值检测", 82)
			} else {
				PrintCenteredTitle("PING-Test", 82)
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
			PrintCenteredTitle("PING值检测", 82)
		} else {
			PrintCenteredTitle("PING-Test", 82)
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
			PrintCenteredTitle("就近节点测速", 82)
		} else {
			PrintCenteredTitle("Speed-Test", 82)
		}
		ecsapi.SpeedTestShowHead(language)

		// 根据上传/下载配置进行测试
		if testUpload || testDownload {
			ecsapi.SpeedTestNearby()
			if language == "zh" {
				ecsapi.SpeedTestCustom("net", "global", spNum, language)
			} else {
				ecsapi.SpeedTestCustom("net", "global", -1, language)
			}
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
	PrintCenteredTitle("", 82)
	if language == "zh" {
		fmt.Printf("花费          : %d 分 %d 秒\n", minutes, seconds)
		fmt.Printf("时间          : %s\n", currentTime)
	} else {
		fmt.Printf("Cost    Time          : %d min %d sec\n", minutes, seconds)
		fmt.Printf("Current Time          : %s\n", currentTime)
	}
	PrintCenteredTitle("", 82)
	outputMutex.Unlock()

	// 清理
	_ = uploadDone
	_ = sig
	_ = output
	_ = tempOutput

	return nil
}
