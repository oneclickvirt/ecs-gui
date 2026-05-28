//go:build windows

package ui

import (
	"os"
	"runtime"
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
