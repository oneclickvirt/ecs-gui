package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
)

const (
	structuredOverviewWidth = 52
	structuredOverviewLabel = 20
)

func formatStructuredDetails(report StructuredRunResult, languages ...string) string {
	language := langZH
	if len(languages) > 0 && languages[0] == langEN {
		language = langEN
	}
	zh := language != langEN
	var builder strings.Builder
	overviewSection(&builder, overviewPick(zh, "运行概览", "Run Overview"))
	overviewRow(&builder, overviewPick(zh, "总体状态", "Overall Status"), overviewStatus(report.Status, zh))
	if report.ECSVersion != "" {
		overviewRow(&builder, overviewPick(zh, "版本", "Version"), report.ECSVersion)
	}
	if report.DurationMS > 0 {
		overviewRow(&builder, overviewPick(zh, "用时", "Duration"), overviewDuration(report.DurationMS))
	}
	overviewRow(&builder, overviewPick(zh, "测试模式", "Mode"), overviewMode(report.DeepMode, report.PrivacyMode, zh))

	overviewSection(&builder, overviewPick(zh, "章节进度", "Section Progress"))
	for _, section := range report.Sections {
		if !section.Enabled && section.Status == "skipped" {
			continue
		}
		value := overviewStatus(section.Status, zh)
		if section.Reason != "" {
			value += " | " + section.Reason
		}
		overviewRow(&builder, overviewComponentName(section.Name, zh), value)
	}

	if len(report.DataFiles) > 0 {
		overviewSection(&builder, overviewPick(zh, "数据源状态", "Data Sources"))
		for _, file := range report.DataFiles {
			source := overviewDataSource(file.Source)
			if file.Fallback != "" {
				fallback := overviewDataSource(file.Fallback)
				if strings.EqualFold(source, fallback) {
					source += overviewPick(zh, " (回退)", " (fallback)")
				} else {
					source += " -> " + fallback
				}
			}
			value := fmt.Sprintf("%s | %d | %s", source, file.Count, overviewStatus(file.Status, zh))
			overviewRow(&builder, strings.TrimSuffix(file.File, ".json"), value)
		}
		if !report.DataFiles[0].GeneratedAt.IsZero() {
			overviewRow(&builder, overviewPick(zh, "同步时间", "Synced At"), report.DataFiles[0].GeneratedAt.Local().Format("2006-01-02 15:04:05"))
		}
	}

	overviewSection(&builder, overviewPick(zh, "组件结果", "Component Results"))
	if len(report.Components) == 0 {
		overviewRow(&builder, overviewPick(zh, "组件", "Components"), overviewPick(zh, "无", "none"))
		return builder.String()
	}
	for _, component := range report.Components {
		value := overviewStatus(component.Status, zh)
		if component.DurationMS > 0 {
			value += " | " + overviewDuration(component.DurationMS)
		}
		if component.Reason != "" {
			value += " | " + component.Reason
		}
		overviewRow(&builder, overviewComponentName(component.Name, zh), value)
	}
	return builder.String()
}

func overviewSection(builder *strings.Builder, title string) {
	titleWidth := runewidth.StringWidth(title)
	padding := max(structuredOverviewWidth-titleWidth, 0)
	left := padding / 2
	builder.WriteString(strings.Repeat("-", left))
	builder.WriteString(overviewTruncate(title, structuredOverviewWidth))
	builder.WriteString(strings.Repeat("-", padding-left))
	builder.WriteByte('\n')
}

func overviewRow(builder *strings.Builder, label, value string) {
	label = overviewPad(overviewTruncate(label, structuredOverviewLabel), structuredOverviewLabel)
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if value == "" {
		value = "-"
	}
	builder.WriteString(label)
	builder.WriteString(" : ")
	builder.WriteString(overviewTruncate(value, structuredOverviewWidth-structuredOverviewLabel-3))
	builder.WriteByte('\n')
}

