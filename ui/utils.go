package ui

import (
	"fmt"
	"strings"

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
	titleLen := len(title)
	if titleLen >= width {
		fmt.Println(title)
		return
	}
	padding := (width - titleLen) / 2
	fmt.Printf("%s%s%s\n", strings.Repeat("-", padding), title, strings.Repeat("-", width-padding-titleLen))
}

// BasicsAndSecurityCheck 执行基础信息和安全检查
func BasicsAndSecurityCheck(language, nt3CheckType string, securityCheckStatus bool) (string, string, string, string, string) {
	return utils.BasicsAndSecurityCheck(language, nt3CheckType, securityCheckStatus)
}
