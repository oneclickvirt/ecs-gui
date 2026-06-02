# GOECS GUI Version

[![Build All UI APP](https://github.com/oneclickvirt/ecs-gui/actions/workflows/build.yml/badge.svg)](https://github.com/oneclickvirt/ecs-gui/actions/workflows/build.yml)

[中文](README.md)

GOECS GUI is a Fyne-based cross-platform system benchmark and network testing app for Android, macOS, Windows, and Linux.

Upstream project: https://github.com/oneclickvirt/ecs

## Features

- Select basic info, CPU, memory, disk, streaming unlock, route, ping, and speed tests from the GUI
- Show real-time stage progress and the current running item
- Copy results, export a text file, or upload a share link
- Switch between Chinese/English UI and light/dark themes
- Send completion notifications on Android and request Windows UAC elevation when privileged tests need it

## Supported Platforms

### Android

- GitHub Actions build includes APK signing
- Minimum version: Android 7.0 (API Level 24)
- Recommended version: Android 13 (API Level 33) or newer
- Architectures: ARM64, x86_64

### macOS

- GitHub Actions build is unsigned by default
- Minimum version: macOS 11.0
- Architectures: Apple Silicon (ARM64), Intel (AMD64)
- Signing guide: [docs/macos-signing.en.md](docs/macos-signing.en.md)

### Windows

- GitHub Actions build is unsigned by default
- Minimum version: Windows 10
- Architectures: ARM64, AMD64
- Some disk and route tests require Administrator privileges; the GUI requests a UAC restart when needed

### Linux

- GitHub Actions build is unsigned by default
- Architectures: ARM64, AMD64

## Local Build

### Requirements

1. Go 1.25.4 or newer
2. Android SDK, only for Android builds
3. Android NDK 25.2.9519653, only for Android builds
4. JDK 17 or newer, only for Android builds

### Environment

```bash
export ANDROID_NDK_HOME=/path/to/android-ndk
go install fyne.io/tools/cmd/fyne@latest
```

### Commands

```bash
./build.sh desktop
./build.sh android
./build.sh macos
./build.sh windows
./build.sh linux
./build.sh all
```

Artifacts are written to the repository root.

Version sources:

- `VERSION` stores the app semantic version, for example `0.1.139`
- `FyneApp.toml` `Version` should stay in sync with `VERSION`
- GitHub Actions release tags use `vYYYYMMDD-HHMMSS`

## Development

```bash
git clone https://github.com/oneclickvirt/ecs-gui.git
cd ecs-gui
go mod download
go run -ldflags="-checklinkname=0" .
```
