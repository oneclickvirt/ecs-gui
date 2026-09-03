package main

import (
	"testing"

	"github.com/oneclickvirt/basics/network/resolver"
	"github.com/oneclickvirt/ecs-gui/internal/appmeta"
	ecsapi "github.com/oneclickvirt/ecs/api"
	privatepst "github.com/oneclickvirt/privatespeedtest/pst"
	speedtestmodel "github.com/oneclickvirt/speedtest/model"
	showwinspeedtest "github.com/showwin/speedtest-go/speedtest"
)

func TestReleaseDependencyContract(t *testing.T) {
	if got := appmeta.ReleaseVersion(); got != "v0.1.197" {
		t.Fatalf("GUI release version = %q, want v0.1.197", got)
	}
	if got := ecsapi.DefaultVersion; got != appmeta.UpstreamECSVersion {
		t.Fatalf("ECS version = %q, GUI metadata = %q", got, appmeta.UpstreamECSVersion)
	}
	if got := speedtestmodel.SpeedTestVersion; got != "v0.0.25" {
		t.Fatalf("speedtest component version = %q, want v0.0.25", got)
	}
	if got := privatepst.PrivateSpeedTestVersion; got != "v0.0.16" {
		t.Fatalf("private speedtest component version = %q, want v0.0.16", got)
	}
	if got := showwinspeedtest.Version(); got != "1.8.2" {
		t.Fatalf("speedtest-go version = %q, want 1.8.2", got)
	}
	for _, endpoint := range resolver.DefaultEndpoints() {
		if endpoint.Name == "360 Public DNS" && endpoint.URL == "tls://dot.360.cn:853" {
			return
		}
	}
	t.Fatal("GUI dependency graph is missing the validated 360 DoT endpoint")
}

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
