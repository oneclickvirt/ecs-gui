package ui

import "testing"

func TestUpstreamChoiceForPresetMatchesMenuOrder(t *testing.T) {
	wants := map[string]string{
		"full": "1", "full_concurrent": "2", "minimal": "3", "standard": "4",
		"network_focus": "5", "unlock_focus": "6", "network_only": "7",
		"unlock_only": "8", "hardware_only": "9", "ip_quality": "10", "route_only": "11",
		"custom": "",
	}
	for preset, want := range wants {
		if got := upstreamChoiceForPreset(preset); got != want {
			t.Fatalf("preset %q choice = %q, want %q", preset, got, want)
		}
	}
}
