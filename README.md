# GOECS GUI Version

[![Build All UI APP](https://github.com/oneclickvirt/ecs/actions/workflows/build.yml/badge.svg)](https://github.com/oneclickvirt/ecs/actions/workflows/build.yml)

基于 Fyne 框架的跨平台系统测试工具，支持 Android、macOS、Windows 和 Linux 平台。

## 支持平台

### Android
- Action编译已自签证书
- 最低版本: Android 7.0 (API Level 24)
- 推荐版本: Android 13 (API Level 33) 或更高
- 支持架构: ARM64, x86_64

### macOS
- Action编译未自签证书
- 最低版本: macOS 11.0
- 支持架构: Apple Silicon (ARM64), Intel (AMD64)

### Windows
- Action编译未自签证书
- 最低版本: Windows 10
- 支持架构: ARM64, AMD64

### Linux
- Action未编译
- 支持架构: ARM64, AMD64

## 本地构建

### 环境要求

1. Go 1.25.3 或更高版本
2. Android SDK (仅用于构建 Android 版本)
3. Android NDK 25.2.9519653 (仅用于构建 Android 版本)
4. JDK 17 或更高版本 (仅用于构建 Android 版本)

### 环境配置

```bash
# 设置 Android NDK 路径 (仅用于构建 Android 版本)
export ANDROID_NDK_HOME=/path/to/android-ndk

# 安装 Fyne CLI
go install fyne.io/fyne/v2/cmd/fyne@latest
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

### 构建产物说明

- Android: APK 安装包
  - `goecs-android-*.apk` - 多架构版本

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
git clone https://github.com/oneclickvirt/ecs.git
cd ecs

# 切换到 GUI 分支
git checkout gui

# 下载依赖
go mod download

# 运行桌面版本 (用于开发测试)
go run -ldflags="-checklinkname=0" .
```
