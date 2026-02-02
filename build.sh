#!/bin/bash

set -e

BUILD_TYPE=${1:-"desktop"}

# 检查 Fyne CLI 是否安装
check_fyne_cli() {
    if ! command -v fyne &> /dev/null; then
        echo "正在安装 Fyne CLI..."
        go install fyne.io/fyne/v2/cmd/fyne@latest
        if [ $? -ne 0 ]; then
            echo "Fyne CLI 安装失败"
            exit 1
        fi
        echo "Fyne CLI 安装成功"
    else
        echo "Fyne CLI 已安装"
    fi
}

# 桌面端构建（用于快速测试）
build_desktop() {
    # 检测当前平台
    local current_os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local current_arch=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
    
    echo "=========================================="
    echo "  构建桌面端应用"
    echo "  平台: ${current_os}/${current_arch}"
    echo "=========================================="
    
    go build -ldflags="-checklinkname=0 -s -w" -o goecs-desktop .
    
    if [ $? -eq 0 ]; then
        echo "✓ 桌面端编译成功！"
        ls -lh goecs-desktop
    else
        echo "✗ 桌面端编译失败"
        exit 1
    fi
}

# 获取版本信息
get_version() {
    VERSION="v0.0.1-$(date +%Y%m%d)-$(git rev-parse --short HEAD 2>/dev/null || echo 'dev')"
    echo "$VERSION"
}

# macOS 构建
build_macos() {
    VERSION=$(get_version)
    echo "=========================================="
    echo "  构建 macOS 版本"
    echo "  版本: $VERSION"
    echo "=========================================="
    
    echo ""
    echo "构建 macOS ARM64 版本..."
    fyne package -os darwin/arm64 -name goecs
    if [ -f goecs.app ] || [ -d goecs.app ]; then
        tar -czf goecs-macos-arm64-${VERSION}.tar.gz goecs.app
        rm -rf goecs.app
        echo "✓ macOS ARM64 构建成功"
    else
        echo "✗ macOS ARM64 构建失败"
        exit 1
    fi
    
    echo ""
    echo "构建 macOS AMD64 版本..."
    fyne package -os darwin/amd64 -name goecs
    if [ -f goecs.app ] || [ -d goecs.app ]; then
        tar -czf goecs-macos-amd64-${VERSION}.tar.gz goecs.app
        rm -rf goecs.app
        echo "✓ macOS AMD64 构建成功"
    else
        echo "✗ macOS AMD64 构建失败"
        exit 1
    fi
}

# Windows 构建
build_windows() {
    VERSION=$(get_version)
    echo "=========================================="
    echo "  构建 Windows 版本"
    echo "  版本: $VERSION"
    echo "=========================================="
    
    echo ""
    echo "构建 Windows ARM64 版本..."
    fyne package -os windows/arm64 -name goecs
    if [ -f goecs.exe ]; then
        mv goecs.exe goecs-windows-arm64-${VERSION}.exe
        echo "✓ Windows ARM64 构建成功"
    else
        echo "✗ Windows ARM64 构建失败"
        exit 1
    fi
    
    echo ""
    echo "构建 Windows AMD64 版本..."
    fyne package -os windows/amd64 -name goecs
    if [ -f goecs.exe ]; then
        mv goecs.exe goecs-windows-amd64-${VERSION}.exe
        echo "✓ Windows AMD64 构建成功"
    else
        echo "✗ Windows AMD64 构建失败"
        exit 1
    fi
}

# Linux 构建
build_linux() {
    VERSION=$(get_version)
    echo "=========================================="
    echo "  构建 Linux 版本"
    echo "  版本: $VERSION"
    echo "=========================================="
    
    echo ""
    echo "构建 Linux ARM64 版本..."
    fyne package -os linux/arm64 -name goecs
    if [ -f goecs.tar.xz ]; then
        mv goecs.tar.xz goecs-linux-arm64-${VERSION}.tar.xz
        echo "✓ Linux ARM64 构建成功"
    else
        echo "✗ Linux ARM64 构建失败"
        exit 1
    fi
    
    echo ""
    echo "构建 Linux AMD64 版本..."
    fyne package -os linux/amd64 -name goecs
    if [ -f goecs.tar.xz ]; then
        mv goecs.tar.xz goecs-linux-amd64-${VERSION}.tar.xz
        echo "✓ Linux AMD64 构建成功"
    else
        echo "✗ Linux AMD64 构建失败"
        exit 1
    fi
}

# Android 构建
build_android() {
    VERSION=$(get_version)
    echo "=========================================="
    echo "  构建 Android 版本"
    echo "  版本: $VERSION"
    echo "=========================================="
    
    if [ -z "$ANDROID_NDK_HOME" ]; then
        echo "请设置 Android NDK 路径，例如："
        echo "export ANDROID_NDK_HOME=/path/to/android-ndk"
        exit 1
    fi
    
    echo "Android NDK: $ANDROID_NDK_HOME"
    
    echo ""
    echo "构建 Android APK..."
    
    # 构建包含所有架构的 APK
    fyne package -os android -appID com.oneclickvirt.goecs -appVersion "$VERSION"
    
    if [ -f *.apk ]; then
        mv *.apk goecs-android-${VERSION}.apk
        echo "✓ Android APK 构建成功"
    else
        echo "✗ Android APK 构建失败"
        exit 1
    fi
}

# 主流程
case "$BUILD_TYPE" in
    "desktop")
        build_desktop
        ;;
    "android")
        check_fyne_cli
        build_android
        ;;
    "macos")
        check_fyne_cli
        build_macos
        ;;
    "windows")
        check_fyne_cli
        build_windows
        ;;
    "linux")
        check_fyne_cli
        build_linux
        ;;
    "all")
        build_desktop
        echo ""
        check_fyne_cli
        echo ""
        build_macos
        echo ""
        build_windows
        echo ""
        build_linux
        echo ""
        build_android
        ;;
    *)
        echo "用法: $0 [desktop|android|macos|windows|linux|all]"
        echo ""
        echo "  desktop - 构建桌面端应用（默认，用于快速测试）"
        echo "  android - 构建 Android APK (arm64 + x86_64)"
        echo "  macos   - 构建 macOS 应用 (arm64 + amd64)"
        echo "  windows - 构建 Windows 应用 (arm64 + amd64)"
        echo "  linux   - 构建 Linux 应用 (arm64 + amd64)"
        echo "  all     - 构建所有平台"
        exit 1
        ;;
esac

echo ""
echo "=========================================="
echo "  所有构建任务完成"
echo "=========================================="
echo ""
echo "构建产物:"
ls -lh goecs-* 2>/dev/null || echo "无构建产物"
