#!/bin/bash

set -e

BUILD_TYPE=${1:-"desktop"}
FYNE_BUILD_FLAGS_DEFAULT="-trimpath -buildvcs=false -ldflags '-checklinkname=0 -s -w -buildid='"
APP_ID="com.oneclickvirt.goecs"
APP_NAME="goecs"
FYNE_CMD=${FYNE_CMD:-fyne}

# 检查 Fyne CLI 是否安装
check_fyne_cli() {
    if command -v "$FYNE_CMD" &> /dev/null && "$FYNE_CMD" package --help 2>&1 | grep -q -- "--app-id"; then
        echo "Fyne CLI 已安装"
        return
    fi

    local go_bin
    go_bin="$(go env GOPATH)/bin"
    echo "正在安装 Fyne CLI..."
    go install fyne.io/tools/cmd/fyne@latest
    if [ $? -ne 0 ]; then
        echo "Fyne CLI 安装失败"
        exit 1
    fi

    FYNE_CMD="${go_bin}/fyne"
    if ! command -v "$FYNE_CMD" &> /dev/null || ! "$FYNE_CMD" package --help 2>&1 | grep -q -- "--app-id"; then
        echo "Fyne CLI 安装后仍不可用或不支持 --app-id"
        exit 1
    fi
    export FYNE_CMD
    echo "Fyne CLI 安装成功"
}

ensure_fyne_cli() {
    if ! command -v "$FYNE_CMD" &> /dev/null || ! "$FYNE_CMD" package --help 2>&1 | grep -q -- "--app-id"; then
        go install fyne.io/tools/cmd/fyne@latest
        FYNE_CMD="$(go env GOPATH)/bin/fyne"
        if ! command -v "$FYNE_CMD" &> /dev/null || ! "$FYNE_CMD" package --help 2>&1 | grep -q -- "--app-id"; then
            echo "Fyne CLI 不可用或不支持 --app-id"
            exit 1
        fi
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
    
    go build -trimpath -buildvcs=false -ldflags="-checklinkname=0 -s -w -buildid=" -o goecs-desktop .
    
    if [ $? -eq 0 ]; then
        echo "✓ 桌面端编译成功！"
        ls -lh goecs-desktop
    else
        echo "✗ 桌面端编译失败"
        exit 1
    fi
}

# 获取版本信息
get_app_version() {
    local version
    version=$(tr -d '[:space:]' < VERSION 2>/dev/null || true)
    if [ -z "$version" ]; then
        version=$(awk '$1 == "github.com/oneclickvirt/ecs" { print $2; exit }' go.mod | sed 's/^v//')
    fi
    if [ -z "$version" ]; then
        version="0.0.0"
    fi
    echo "$version"
}

get_version() {
    VERSION="v$(get_app_version)-$(date +%Y%m%d)-$(git rev-parse --short HEAD 2>/dev/null || echo 'dev')"
    echo "$VERSION"
}

# macOS 构建
build_macos() {
    ensure_fyne_cli
    VERSION=$(get_version)
    APP_VERSION=$(get_app_version)
    echo "=========================================="
    echo "  构建 macOS 版本"
    echo "  版本: $VERSION"
    echo "  应用版本: $APP_VERSION"
    echo "=========================================="
    
    echo ""
    echo "构建 macOS ARM64 版本..."
    GOOS=darwin GOARCH=arm64 FYNE_BUILD_FLAGS="$FYNE_BUILD_FLAGS_DEFAULT" "$FYNE_CMD" package -os darwin -name "$APP_NAME" --app-id "$APP_ID" --app-version "$APP_VERSION"
    if [ -f goecs.app ] || [ -d goecs.app ]; then
        GZIP=-9 tar -czf goecs-macos-arm64-${VERSION}.tar.gz goecs.app
        rm -rf goecs.app
        echo "✓ macOS ARM64 构建成功"
    else
        echo "✗ macOS ARM64 构建失败"
        exit 1
    fi
    
    echo ""
    echo "构建 macOS AMD64 版本..."
    GOOS=darwin GOARCH=amd64 FYNE_BUILD_FLAGS="$FYNE_BUILD_FLAGS_DEFAULT" "$FYNE_CMD" package -os darwin -name "$APP_NAME" --app-id "$APP_ID" --app-version "$APP_VERSION"
    if [ -f goecs.app ] || [ -d goecs.app ]; then
        GZIP=-9 tar -czf goecs-macos-amd64-${VERSION}.tar.gz goecs.app
        rm -rf goecs.app
        echo "✓ macOS AMD64 构建成功"
    else
        echo "✗ macOS AMD64 构建失败"
        exit 1
    fi
}

