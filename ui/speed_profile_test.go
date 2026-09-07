package ui

import (
	"fmt"
	"strings"
	"testing"
)

type recordedSpeedCall struct {
	kind     string
	operator string
	num      int
	language string
	network  string
}

type recordedSpeedProfileRunner struct {
	calls []recordedSpeedCall
}

func (runner *recordedSpeedProfileRunner) SpeedTestNearbyWithNetwork(network string) {
	runner.calls = append(runner.calls, recordedSpeedCall{kind: "nearby", network: network})
}

func (runner *recordedSpeedProfileRunner) SpeedTestCustomWithNetwork(_ string, operator string, num int, language, network string) {
	runner.calls = append(runner.calls, recordedSpeedCall{kind: "custom", operator: operator, num: num, language: language, network: network})
}

func TestSpeedNetworkForStackPinsDualStackToIPv4(t *testing.T) {
	for stack, want := range map[string]string{
		"DualStack": "tcp4", "IPv4": "tcp4", "IPv6": "tcp6", "None": "",
	} {
		if got := speedNetworkForStack(stack); got != want {
			t.Fatalf("speed network for %q = %q, want %q", stack, got, want)
		}
	}
}

func TestChineseFullSpeedProfilesKeepHistoricalCoverage(t *testing.T) {
	for _, preset := range []string{"full", "full_concurrent"} {
		t.Run(preset, func(t *testing.T) {
			runner := &recordedSpeedProfileRunner{}
			runSpeedProfile(runner, ExecutionConfig{PresetKey: preset, SpNum: 3}, "zh", "tcp4")
			got := formatSpeedCalls(runner.calls)
			want := "nearby/tcp4,global/2/zh/tcp4,cu/3/zh/tcp4,ct/3/zh/tcp4,cmcc/3/zh/tcp4"
			if got != want {
				t.Fatalf("Chinese %s profile = %s, want %s", preset, got, want)
			}
		})
	}
}

func TestChineseFixedSpeedProfilesUseFourDomesticMeasurements(t *testing.T) {
	presets := []string{"minimal", "standard", "network_focus", "unlock_focus", "network_only"}
	for _, preset := range presets {
		t.Run(preset, func(t *testing.T) {
			runner := &recordedSpeedProfileRunner{}
			runSpeedProfile(runner, ExecutionConfig{PresetKey: preset, SpNum: 11}, "zh", "tcp4")
			got := formatSpeedCalls(runner.calls)
			want := "nearby/tcp4,ct/1/zh/tcp4,cu/1/zh/tcp4,cmcc/1/zh/tcp4"
			if got != want {
				t.Fatalf("Chinese %s profile = %s, want %s", preset, got, want)
			}
			if strings.Contains(got, "global") || strings.Contains(got, "other") {
				t.Fatalf("Chinese preset retained international nodes: %s", got)
			}
		})
	}
}

func TestCustomChineseSpeedProfileKeepsCallerNodeCountAndIPv6(t *testing.T) {
	runner := &recordedSpeedProfileRunner{}
	runSpeedProfile(runner, ExecutionConfig{PresetKey: "custom", SpNum: 3}, "zh", "tcp6")
	got := formatSpeedCalls(runner.calls)
	want := "nearby/tcp6,ct/3/zh/tcp6,cu/3/zh/tcp6,cmcc/3/zh/tcp6"
	if got != want {
		t.Fatalf("custom Chinese profile = %s, want %s", got, want)
	}
}

func TestEnglishSpeedProfilesKeepGlobalSelection(t *testing.T) {
	for preset, want := range map[string]string{
		"full": "global/4/en/tcp4", "full_concurrent": "global/4/en/tcp4", "network_only": "global/11/en/tcp4",
	} {
		t.Run(preset, func(t *testing.T) {
			runner := &recordedSpeedProfileRunner{}
			runSpeedProfile(runner, ExecutionConfig{PresetKey: preset}, "en", "tcp4")
			if got := formatSpeedCalls(runner.calls); got != want {
				t.Fatalf("English %s profile = %s, want %s", preset, got, want)
			}
		})
	}
}

func formatSpeedCalls(calls []recordedSpeedCall) string {
	parts := make([]string, 0, len(calls))
	for _, call := range calls {
		if call.kind == "nearby" {
			parts = append(parts, fmt.Sprintf("nearby/%s", call.network))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s/%d/%s/%s", call.operator, call.num, call.language, call.network))
	}
	return strings.Join(parts, ",")
}
