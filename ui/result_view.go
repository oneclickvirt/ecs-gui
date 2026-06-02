package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// createResultTab 创建测试结果页面
func (ui *TestUI) createResultTab() fyne.CanvasObject {
	statusRow := container.NewHBox(
		ui.StatusLabel,
		layout.NewSpacer(),
		ui.StatusBadge,
	)
	statusBar := container.NewVBox(statusRow, ui.CurrentItem, ui.ProgressBar)

	copyButton := widget.NewButtonWithIcon(ui.tr("button.copy"), theme.ContentCopyIcon(), ui.copyResults)
	exportButton := widget.NewButtonWithIcon(ui.tr("button.export"), theme.DownloadIcon(), ui.exportResults)
	shareButton := widget.NewButtonWithIcon(ui.tr("button.share"), theme.MailForwardIcon(), ui.shareResults)
	clearButton := widget.NewButtonWithIcon(ui.tr("button.clear"), theme.DeleteIcon(), ui.clearResults)

	actions := []fyne.CanvasObject{clearButton, copyButton, exportButton, shareButton}
	actionsBar := container.NewHBox(actions...)
	if isMobilePlatform() {
		actionsBar = container.NewAdaptiveGrid(2, actions...)
	} else {
		actionsBar = container.NewHBox(layout.NewSpacer(), clearButton, copyButton, exportButton, shareButton)
	}

	header := widget.NewCard("", "", container.NewVBox(
		statusBar,
		actionsBar,
	))

	terminalScroll := container.NewScroll(container.NewPadded(ui.Terminal))

	return container.NewBorder(
		header,
		nil,
		nil,
		nil,
		terminalScroll,
	)
}
