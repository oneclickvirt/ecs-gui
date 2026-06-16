package ui

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/oneclickvirt/ecs-gui/utils"
)

// PrintHead 打印标题头
func PrintHead(language string, width int, version string) {
	if language == "zh" {
		PrintCenteredTitle("VPS融合怪测试", width)
		fmt.Printf("版本：%s\n", version)
		fmt.Println("测评频道: https://t.me/+UHVoo2U4VyA5NTQ1\n" +
			"Go项目地址：https://github.com/oneclickvirt/ecs\n" +
			"Shell项目地址：https://github.com/spiritLHLS/ecs")
	} else {
		PrintCenteredTitle("VPS Fusion Monster Test", width)
		fmt.Printf("Version: %s\n", version)
		fmt.Println("Review Channel: https://t.me/+UHVoo2U4VyA5NTQ1\n" +
			"Go Project: https://github.com/oneclickvirt/ecs\n" +
			"Shell Project: https://github.com/spiritLHLS/ecs")
	}
}

// PrintCenteredTitle 打印居中的标题
func PrintCenteredTitle(title string, width int) {
	if title == "" {
		fmt.Println(strings.Repeat("-", width))
		return
	}
	titleLen := runewidth.StringWidth(title)
	if titleLen >= width {
		fmt.Println(title)
		return
	}
	padding := (width - titleLen) / 2
	fmt.Printf("%s%s%s\n", strings.Repeat("-", padding), title, strings.Repeat("-", width-padding-titleLen))
}

func BuildResultSummary(language, output string) string {
	if strings.TrimSpace(output) == "" {
		return ""
	}

	sections := []struct {
		zh   string
		en   string
		keys []string
	}{
		{"基础信息", "Basic", []string{"系统基础信息", "System-Basic-Information"}},
		{"CPU", "CPU", []string{"CPU测试", "CPU-Test"}},
		{"内存", "Memory", []string{"内存测试", "Memory-Test"}},
		{"磁盘", "Disk", []string{"硬盘测试", "Disk-Test"}},
		{"流媒体解锁", "Unlock", []string{"跨国流媒体解锁", "Streaming-Media-Unlock"}},
		{"IP质量", "IP Quality", []string{"IP质量检测", "IP-Quality-Check"}},
		{"邮件端口", "Email", []string{"邮件端口检测", "Email-Port-Check"}},
		{"回程线路", "Backtrace", []string{"上游及回程线路检测", "Upstreams-Backtrace-Check"}},
		{"NT3路由", "NT3", []string{"三网回程路由检测", "NextTrace-3Networks-Check"}},
		{"PING", "Ping", []string{"PING值检测", "PING-Test"}},
		{"测速", "Speed", []string{"就近节点测速", "Speed-Test"}},
	}

	var done []string
	for _, section := range sections {
		for _, key := range section.keys {
			if strings.Contains(output, key) {
				if language == "en" {
					done = append(done, section.en)
				} else {
					done = append(done, section.zh)
				}
				break
			}
		}
	}

	speedRows := strings.Count(output, "Mbps")
	warnings := 0
	for _, mark := range []string{"错误", "失败", "Error", "failed", "timeout", "超时"} {
		warnings += strings.Count(output, mark)
	}

	var b strings.Builder
	if language == "en" {
		b.WriteString("\n-----------------------------Result Summary-----------------------------\n")
		if len(done) > 0 {
			b.WriteString("Completed Sections : " + strings.Join(done, ", ") + "\n")
		}
		b.WriteString(fmt.Sprintf("Speed Rows         : %d\n", speedRows))
		b.WriteString(fmt.Sprintf("Warning Markers    : %d\n", warnings))
		b.WriteString("------------------------------------------------------------------------\n")
		return b.String()
	}

	b.WriteString("\n-----------------------------结果摘要-----------------------------\n")
	if len(done) > 0 {
		b.WriteString("已输出测试段    : " + strings.Join(done, "、") + "\n")
	}
	b.WriteString(fmt.Sprintf("测速结果行      : %d\n", speedRows))
	b.WriteString(fmt.Sprintf("异常提示标记    : %d\n", warnings))
	b.WriteString("------------------------------------------------------------------\n")
	return b.String()
}

// BasicsAndSecurityCheck 执行基础信息和安全检查
func BasicsAndSecurityCheck(language, nt3CheckType string, securityCheckStatus bool) (string, string, string, string, string) {
	return utils.BasicsAndSecurityCheck(language, nt3CheckType, securityCheckStatus)
}

// OnlyBasicsIpInfo detects public IPv4/IPv6 addresses without printing full system details.
func OnlyBasicsIpInfo(language string) (string, string, string) {
	return utils.OnlyBasicsIpInfo(language)
}

// joinStrings joins a slice of strings with the given separator.
func joinStrings(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}
