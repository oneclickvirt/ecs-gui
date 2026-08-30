package ui

import "testing"

func TestSpecialUnlockRegionRoundTrips(t *testing.T) {
	for _, language := range []string{langZH, langEN} {
		for _, code := range []string{"21", "22"} {
			label := unlockRegionCodeToLabel(code, language)
			if got := unlockRegionLabelToCode(label, language); got != code {
				t.Fatalf("language %q: special region %s round trip = %q", language, code, got)
			}
		}
	}
}
