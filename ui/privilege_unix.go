//go:build !windows

package ui

import (
	"os"
	"runtime"
)

// isPrivileged returns true when running as root (uid 0) on Unix-like systems.
// On mobile platforms (Android/iOS) this check is skipped (returns true).
func isPrivileged() bool {
	switch runtime.GOOS {
	case "android", "ios":
		return true
	default:
		return os.Getuid() == 0
	}
}

// needsPrivilege reports whether the given config requires elevated privileges,
// and returns a human-readable list of affected tests.
func needsPrivilege(config ExecutionConfig) (needs bool, testsZH string, testsEN string) {
	switch runtime.GOOS {
	case "android", "ios":
		return false, "", ""
	}

	var zh, en []string

	// Route tracing requires raw ICMP sockets → root or CAP_NET_RAW.
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
