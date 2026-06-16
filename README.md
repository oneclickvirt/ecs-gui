# GOECS GUI 版本

[![Build All UI APP](https://github.com/oneclickvirt/ecs-gui/actions/workflows/build.yml/badge.svg)](https://github.com/oneclickvirt/ecs-gui/actions/workflows/build.yml)

[English](README.en.md)

基于融合怪 Go 版本在 Fyne 框架上开发的跨平台系统测试工具，支持 Android、macOS、Windows 和 Linux 平台。

原项目：https://github.com/oneclickvirt/ecs

## 功能概览

- 图形化选择基础信息、CPU、内存、磁盘、流媒体解锁、路由、PING、测速等测试项
- 运行时显示阶段进度和当前执行项
- 支持结果复制、导出文本文件、上传生成分享链接
- 支持中文/英文界面和深色/浅色主题
- Android 支持测试完成系统通知，Windows 会在需要时请求 UAC 管理员启动

## 快速开始

1. 安装 Go 1.25.4 或更高版本。
2. 克隆仓库并进入 `ecs-gui` 目录。
3. 运行 `go mod download` 下载依赖。
4. 运行 `go run -ldflags="-checklinkname=0" .` 启动开发版。
5. 需要本机产物时运行 `./build.sh desktop`。

## 架构概览

```mermaid
flowchart LR
  User["用户"] --> App["Fyne App<br/>main.go"]
  App --> UI["UI 层<br/>ui/"]
  UI --> Executor["CommandExecutor<br/>进度/取消/输出捕获"]
  Executor --> Core["CoreRunner 接口"]
  Core --> ECS["goecs 与测试模块"]
  Executor --> Terminal["终端结果视图"]
  Terminal --> Export["复制/导出/分享"]
  UI --> Build["FyneApp.toml<br/>build.sh / GitHub Actions"]
```

## 支持平台

### Android
- Action 编译已集成 APK 签名流程
- 最低版本: Android 7.0 (API Level 24)
- 推荐版本: Android 13 (API Level 33) 或更高
- 支持架构: ARM64, x86_64

### macOS
- Action 编译默认未代码签名
- 最低版本: macOS 11.0
- 支持架构: Apple Silicon (ARM64), Intel (AMD64)
- 签名说明: [docs/macos-signing.zh.md](docs/macos-signing.zh.md)

### Windows
- Action 编译默认未代码签名
- 最低版本: Windows 10
- 支持架构: ARM64, AMD64
- 部分磁盘和路由测试需要管理员权限，GUI 会尝试通过 UAC 重新启动

### Linux
- Action 编译默认未代码签名
- 支持架构: ARM64, AMD64

## 本地构建

### 环境要求

1. Go 1.25.4 或更高版本
2. Android SDK (仅用于构建 Android 版本)
3. Android NDK 25.2.9519653 (仅用于构建 Android 版本)
4. JDK 17 或更高版本 (仅用于构建 Android 版本)

### 环境配置

```bash
# 设置 Android NDK 路径 (仅用于构建 Android 版本)
export ANDROID_NDK_HOME=/path/to/android-ndk

# 安装 Fyne CLI
go install fyne.io/tools/cmd/fyne@latest
```

### 构建命令

```bash
# 构建桌面版本 (用于快速测试)
./build.sh desktop

# 构建 Android APK
./build.sh android

# 构建 macOS 应用程序
./build.sh macos

# 构建 Windows 应用程序
./build.sh windows

# 构建 Linux 应用程序
./build.sh linux

# 构建所有平台
./build.sh all
```

构建产物将直接输出到当前目录。

版本来源：

- `VERSION` 保存应用语义版本，例如 `0.1.139`
- `FyneApp.toml` 的 `Version` 应与 `VERSION` 保持一致
- GitHub Actions 发布标签使用 `vYYYYMMDD-HHMMSS`

### 环境变量与 CI Secrets

| 名称 | 位置 | 是否必需 | 用途 |
|------|------|----------|------|
| `ANDROID_NDK_HOME` | 本地/CI | 仅 Android 构建 | 指向 Android NDK 25.2.9519653 |
| `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24` | CI | 是 | 提前启用 Node 24 Actions 运行时 |
| `GOPRIVATE` | CI | 私有模块构建必需 | 允许下载 `github.com/oneclickvirt/security` 等私有模块 |
| `GHT` | GitHub Secret | CI 必需 | 访问私有 Go 模块和签名仓库 |
| `KEYSTORE_PASSWORD` | GitHub Secret | Android 签名必需 | APK keystore 密码 |
| `KEY_PASSWORD` | GitHub Secret | Android 签名必需 | APK key alias 密码 |
| `GITHUB_TOKEN` | GitHub 自动注入 | CI 必需 | 创建 Release 和上传产物 |
| `APPLE_ID` / `APPLE_TEAM_ID` / `APPLE_APP_PASSWORD` | GitHub Secret | macOS 公证可选 | 参考 macOS 签名文档 |
| `MACOS_CERTIFICATE_P12` / `MACOS_CERTIFICATE_PASSWORD` | GitHub Secret | macOS 签名可选 | 参考 macOS 签名文档 |

### 构建产物说明

- Android: APK 安装包
  - `ecs-gui-android-*.apk` - 多架构版本

- macOS: TAR.GZ 压缩包 (包含 .app 应用程序)
  - `goecs-macos-arm64-*.tar.gz` - Apple Silicon 版本
  - `goecs-macos-amd64-*.tar.gz` - Intel 版本

- Windows: EXE 可执行文件
  - `goecs-windows-arm64-*.exe` - ARM64 版本
  - `goecs-windows-amd64-*.exe` - AMD64 版本

- Linux: TAR.XZ 压缩包
  - `goecs-linux-arm64-*.tar.xz` - ARM64 版本
  - `goecs-linux-amd64-*.tar.xz` - AMD64 版本

## 开发调试

```bash
# 克隆仓库
git clone https://github.com/oneclickvirt/ecs-gui.git
cd ecs-gui

# 下载依赖
go mod download

# 运行桌面版本 (用于开发测试)
go run -ldflags="-checklinkname=0" .
```

## FAQ

- 为什么开发运行需要 `-ldflags="-checklinkname=0"`？
  上游依赖包含 `linkname` 用法，当前 Go 工具链需要显式放宽检查。
- 项目是否使用数据库？
  不使用。当前没有 DB 初始化、迁移或连接池逻辑。
- 分享链接有什么限制？
  GUI 会把导出的 Markdown 结果上传到分享服务，单次上传限制为 25KB。
- macOS 为什么会提示未签名？
  默认 Actions 产物未签名；生产分发请按 [macOS 签名文档](docs/macos-signing.zh.md) 配置证书和公证。
