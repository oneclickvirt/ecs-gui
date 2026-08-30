package ui

import (
	"strings"
	"testing"
)

func TestNormalizeGUITraceBoundariesSeparatesStopReasonAndNextHeader(t *testing.T) {
	input := "Trace Stopped: Destination Reached at Hop 19 (ICMP Echo Reply)\x1b[33m\x1b[01m广州移动 - ICMP v4 -\x1b[0mtraceroute to 120.196.165.24, 30 hops max"
	got := sanitizeGUIText(input)
	if !strings.Contains(got, "ICMP Echo Reply)\n\x1b[33m\x1b[01m广州移动 - ICMP v4 -") {
		t.Fatalf("trace boundary was not repaired: %q", got)
	}
	if strings.Contains(got, "ICMP Echo Reply)广州移动") {
		t.Fatalf("next carrier header was concatenated: %q", got)
	}
}

func TestNormalizeGUITraceBoundariesLeavesOrdinaryICMPTextUntouched(t *testing.T) {
	input := "1.00 ms AS4134 ICMP Echo Reply"
	if got := sanitizeGUIText(input); got != input {
		t.Fatalf("ordinary ICMP output changed: %q", got)
	}
}

func TestNormalizeGUITraceBoundariesHandlesHeaderWithoutANSI(t *testing.T) {
	input := "Trace Stopped: Maximum Hops Reached at Hop 30 (No Destination Response)广州电信 - ICMP v6 - traceroute to 2001:db8::1"
	want := "Trace Stopped: Maximum Hops Reached at Hop 30 (No Destination Response)\n广州电信 - ICMP v6 - traceroute to 2001:db8::1"
	if got := sanitizeGUIText(input); got != want {
		t.Fatalf("plain trace boundary = %q, want %q", got, want)
	}
}

func TestNormalizeGUITraceBoundariesRepairsMultipleBoundariesOnOneLine(t *testing.T) {
	input := "Trace Stopped: Destination Reached at Hop 1 (ICMP Echo Reply)广州移动 - ICMP v4 - traceroute to 120.196.165.24 Trace Stopped: Destination Reached at Hop 2 (ICMP Echo Reply)广州联通 - ICMP v4 - traceroute to 210.21.196.6"
	want := "Trace Stopped: Destination Reached at Hop 1 (ICMP Echo Reply)\n广州移动 - ICMP v4 - traceroute to 120.196.165.24 Trace Stopped: Destination Reached at Hop 2 (ICMP Echo Reply)\n广州联通 - ICMP v4 - traceroute to 210.21.196.6"
	if got := sanitizeGUIText(input); got != want {
		t.Fatalf("multiple trace boundaries = %q, want %q", got, want)
	}
}

func TestNormalizeGUITraceBoundariesHandlesStopReasonWithoutResponseParentheses(t *testing.T) {
	input := "Trace Stopped: custom ICMP failure at Hop 5广州移动 - ICMP v4 - traceroute to 120.196.165.24"
	want := "Trace Stopped: custom ICMP failure at Hop 5\n广州移动 - ICMP v4 - traceroute to 120.196.165.24"
	if got := sanitizeGUIText(input); got != want {
		t.Fatalf("no-parentheses trace boundary = %q, want %q", got, want)
	}
}

func TestNormalizeGUITraceBoundariesPreservesAlreadySeparatedTraceLines(t *testing.T) {
	input := "Trace Stopped: Destination Reached at Hop 19 (ICMP Echo Reply)\n广州移动 - ICMP v4 - traceroute to 120.196.165.24"
	if got := sanitizeGUIText(input); got != input {
		t.Fatalf("already separated trace lines changed: %q", got)
	}
}

func TestTerminalOutputRepairsBoundarySplitAcrossPendingChunks(t *testing.T) {
	terminal := &TerminalOutput{
		maxBytes:   1024 * 1024,
		maxLines:   5000,
		maxPending: 1024,
		updateChan: make(chan string, 2),
		stopChan:   make(chan struct{}),
	}
	terminal.AppendText("Trace Stopped: No Continuing Route Observed at Hop 4 (ICMP Host Unreachable (!H))")
	terminal.AppendText("广州移动 - ICMP v4 - traceroute to 120.196.165.24")

	terminal.mu.Lock()
	for len(terminal.updateChan) > 0 {
		terminal.appendPendingLocked(<-terminal.updateChan)
	}
	terminal.mu.Unlock()

	want := "Trace Stopped: No Continuing Route Observed at Hop 4 (ICMP Host Unreachable (!H))\n广州移动 - ICMP v4 - traceroute to 120.196.165.24"
	if got := terminal.GetText(); got != want {
		t.Fatalf("pending trace boundary = %q, want %q", got, want)
	}
}
