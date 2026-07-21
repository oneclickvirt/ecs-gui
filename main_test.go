package main

import "testing"

func TestParseGUIFlagsUsesPrivateFlagSet(t *testing.T) {
	showVersion, showHelp, err := parseGUIFlags([]string{"-v"})
	if err != nil || !showVersion || showHelp {
		t.Fatalf("version flags: version=%t help=%t err=%v", showVersion, showHelp, err)
	}

	showVersion, showHelp, err = parseGUIFlags([]string{"-help"})
	if err != nil || showVersion || !showHelp {
		t.Fatalf("help flags: version=%t help=%t err=%v", showVersion, showHelp, err)
	}
}

func TestParseGUIFlagsRejectsUnknownOption(t *testing.T) {
	if _, _, err := parseGUIFlags([]string{"-unknown"}); err == nil {
		t.Fatal("unknown option was accepted")
	}
}
