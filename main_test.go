package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/oneclickvirt/basics/network/resolver"
	"github.com/oneclickvirt/ecs-gui/internal/appmeta"
	ecsapi "github.com/oneclickvirt/ecs/api"
	privatepst "github.com/oneclickvirt/privatespeedtest/pst"
	speedtestmodel "github.com/oneclickvirt/speedtest/model"
	showwinspeedtest "github.com/showwin/speedtest-go/speedtest"
)

func TestReleaseDependencyContract(t *testing.T) {
	if got := appmeta.ReleaseVersion(); got != "v0.2.12" {
		t.Fatalf("GUI release version = %q, want v0.2.12", got)
	}
	if got := ecsapi.DefaultVersion; got != appmeta.UpstreamECSVersion {
		t.Fatalf("ECS version = %q, GUI metadata = %q", got, appmeta.UpstreamECSVersion)
	}
	if got := speedtestmodel.SpeedTestVersion; got != "v0.0.37" {
		t.Fatalf("speedtest component version = %q, want v0.0.37", got)
	}
	if got := privatepst.PrivateSpeedTestVersion; got != "v0.0.24" {
		t.Fatalf("private speedtest component version = %q, want v0.0.24", got)
	}
	if got := showwinspeedtest.Version(); got != "1.8.3" {
		t.Fatalf("speedtest-go version = %q, want 1.8.3", got)
	}
	client := speedtestmodel.NewThroughputHTTPClient(speedtestmodel.NetworkIPv4, time.Second)
	transport, ok := client.Transport.(*http.Transport)
	if !ok || !transport.DisableKeepAlives {
		t.Fatalf("speedtest throughput transport = %#v, want isolated HTTP connections", client.Transport)
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
