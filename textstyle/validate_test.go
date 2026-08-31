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
		{"colon", "key：value", "全角コロン"},
		{"parentheses", "text（value）", "全角括弧"},
		{"middle dot", "・ item", "中黒"},
		{"indented middle dot", "text\n  ・item", "中黒"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := Validate(tt.text)
			if len(issues) == 0 || !strings.Contains(issues[0], tt.want) {
				t.Fatalf("issues = %#v, want %q", issues, tt.want)
			}
		})
	}
}

func TestValidateAllowsTextWithoutCommonIssues(t *testing.T) {
	for _, text := range []string{
		"normal text\n- item\n12:30 is allowed here",
		"text\n---\nmore",
		"# heading",
		"**bold**",
		"りんご・みかんを買う",
		"",
	} {
		if issues := Validate(text); len(issues) != 0 {
			t.Fatalf("text %q issues = %#v", text, issues)
		}
	}
}

func TestValidateMultipleViolations(t *testing.T) {
	issues := Validate("key：value（補足）\n・item")
	if len(issues) < 3 {
		t.Fatalf("issues = %#v, want at least 3", issues)
	}
}

func TestValidateHeading(t *testing.T) {
	for _, text := range []string{"# heading", "text\n### heading"} {
		issues := ValidateHeading(text)
		if len(issues) == 0 || !strings.Contains(issues[0], "見出し") {
			t.Fatalf("text %q issues = %#v", text, issues)
		}
	}
	for _, text := range []string{"#hashtag", "text"} {
		if issues := ValidateHeading(text); len(issues) != 0 {
			t.Fatalf("text %q issues = %#v", text, issues)
		}
	}
}

func TestValidateBold(t *testing.T) {
	issues := ValidateBold("**bold**")
	if len(issues) == 0 || !strings.Contains(issues[0], "ボールド") {
		t.Fatalf("issues = %#v", issues)
	}
	if issues := ValidateBold("a * b * c"); len(issues) != 0 {
		t.Fatalf("issues = %#v", issues)
	}
}
