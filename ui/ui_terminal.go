package ui

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// TerminalOutput 是一个类似终端的输出组件
type TerminalOutput struct {
	widget.Entry
	mu           sync.Mutex
	content      string        // 存储完整内容
	maxBytes     int           // 最大字节数限制
	pendingText  string        // 待刷新的文本
	lastRefresh  time.Time     // 上次刷新时间
	refreshTimer *time.Timer   // 刷新定时器
	updateChan   chan string   // 更新通道
	stopChan     chan struct{} // 停止通道
}

// NewTerminalOutput 创建新的终端输出组件
func NewTerminalOutput() *TerminalOutput {
	terminal := &TerminalOutput{
		content:     "",
		maxBytes:    1024 * 1024 * 10, // 最大10MB
		lastRefresh: time.Now(),
		updateChan:  make(chan string, 1000), // 缓冲通道，避免阻塞
		stopChan:    make(chan struct{}),
	}
	terminal.ExtendBaseWidget(terminal)
	terminal.MultiLine = true
	terminal.Wrapping = fyne.TextWrapOff // 禁用自动换行，支持水平滚动
	terminal.TextStyle = fyne.TextStyle{Monospace: true}
	terminal.Disable() // 禁用编辑

	// 启动批量更新 goroutine
	go terminal.batchUpdateLoop()

	return terminal
}

// batchUpdateLoop 批量更新循环，减少UI刷新频率
func (t *TerminalOutput) batchUpdateLoop() {
	ticker := time.NewTicker(100 * time.Millisecond) // 每100ms最多刷新一次
	defer ticker.Stop()

	for {
		select {
		case <-t.stopChan:
			return
		case text := <-t.updateChan:
			t.mu.Lock()
			t.pendingText += text
			t.mu.Unlock()
		case <-ticker.C:
			t.mu.Lock()
			if t.pendingText != "" {
				// 有待处理的文本，执行更新
				t.content += t.pendingText
				t.pendingText = ""

				// 限制最大字节数
				if len(t.content) > t.maxBytes {
					t.content = t.content[len(t.content)-t.maxBytes:]
					if idx := strings.Index(t.content, "\n"); idx > 0 {
						t.content = t.content[idx+1:]
					}
				}

				currentContent := t.content
				t.mu.Unlock()

				// 使用 fyne.CurrentApp().Driver().DoInMainThread 确保线程安全
				// 但由于 Fyne 的 Entry.Text 更新已经是线程安全的，我们可以直接更新
				t.Entry.Text = currentContent
				t.Refresh()
			} else {
				t.mu.Unlock()
			}
		}
	}
}

// AppendText 追加文本到终端（线程安全）
func (t *TerminalOutput) AppendText(text string) {
	// 移除ANSI颜色代码
	cleanText := t.stripANSI(text)

	// 发送到更新通道，非阻塞
	select {
	case t.updateChan <- cleanText:
		// 成功发送
	default:
		// 通道满了，直接更新（避免丢失数据）
		t.mu.Lock()
		t.pendingText += cleanText
		t.mu.Unlock()
	}
}

// Clear 清空终端内容
func (t *TerminalOutput) Clear() {
	t.mu.Lock()
	t.content = ""
	t.pendingText = ""
	t.mu.Unlock()

	t.Entry.Text = ""
	t.Refresh()
}

// SetFullText 设置完整文本（覆盖现有内容）
func (t *TerminalOutput) SetFullText(text string) {
	t.mu.Lock()

	cleanText := t.stripANSI(text)
	t.content = cleanText
	t.pendingText = ""

	// 限制最大字节数
	if len(t.content) > t.maxBytes {
		t.content = t.content[len(t.content)-t.maxBytes:]
		if idx := strings.Index(t.content, "\n"); idx > 0 {
			t.content = t.content[idx+1:]
		}
	}

	currentContent := t.content
	t.mu.Unlock()

	t.Entry.Text = currentContent
	t.Refresh()
}

// Destroy 销毁终端输出组件，清理资源
func (t *TerminalOutput) Destroy() {
	close(t.stopChan)
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
