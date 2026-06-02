//go:build windows

package ui

import (
	"os"
	"runtime"
	"strings"

	"golang.org/x/sys/windows"
)

// isPrivileged returns true if the process is running as Administrator on Windows.
func isPrivileged() bool {
	// Attempt to open a device that requires Administrator access.
	f, err := os.Open(`\\.\PHYSICALDRIVE0`)
	if err == nil {
		f.Close()
		return true
	}
	return false
}

func requestPrivilegeRestart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	verb, err := windows.UTF16PtrFromString("runas")
	if err != nil {
		return err
	}
	file, err := windows.UTF16PtrFromString(exe)
	if err != nil {
		return err
	}
	args, err := windows.UTF16PtrFromString(shellArgumentString(os.Args[1:]))
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}
	cwdPtr, err := windows.UTF16PtrFromString(cwd)
	if err != nil {
		return err
	}
	return windows.ShellExecute(0, verb, file, args, cwdPtr, windows.SW_SHOWNORMAL)
}

func shellArgumentString(args []string) string {
	if len(args) == 0 {
		return ""
	}
	escaped := make([]string, 0, len(args))
	for _, arg := range args {
		if !strings.ContainsAny(arg, " \t\"") {
			escaped = append(escaped, arg)
			continue
		}
		escaped = append(escaped, `"`+strings.ReplaceAll(arg, `"`, `\"`)+`"`)
	}
	return strings.Join(escaped, " ")
}

// needsPrivilege reports whether the given config requires elevated privileges,
// and returns a human-readable list of affected tests.
func needsPrivilege(config ExecutionConfig) (needs bool, testsZH string, testsEN string) {
	_ = runtime.GOOS // windows
	var zh, en []string

	// Disk test on Windows always uses winsat which requires Administrator.
	if config.SelectedOptions["disk"] {
		zh = append(zh, "磁盘测试 (winsat)")
		en = append(en, "Disk Test (winsat)")
	}
	// Route tracing (nt3 / backtrace) requires raw ICMP sockets → Administrator.
	if config.SelectedOptions["nt3"] {
		zh = append(zh, "三网回程路由检测")
		en = append(en, "3-Net Route Trace")
	}
	if config.SelectedOptions["backtrace"] {
		zh = append(zh, "上游及回程线路检测")
		en = append(en, "Upstream & Backtrace")
	}

	if len(zh) == 0 {
		return false, "", ""
	}
	return true, joinStrings(zh, "、"), joinStrings(en, ", ")
}
