package main

import (
	"net/http"
	"os"
	"regexp"
	"strings"
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
	if got := appmeta.ReleaseVersion(); got != pinnedGUIVersion(t) {
		t.Fatalf("GUI release version = %q, want %q", got, pinnedGUIVersion(t))
	}
	if got := ecsapi.DefaultVersion; got != appmeta.UpstreamECSVersion {
		t.Fatalf("ECS version = %q, GUI metadata = %q", got, appmeta.UpstreamECSVersion)
	}
	if got := speedtestmodel.SpeedTestVersion; got != pinnedModuleVersion(t, "github.com/oneclickvirt/speedtest") {
		t.Fatalf("speedtest component version = %q, want source-pinned version", got)
	}
	if got := privatepst.PrivateSpeedTestVersion; got != pinnedModuleVersion(t, "github.com/oneclickvirt/privatespeedtest") {
		t.Fatalf("private speedtest component version = %q, want source-pinned version", got)
	}
	if got := showwinspeedtest.Version(); got != strings.TrimPrefix(pinnedModuleVersion(t, "github.com/showwin/speedtest-go"), "v") {
		t.Fatalf("speedtest-go version = %q, want source-pinned version", got)
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

func pinnedModuleVersion(t *testing.T, module string) string {
	t.Helper()
	contents, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	match := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(module) + `\s+(\S+)`).FindSubmatch(contents)
	if len(match) != 2 {
		t.Fatalf("module %s is not pinned", module)
	}
	return string(match[1])
}
func pinnedGUIVersion(t *testing.T) string {
	t.Helper()
	contents, err := os.ReadFile("VERSION")
	if err != nil {
		t.Fatal(err)
	}
	return "v" + strings.TrimSpace(string(contents))
}