# Windows 构建
build_windows() {
    ensure_fyne_cli
    VERSION=$(get_version)
    APP_VERSION=$(get_app_version)
    echo "=========================================="
    echo "  构建 Windows 版本"
    echo "  版本: $VERSION"
    echo "  应用版本: $APP_VERSION"
    echo "=========================================="
    
    echo ""
    echo "构建 Windows ARM64 版本..."
    GOOS=windows GOARCH=arm64 FYNE_BUILD_FLAGS="$FYNE_BUILD_FLAGS_DEFAULT" "$FYNE_CMD" package -os windows -name "$APP_NAME" --app-id "$APP_ID" --app-version "$APP_VERSION"
    if [ -f goecs.exe ]; then
        mv goecs.exe goecs-windows-arm64-${VERSION}.exe
        echo "✓ Windows ARM64 构建成功"
    else
        echo "✗ Windows ARM64 构建失败"
        exit 1
    fi
    
    echo ""
    echo "构建 Windows AMD64 版本..."
    GOOS=windows GOARCH=amd64 FYNE_BUILD_FLAGS="$FYNE_BUILD_FLAGS_DEFAULT" "$FYNE_CMD" package -os windows -name "$APP_NAME" --app-id "$APP_ID" --app-version "$APP_VERSION"
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
    ensure_fyne_cli
    VERSION=$(get_version)
    APP_VERSION=$(get_app_version)
    echo "=========================================="
    echo "  构建 Linux 版本"
    echo "  版本: $VERSION"
    echo "  应用版本: $APP_VERSION"
    echo "=========================================="
    
    echo ""
    echo "构建 Linux ARM64 版本..."
    GOOS=linux GOARCH=arm64 FYNE_BUILD_FLAGS="$FYNE_BUILD_FLAGS_DEFAULT" "$FYNE_CMD" package -os linux -name "$APP_NAME" --app-id "$APP_ID" --app-version "$APP_VERSION"
    if [ -f goecs.tar.xz ]; then
        mv goecs.tar.xz goecs-linux-arm64-${VERSION}.tar.xz
        echo "✓ Linux ARM64 构建成功"
    else
        echo "✗ Linux ARM64 构建失败"
        exit 1
    fi
    
    echo ""
    echo "构建 Linux AMD64 版本..."
    GOOS=linux GOARCH=amd64 FYNE_BUILD_FLAGS="$FYNE_BUILD_FLAGS_DEFAULT" "$FYNE_CMD" package -os linux -name "$APP_NAME" --app-id "$APP_ID" --app-version "$APP_VERSION"
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
    ensure_fyne_cli
    VERSION=$(get_version)
    APP_VERSION=$(get_app_version)
    echo "=========================================="
    echo "  构建 Android 版本"
    echo "  版本: $VERSION"
    echo "  应用版本: $APP_VERSION"
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
    FYNE_BUILD_FLAGS="$FYNE_BUILD_FLAGS_DEFAULT" "$FYNE_CMD" package -os android --app-id "$APP_ID" --app-version "$APP_VERSION"
    
    if [ -f *.apk ]; then
        mv *.apk ecs-gui-android-${VERSION}.apk
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
ARTIFACTS=$(find . -maxdepth 1 -type f \( -name 'goecs-*' -o -name 'ecs-gui-*' \) -print | sort)
if [ -n "$ARTIFACTS" ]; then
    # shellcheck disable=SC2086
    ls -lh $ARTIFACTS
else
    echo "无构建产物"
fi
