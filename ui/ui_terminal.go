package ui

import (
	"regexp"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// TerminalOutput 是一个类似终端的输出组件
type TerminalOutput struct {
	widget.Entry
	mu       sync.Mutex
	content  string // 存储完整内容
	maxBytes int    // 最大字节数限制
}

// NewTerminalOutput 创建新的终端输出组件
func NewTerminalOutput() *TerminalOutput {
	terminal := &TerminalOutput{
		content:  "",
		maxBytes: 1024 * 1024 * 10, // 最大10MB
	}
	terminal.ExtendBaseWidget(terminal)
	terminal.MultiLine = true
	terminal.Wrapping = fyne.TextWrapOff // 禁用自动换行，支持水平滚动
	terminal.TextStyle = fyne.TextStyle{Monospace: true}
	terminal.Disable() // 禁用编辑
	return terminal
}

// AppendText 追加文本到终端
func (t *TerminalOutput) AppendText(text string) {
	// 移除ANSI颜色代码
	cleanText := t.stripANSI(text)

	t.mu.Lock()
	// 追加到现有内容
	t.content += cleanText

	// 限制最大字节数，保留最新的内容
	if len(t.content) > t.maxBytes {
		// 保留最后的 maxBytes 字节
		t.content = t.content[len(t.content)-t.maxBytes:]
		// 找到第一个换行符，从那里开始（避免截断半行）
		if idx := strings.Index(t.content, "\n"); idx > 0 {
			t.content = t.content[idx+1:]
		}
	}

	// 保存当前内容用于UI更新
	currentContent := t.content
	t.mu.Unlock()

	// 更新 UI - Fyne 会处理线程安全
	// 我们已在 main.go 中设置了 FYNE_DISABLE_DRIVER_THREAD_CHECK=1
	t.Entry.Text = currentContent
	t.Refresh()
}

// Clear 清空终端内容
func (t *TerminalOutput) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.content = ""
	t.Entry.Text = ""
	t.Refresh()
}

// SetFullText 设置完整文本（覆盖现有内容）
func (t *TerminalOutput) SetFullText(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	cleanText := t.stripANSI(text)
	t.content = cleanText

	// 限制最大字节数
	if len(t.content) > t.maxBytes {
		t.content = t.content[len(t.content)-t.maxBytes:]
		if idx := strings.Index(t.content, "\n"); idx > 0 {
			t.content = t.content[idx+1:]
		}
	}

	t.Entry.Text = t.content
	t.Refresh()
}

// stripANSI 移除ANSI转义序列
func (t *TerminalOutput) stripANSI(text string) string {
	ansiRegex := regexp.MustCompile(`\x1B\[[0-9;]*[a-zA-Z]`)
	return ansiRegex.ReplaceAllString(text, "")
}

// GetText 获取当前文本内容
func (t *TerminalOutput) GetText() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.content
}
