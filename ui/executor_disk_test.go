package ui

import "testing"

func TestDiskResultTextNeverLeavesAnEmptyEnglishSection(t *testing.T) {
	if got := diskResultText(" EN ", "\n\t"); got != " Disk test unavailable\n" {
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
	if got := diskResultText("zh", ""); got != " 硬盘测试不可用\n" {
		t.Fatalf("Chinese empty disk result = %q", got)
	}
}
