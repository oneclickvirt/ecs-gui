package main

import (
	"flag"
	"fmt"
	"os"

	"fyne.io/fyne/v2/app"
	"github.com/oneclickvirt/ecs-gui/internal/appmeta"
	"github.com/oneclickvirt/ecs-gui/ui"
)

func main() {
	// 添加全局错误恢复
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "程序发生严重错误: %v\n", r)
			os.Exit(1)
		}
	}()

	showVersion, showHelp, err := parseGUIFlags(os.Args[1:])
	if err != nil {
		os.Exit(2)
	}

	if showVersion {
		fmt.Printf("%s %s (upstream ecs %s)\n", appmeta.AppName, appmeta.Version, appmeta.UpstreamECSVersion)
		os.Exit(0)
	}

	if showHelp {
		printHelp()
		os.Exit(0)
	}

	// 启动图形界面
	runGUIMode()
}

func parseGUIFlags(args []string) (showVersion, showHelp bool, err error) {
	flags := flag.NewFlagSet("ecs-gui", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.BoolVar(&showVersion, "version", false, "显示版本信息")
	flags.BoolVar(&showVersion, "v", false, "显示版本信息")
	flags.BoolVar(&showHelp, "help", false, "显示帮助信息")
	flags.BoolVar(&showHelp, "h", false, "显示帮助信息")
	err = flags.Parse(args)
	return
}

func runGUIMode() {
	myApp := app.NewWithID(appmeta.AppID)
	myApp.SetIcon(appIconResource())

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
  GUI项目: https://github.com/oneclickvirt/ecs-gui
  上游项目: https://github.com/oneclickvirt/ecs`)
}
