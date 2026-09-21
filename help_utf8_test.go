package flags

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestWrapTextPreservesUTF8(t *testing.T) {
	for _, text := range []string{"12345678éabcdef", "1234567😀abcdef", "你好世界这是测试文本内容"} {
		t.Run(text, func(t *testing.T) {
			got := wrapText(text, 10, "> ")
			if !utf8.ValidString(got) {
				t.Fatalf("invalid UTF-8: %q", got)
			}
			if strings.Contains(got, "\n\n") {
				t.Errorf("extra blank line in wrapped word: %q", got)
			}
			restored := strings.ReplaceAll(got, "-\n> ", "")
			if restored != text {
				t.Errorf("wrapped text lost content: %q", got)
			}
			for i, line := range strings.Split(got, "\n") {
				if i > 0 {
					line = strings.TrimPrefix(line, "> ")
				}
				if utf8.RuneCountInString(line) > 10 {
					t.Errorf("line exceeds rune limit: %q", line)
				}
			}
		})
	}
}

func TestWrapLongWordHasOneLineBreak(t *testing.T) {
	got := wrapText("abcdefghijklmn", 10, "> ")
	if want := "abcdefghi-\n> jklmn"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
