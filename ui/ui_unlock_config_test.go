package ui

import "testing"

func TestAIOnlyUnlockRegionRoundTrip(t *testing.T) {
	for _, language := range []string{langZH, langEN} {
		label := unlockRegionCodeToLabel("21", language)
		if got := unlockRegionLabelToCode(label, language); got != "21" {
			t.Fatalf("language %q: AI-only region round trip = %q", language, got)
		}
	}
}
