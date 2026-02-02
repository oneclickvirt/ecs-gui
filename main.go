package main

import (
	"flag"
	"fmt"
	"os"

	"fyne.io/fyne/v2/app"
	"github.com/oneclickvirt/ecs-gui/ui"
)

var (
	showVersion bool
	showHelp    bool
)

func init() {
	flag.BoolVar(&showVersion, "version", false, "显示版本信息")
	flag.BoolVar(&showVersion, "v", false, "显示版本信息")
	flag.BoolVar(&showHelp, "help", false, "显示帮助信息")
	flag.BoolVar(&showHelp, "h", false, "显示帮助信息")
}

func main() {
	// 禁用 Fyne 的线程警告
	os.Setenv("FYNE_DISABLE_DRIVER_THREAD_CHECK", "1")

	flag.Parse()

	if showVersion {
		fmt.Println("仅支持图形界面模式")
		os.Exit(0)
	}

	if showHelp {
		printHelp()
		os.Exit(0)
	}

	// 启动图形界面
	runGUIMode()
}

func runGUIMode() {
	myApp := app.NewWithID("com.oneclickvirt.goecs")
	myApp.Settings().SetTheme(&ui.CustomTheme{})

	// 设置为浅色主题
	myApp.Settings().SetTheme(&ui.CustomTheme{})

	testUI := ui.NewTestUI(myApp)
	testUI.Window.ShowAndRun()
}

func printHelp() {
	fmt.Println(`说明：
用法:
  ecs-gui                    启动图形界面

选项:
  -version, -v               显示版本信息
  -help, -h                  显示此帮助信息

功能:
  本应用提供图形界面，支持以下测试：
  - 基础信息测试
  - CPU 性能测试
  - 内存性能测试
  - 磁盘性能测试
  - 网络测速
  - 流媒体解锁测试
  - 路由追踪测试

更多信息:
  GitHub: https://github.com/oneclickvirt/ecs
  分支: android-app`)
}