func overviewStatus(status string, zh bool) string {
	if !zh {
		if status == "" {
			return "-"
		}
		return strings.ReplaceAll(status, "_", " ")
	}
	values := map[string]string{
		"ok": "正常", "partial": "部分完成", "unavailable": "不可用", "timeout": "超时",
		"canceled": "已取消", "error": "错误", "skipped": "已跳过",
	}
	if value, ok := values[status]; ok {
		return value
	}
	if status == "" {
		return "-"
	}
	return status
}

func overviewComponentName(name string, zh bool) string {
	names := map[string][2]string{
		"basics": {"系统基础信息", "System Basics"}, "cpu": {"CPU测试", "CPU"}, "cputest": {"CPU测试", "CPU"},
		"memory": {"内存测试", "Memory"}, "memorytest": {"内存测试", "Memory"}, "disk": {"磁盘测试", "Disk"}, "disktest": {"磁盘测试", "Disk"},
		"media": {"平台解锁", "Media"}, "unlocktests.media": {"平台解锁", "Media"}, "security": {"IP质量", "IP Quality"},
		"security.evidence": {"IP质量", "IP Quality"}, "email": {"邮件端口", "Email"}, "portchecker.email": {"邮件端口", "Email"},
		"backtrace": {"上游及注册信息", "Upstream"}, "backtrace.ip_bgp": {"上游及注册信息", "Upstream"},
		"routes": {"三网路由", "Routes"}, "nt3.province_latency": {"全国三网延迟", "Province Latency"},
		"nt3.province_routes": {"全国详细路由", "Province Routes"}, "ping": {"PING值", "PING"}, "ping.icmp": {"PING值", "PING"},
		"tgdc": {"Telegram DC", "Telegram DC"}, "ping.telegram": {"Telegram DC", "Telegram DC"},
		"web": {"网站连接", "Website TCP"}, "ping.web_tcp": {"网站连接", "Website TCP"}, "tcp": {"TCP连接", "TCP"},
		"speed": {"节点测速", "Speed"}, "speed.registry": {"节点测速", "Speed"}, "nat": {"NAT类型", "NAT"}, "gostun.nat": {"NAT类型", "NAT"},
		"deep_hardware": {"深度硬件", "Deep Hardware"}, "disktest.deep_multi": {"多目录磁盘", "Multi-Path Disk"},
		"basics.smart_selftest": {"SMART自检", "SMART Self-Test"}, "cputest.burn": {"CPU压力", "CPU Burn"},
		"basics.gpu_compute": {"GPU计算", "GPU Compute"},
	}
	if value, ok := names[name]; ok {
		if zh {
			return value[0]
		}
		return value[1]
	}
	return strings.ReplaceAll(name, "_", " ")
}

func overviewDataSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "cdn":
		return "CDN"
	case "raw", "github raw":
		return "GitHub Raw"
	case "embedded":
		return "embedded"
	default:
		return source
	}
}

func overviewMode(deep, privacy, zh bool) string {
	mode := overviewPick(zh, "标准", "standard")
	if deep {
		mode = overviewPick(zh, "深度", "deep")
	}
	if privacy {
		mode += overviewPick(zh, " / 隐私", " / privacy")
	}
	return mode
}

func overviewDuration(milliseconds int64) string {
	if milliseconds <= 0 {
		return "-"
	}
	duration := time.Duration(milliseconds) * time.Millisecond
	if duration >= time.Minute {
		return fmt.Sprintf("%d min %d sec", int(duration.Minutes()), int(duration.Seconds())%60)
	}
	if duration >= time.Second {
		return fmt.Sprintf("%.2f s", duration.Seconds())
	}
	return fmt.Sprintf("%d ms", milliseconds)
}

func overviewPick(zh bool, chinese, english string) string {
	if zh {
		return chinese
	}
	return english
}

func overviewTruncate(value string, width int) string {
	if runewidth.StringWidth(value) <= width {
		return value
	}
	if width <= 3 {
		return runewidth.Truncate(value, width, "")
	}
	return runewidth.Truncate(value, width-3, "") + "..."
}

func overviewPad(value string, width int) string {
	padding := width - runewidth.StringWidth(value)
	if padding <= 0 {
		return value
	}
	return value + strings.Repeat(" ", padding)
}
