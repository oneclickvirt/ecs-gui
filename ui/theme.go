package ui

import (
	"image/color"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const (
	themeModeLight = "light"
	themeModeDark  = "dark"

	themePreferenceKey = "theme_mode"
)

type CustomTheme struct {
	Variant      fyne.ThemeVariant
	forceVariant bool
}

var _ fyne.Theme = (*CustomTheme)(nil)

func NewCustomTheme(mode string) *CustomTheme {
	variant := theme.VariantLight
	if mode == themeModeDark {
		variant = theme.VariantDark
	}
	return &CustomTheme{Variant: variant, forceVariant: true}
}

func normalizeThemeMode(mode string) string {
	if mode == themeModeDark {
		return themeModeDark
	}
	return themeModeLight
}

func (ui *TestUI) applyThemeMode(mode string) {
	ui.themeMode = normalizeThemeMode(mode)
	if ui.App != nil {
		ui.App.Preferences().SetString(themePreferenceKey, ui.themeMode)
		ui.App.Settings().SetTheme(NewCustomTheme(ui.themeMode))
	}
}

func (ui *TestUI) themeLabelByMode(mode string) string {
	if normalizeThemeMode(mode) == themeModeDark {
		return ui.tr("theme.dark")
	}
	return ui.tr("theme.light")
}

func (ui *TestUI) themeModeByLabel(label string) string {
	if label == ui.tr("theme.dark") || label == "Dark" || label == "深色" {
		return themeModeDark
	}
	return themeModeLight
}

func (m *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if m != nil && m.forceVariant {
		variant = m.Variant
	}
	// 禁用状态的文字也使用深色显示（而不是默认的淡色）
	if name == theme.ColorNameDisabled {
		return theme.DefaultTheme().Color(theme.ColorNameForeground, variant)
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (m *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	// 使用 Fyne 内置字体资源，支持中文
	// Fyne 2.4+ 内置了 Noto Sans 字体，包含中文支持
	if style.Monospace {
		return theme.DefaultTheme().Font(fyne.TextStyle{Monospace: true})
	}
	if style.Bold {
		if style.Italic {
			return theme.DefaultTheme().Font(fyne.TextStyle{Bold: true, Italic: true})
		}
		return theme.DefaultTheme().Font(fyne.TextStyle{Bold: true})
	}
	if style.Italic {
		return theme.DefaultTheme().Font(fyne.TextStyle{Italic: true})
	}
	// 返回默认字体
	return theme.DefaultTheme().Font(fyne.TextStyle{})
}

func (m *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	compact := runtime.GOOS == "android" || runtime.GOOS == "ios"

	// 统一字号和间距节奏：移动端更紧凑，桌面端更舒展
	switch name {
	case theme.SizeNameText:
		if compact {
			return 14
		}
		return 15
	case theme.SizeNameHeadingText:
		if compact {
			return 19
		}
		return 22
	case theme.SizeNameSubHeadingText:
		if compact {
			return 16
		}
		return 18
	case theme.SizeNameCaptionText:
		if compact {
			return 11
		}
		return 12
	case theme.SizeNamePadding:
		if compact {
			return 4
		}
		return 8
	case theme.SizeNameInlineIcon:
		if compact {
			return 18
		}
		return 20
	default:
		return theme.DefaultTheme().Size(name)
	}
}
