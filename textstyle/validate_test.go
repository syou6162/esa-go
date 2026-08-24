package textstyle

import (
	"strings"
	"testing"
)

func TestValidateIssues(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"separator", "text\n---\nmore", "区切り"},
		{"table separator", "| a | b |\n| --- | --- |", ""},
		{"heading", "# heading", "見出し"},
		{"heading on second line", "text\n### heading", "見出し"},
		{"bold", "**bold**", "ボールド"},
		{"colon", "key：value", "全角コロン"},
		{"parentheses", "text（value）", "全角括弧"},
		{"middle dot", "・ item", "中黒"},
		{"indented middle dot", "text\n  ・item", "中黒"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := Validate(tt.text)
			if tt.want == "" {
				if len(issues) != 0 {
					t.Fatalf("issues = %#v", issues)
				}
				return
			}
			if len(issues) == 0 || !strings.Contains(issues[0], tt.want) {
				t.Fatalf("issues = %#v, want %q", issues, tt.want)
			}
		})
	}
}

func TestValidateAllowsNormalText(t *testing.T) {
	for _, text := range []string{
		"normal text\n- item\n12:30 is allowed here",
		"りんご・みかんを買う",
		"#hashtag",
		"",
	} {
		if issues := Validate(text); len(issues) != 0 {
			t.Fatalf("text %q issues = %#v", text, issues)
		}
	}
}

func TestValidateMultipleViolations(t *testing.T) {
	issues := Validate("# heading\n---\n**bold**")
	if len(issues) < 3 {
		t.Fatalf("issues = %#v, want at least 3", issues)
	}
}
