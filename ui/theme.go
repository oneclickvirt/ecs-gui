package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type CustomTheme struct{}

var _ fyne.Theme = (*CustomTheme)(nil)

func (m *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// 禁用状态的文字也使用深色显示（而不是默认的淡色）
	if name == theme.ColorNameDisabled {
		return theme.DefaultTheme().Color(theme.ColorNameForeground, theme.VariantLight)
	}
	// 强制使用浅色主题
	return theme.DefaultTheme().Color(name, theme.VariantLight)
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
	// 增大字体以提高可读性
	switch name {
	case theme.SizeNameText:
		return 16 // 默认 14
	case theme.SizeNameHeadingText:
		return 22 // 默认 20
	case theme.SizeNameSubHeadingText:
		return 18 // 默认 16
	case theme.SizeNameCaptionText:
		return 13 // 默认 11
	case theme.SizeNamePadding:
		return 6 // 增加间距
	default:
		return theme.DefaultTheme().Size(name)
	}
}
