package ui

import (
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var ansiRegex = regexp.MustCompile(`\x1B\[[0-9;]*[a-zA-Z]`)

// TerminalOutput 是一个类似终端的输出组件
type TerminalOutput struct {
	widget.Entry
	mu          sync.Mutex
	closeOnce   sync.Once
	content     string        // 存储完整内容
	maxBytes    int           // 最大字节数限制
	maxLines    int           // 最大显示行数
	maxPending  int           // 待刷新文本最大字节数
	pendingText string        // 待刷新的文本
	updateChan  chan string   // 更新通道
	stopChan    chan struct{} // 停止通道
}

// NewTerminalOutput 创建新的终端输出组件
func NewTerminalOutput() *TerminalOutput {
	maxBytes := 2 * 1024 * 1024
	maxPending := 128 * 1024
	maxLines := 5000
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		maxBytes = 768 * 1024
		maxPending = 64 * 1024
		maxLines = 1600
	}
	terminal := &TerminalOutput{
		content:    "",
		maxBytes:   maxBytes,
		maxLines:   maxLines,
		maxPending: maxPending,
		updateChan: make(chan string, 96),
		stopChan:   make(chan struct{}),
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
	interval := 180 * time.Millisecond
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		interval = 320 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-t.stopChan:
			return
		case text := <-t.updateChan:
			t.mu.Lock()
			t.appendPendingLocked(text)
			t.mu.Unlock()
		case <-ticker.C:
			t.mu.Lock()
			if t.pendingText != "" {
				t.content += t.pendingText
				t.pendingText = ""
				t.trimToMaxContentLocked()

				currentContent := t.content
				t.mu.Unlock()

				fyne.Do(func() {
					t.Entry.SetText(currentContent)
					t.Refresh()
				})
			} else {
				t.mu.Unlock()
			}
		}
	}
}

// AppendText 追加文本到终端（线程安全）
func (t *TerminalOutput) AppendText(text string) {
	cleanText := t.stripANSI(text)

	// 发送到更新通道，非阻塞
	select {
	case t.updateChan <- cleanText:
		// 成功发送
	default:
		t.mu.Lock()
		t.appendPendingLocked(cleanText)
		t.mu.Unlock()
	}
}

// Clear 清空终端内容
func (t *TerminalOutput) Clear() {
	t.mu.Lock()
	t.content = ""
	t.pendingText = ""
	t.mu.Unlock()

	fyne.Do(func() {
		t.Entry.SetText("")
		t.Refresh()
	})
}

// SetFullText 设置完整文本（覆盖现有内容）
func (t *TerminalOutput) SetFullText(text string) {
	t.mu.Lock()

	cleanText := t.stripANSI(text)
	t.content = cleanText
	t.pendingText = ""
	t.trimToMaxContentLocked()

	currentContent := t.content
	t.mu.Unlock()

	fyne.Do(func() {
		t.Entry.SetText(currentContent)
		t.Refresh()
	})
}

// Destroy 销毁终端输出组件，清理资源
func (t *TerminalOutput) Destroy() {
	t.closeOnce.Do(func() {
		close(t.stopChan)
	})
}

// stripANSI 移除ANSI转义序列
func (t *TerminalOutput) stripANSI(text string) string {
	return ansiRegex.ReplaceAllString(text, "")
}

func (t *TerminalOutput) appendPendingLocked(text string) {
	t.pendingText += text
	if len(t.pendingText) <= t.maxPending {
		return
	}

	keep := t.pendingText[len(t.pendingText)-t.maxPending:]
	if idx := strings.Index(keep, "\n"); idx > 0 {
		keep = keep[idx+1:]
	}
	t.pendingText = "\n[输出过快，已丢弃部分历史日志]\n" + keep
}

func (t *TerminalOutput) trimToMaxContentLocked() {
	if len(t.content) <= t.maxBytes {
		t.trimLinesLocked()
		return
	}
	t.content = t.content[len(t.content)-t.maxBytes:]
	if idx := strings.Index(t.content, "\n"); idx > 0 {
		t.content = t.content[idx+1:]
	}
	t.trimLinesLocked()
}

func (t *TerminalOutput) trimLinesLocked() {
	if t.maxLines <= 0 {
		return
	}
	lineCount := strings.Count(t.content, "\n")
	if lineCount <= t.maxLines {
		return
	}
	cut := 0
	toDrop := lineCount - t.maxLines
	for i := 0; i < toDrop; i++ {
		next := strings.Index(t.content[cut:], "\n")
		if next < 0 {
			break
		}
		cut += next + 1
	}
	if cut > 0 && cut < len(t.content) {
		t.content = "[历史输出过长，已保留最近内容]\n" + t.content[cut:]
	}
}

// GetText 获取当前文本内容
func (t *TerminalOutput) GetText() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.pendingText != "" {
		t.content += t.pendingText
		t.pendingText = ""
		t.trimToMaxContentLocked()
	}
	return t.content
}
