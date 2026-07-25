package ui

import "testing"

func TestDiskResultTextNeverLeavesAnEmptyEnglishSection(t *testing.T) {
	if got := diskResultText(" EN ", "\n\t"); got != " Disk benchmark returned no usable data.\n" {
		t.Fatalf("English empty disk result = %q", got)
	}
}

func TestDiskResultTextKeepsExistingOutputUnchanged(t *testing.T) {
	const result = " Test Path  Block  Read(IOPS)\n /          4k     1000\n"
	if got := diskResultText("en", result); got != result {
		t.Fatalf("non-empty disk result changed:\n%s", got)
	}
}

func TestDiskResultTextUsesChineseFallback(t *testing.T) {
	if got := diskResultText("zh", ""); got != " 磁盘测试未返回可用的性能数据。\n" {
		t.Fatalf("Chinese empty disk result = %q", got)
	}
}
