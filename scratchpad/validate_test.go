package scratchpad

import (
	"errors"
	"strings"
	"testing"
)

func TestValidatePostTextIssues(t *testing.T) {
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
		{"time", "15:30 memo", "時刻"},
		{"list", "- item", "マークダウン"},
		{"asterisk list", "* item", "マークダウン"},
		{"middle dot", "・ item", "中黒"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := ValidatePostText(tt.text)
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

func TestValidatePostTextAllowsNormalText(t *testing.T) {
	for _, text := range []string{
		"normal text\n- item\n12:30 is allowed here",
		"  9:05 is allowed here",
		"#hashtag",
	} {
		if issues := ValidatePostText(text); len(issues) != 0 {
			t.Fatalf("text %q issues = %#v", text, issues)
		}
	}
}

func TestValidatePostTextMultipleViolations(t *testing.T) {
	issues := ValidatePostText("# heading\n---\n**bold**")
	if len(issues) < 3 {
		t.Fatalf("issues = %#v, want at least 3", issues)
	}
}

func TestValidatePostTextMiddleDotScope(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{"first line", "・ item", true},
		{"second line", "heading\n・ item", true},
		{"middle of line", "A・B", false},
		{"leading whitespace", " ・ item", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := ValidatePostText(tt.text)
			hasMiddleDotIssue := false
			for _, issue := range issues {
				if strings.Contains(issue, "中黒") {
					hasMiddleDotIssue = true
					break
				}
			}
			if hasMiddleDotIssue != tt.want {
				t.Fatalf("issues = %#v, middle-dot issue = %v, want %v", issues, hasMiddleDotIssue, tt.want)
			}
		})
	}

	issues := ValidatePostText("・ item\n・ another\n# heading")
	if len(issues) != 2 || !strings.Contains(issues[0], "見出し") && !strings.Contains(issues[1], "見出し") {
		t.Fatalf("issues = %#v, want one heading issue and one middle-dot issue", issues)
	}
	middleDotCount := 0
	for _, issue := range issues {
		if strings.Contains(issue, "中黒") {
			middleDotCount++
		}
	}
	if middleDotCount != 1 {
		t.Fatalf("issues = %#v, want one middle-dot issue for multiple matching lines", issues)
	}

	issues = ValidatePostText("・ item\n---\n# heading")
	if len(issues) != 3 {
		t.Fatalf("issues = %#v, want separator, heading, and middle-dot issues", issues)
	}
}

func TestCheckTextValidationReturnsTypedError(t *testing.T) {
	if err := CheckTextValidation("normal"); err != nil {
		t.Fatalf("valid text error = %v", err)
	}
	err := CheckTextValidation("# invalid")
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) || validationErr.Message == "" {
		t.Fatalf("error = %#v", err)
	}
	constructed := NewValidationError([]string{"custom"})
	if constructed.Error() != "バリデーションエラー: custom" {
		t.Fatalf("NewValidationError() = %q", constructed)
	}
}

func TestValidateTitleIssues(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"slash", "title/path", "スラッシュ"},
		{"fullwidth space", "title　name", "全角スペース"},
		{"leading symbol", "#title", "先頭"},
		{"period", "title.name", "ピリオド"},
		{"full stop", "title。name", "ピリオド"},
		{"punctuation", "what?!", "感嘆符"},
		{"newline", "title\nname", "改行"},
		{"carriage return", "title\rname", "改行"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := ValidateScratchpadTitle(tt.title)
			if len(issues) == 0 || !strings.Contains(issues[0], tt.want) {
				t.Fatalf("issues = %#v, want %q", issues, tt.want)
			}
		})
	}
}

func TestValidateTitleAllowsNormalTitle(t *testing.T) {
	for _, title := range []string{"今日のメモ", "Example title #1", "カタカナ"} {
		if issues := ValidateScratchpadTitle(title); len(issues) != 0 {
			t.Fatalf("title %q issues = %#v", title, issues)
		}
	}
	if issues := ValidateScratchpadTitle("ｶﾀｶﾅ"); len(issues) != 0 {
		t.Fatalf("halfwidth katakana title issues = %#v", issues)
	}
}

func TestCheckTitleValidationReturnsTypedError(t *testing.T) {
	if err := CheckTitleValidation("valid title"); err != nil {
		t.Fatalf("valid title error = %v", err)
	}
	err := CheckTitleValidation("invalid/title")
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %#v", err)
	}
}
