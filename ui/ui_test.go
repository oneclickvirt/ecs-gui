package ui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2/test"
)

func newTestUIForTest(t *testing.T) *TestUI {
	t.Helper()
	app := test.NewApp()
	ui := NewTestUI(app)
	t.Cleanup(func() {
		if ui.Terminal != nil {
			ui.Terminal.Destroy()
		}
		app.Quit()
	})
	return ui
}

func TestCollectExecutionConfigClampsNumericValues(t *testing.T) {
	ui := newTestUIForTest(t)

	ui.SpNumEntry.SetText("999")
	ui.OutputWidthEntry.SetText("999")
	ui.OutputFileEntry.SetText("")

	config := ui.collectExecutionConfig()
	if config.SpNum != 20 {
		t.Fatalf("expected sp num to clamp to 20, got %d", config.SpNum)
	}
	if config.OutputWidth != 120 {
		t.Fatalf("expected output width to clamp to 120, got %d", config.OutputWidth)
	}
	if config.FilePath != "goecs.md" {
		t.Fatalf("expected default file path, got %q", config.FilePath)
	}
}

func TestStandardPresetSelectsExpectedOptions(t *testing.T) {
	ui := newTestUIForTest(t)

	ui.onPresetChanged(ui.presetLabelByKey("standard"))

	if !ui.BasicCheck.Checked || !ui.CpuCheck.Checked || !ui.MemoryCheck.Checked || !ui.DiskCheck.Checked {
		t.Fatal("standard preset should enable core performance checks")
	}
	if !ui.UnlockCheck.Checked || !ui.Nt3Check.Checked || !ui.SpeedCheck.Checked {
		t.Fatal("standard preset should enable unlock, route, and speed checks")
	}
	if ui.PingCheck.Checked || ui.PingTgdcCheck.Checked || ui.PingWebCheck.Checked {
		t.Fatal("standard preset should not enable ping extensions")
	}
	if ui.SpNumEntry.Text != "5" {
		t.Fatalf("standard preset should set speed node count to 5, got %q", ui.SpNumEntry.Text)
	}
}

func TestBuildProgressStepsHonorsChinaMode(t *testing.T) {
	steps := buildProgressSteps(ExecutionConfig{
		SelectedOptions: map[string]bool{
			"unlock": true,
			"ping":   false,
		},
		ChinaModeEnabled: true,
	}, true)

	joined := strings.Join(steps, ",")
	if strings.Contains(joined, "progress.unlock") {
		t.Fatal("china mode should remove streaming unlock from progress steps")
	}
	if !strings.Contains(joined, "progress.ping") {
		t.Fatal("china mode should force ping progress step")
	}
}

func TestFormatResultExportWrapsPlainTextOnce(t *testing.T) {
	exported := formatResultExport("hello\n")
	if !strings.HasPrefix(exported, "# GOECS Result\n\n```text\nhello\n```") {
		t.Fatalf("unexpected export format: %q", exported)
	}

	again := formatResultExport(exported)
	if strings.Count(again, "# GOECS Result") != 1 {
		t.Fatalf("export header duplicated: %q", again)
	}
}

func TestBuildResultSummaryDetectsSections(t *testing.T) {
	output := "CPU-Test\nSpeed-Test\n10.00 Mbps\nError: timeout\n"
	summary := BuildResultSummary("en", output)

	if !strings.Contains(summary, "CPU") || !strings.Contains(summary, "Speed") {
		t.Fatalf("summary did not detect sections: %q", summary)
	}
	if !strings.Contains(summary, "Speed Rows         : 1") {
		t.Fatalf("summary did not count speed rows: %q", summary)
	}
}
