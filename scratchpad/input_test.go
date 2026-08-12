package scratchpad

import (
	"errors"
	"testing"
)

func TestParsePostText(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{"empty", "", "バリデーションエラー: 本文が空です"},
		{"separator", "text\n---", "バリデーションエラー: 区切り線(---)が含まれています"},
		{"heading", "# heading", "バリデーションエラー: Markdown見出し(#)が使用されています"},
		{"heading on second line", "text\n### heading", "バリデーションエラー: Markdown見出し(#)が使用されています"},
		{"bold", "**bold**", "バリデーションエラー: ボールド体(**テキスト**)が使用されています"},
		{"colon", "key：value", "バリデーションエラー: 全角コロン(：)が使用されています。半角コロン(:)を使ってください"},
		{"parentheses", "text（value）", "バリデーションエラー: 全角括弧が使用されています。半角括弧を使ってください"},
		{"leading time", "15:30 memo", "バリデーションエラー: 先頭（行頭）に時刻が含まれています。時刻はシステムが自動挿入するため不要です"},
		{"leading list", "- item", "バリデーションエラー: 先頭（行頭）にマークダウンリスト記法(- / *)が使用されています。タイムスタンプ挿入でスタイルが崩れるため使用できません"},
		{"leading asterisk list", "* item", "バリデーションエラー: 先頭（行頭）にマークダウンリスト記法(- / *)が使用されています。タイムスタンプ挿入でスタイルが崩れるため使用できません"},
		{"leading middle dot", "・ item", "バリデーションエラー: 行頭の中黒(・)は使用できません。マークダウンリスト記法(- )を使ってください"},
		{"indented middle dot", "heading\n　・ item", "バリデーションエラー: 行頭の中黒(・)は使用できません。マークダウンリスト記法(- )を使ってください"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePostText(tt.input)
			if err == nil {
				t.Fatalf("ParsePostText(%q) succeeded", tt.input)
			}
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("error = %T, want *ValidationError", err)
			}
			if err.Error() != tt.want {
				t.Fatalf("error = %q, want %q", err, tt.want)
			}
		})
	}
}

func TestParsePostTextCombinesValidationIssues(t *testing.T) {
	_, err := ParsePostText("・ item\n---\n# heading")
	const want = "バリデーションエラー: 区切り線(---)が含まれています; Markdown見出し(#)が使用されています; 行頭の中黒(・)は使用できません。マークダウンリスト記法(- )を使ってください"
	if err == nil || err.Error() != want {
		t.Fatalf("error = %v, want %q", err, want)
	}
}

func TestParsePostTextPreservesRawValue(t *testing.T) {
	const raw = "  normal text  \n- detail"
	got, err := ParsePostText(raw)
	if err != nil {
		t.Fatalf("ParsePostText() error = %v", err)
	}
	if got.String() != raw {
		t.Fatalf("String() = %q, want %q", got.String(), raw)
	}
}

func TestPostTextZeroValueAndComparable(t *testing.T) {
	var zero PostText
	if zero.String() != "" {
		t.Fatalf("zero String() = %q", zero.String())
	}
	first, err := ParsePostText("first")
	if err != nil {
		t.Fatal(err)
	}
	second, err := ParsePostText("first")
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("equal parsed PostText values are not comparable")
	}
	values := map[PostText]string{first: "value"}
	if values[second] != "value" {
		t.Fatal("PostText cannot be used as an equivalent map key")
	}
}

func TestParsePostTitle(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{"empty", "", "バリデーションエラー: タイトルが空です"},
		{"surrounding space", " title ", "バリデーションエラー: タイトルの前後に空白は使用できません"},
		{"slash", "title/path", "バリデーションエラー: スラッシュ(/)はカテゴリ区切りとして解釈されるため使用できません"},
		{"fullwidth space", "title　name", "バリデーションエラー: 全角スペースは使用できません"},
		{"leading symbol", "#title", "バリデーションエラー: 先頭に記号は使用できません"},
		{"period", "title.name", "バリデーションエラー: ピリオド(.・。)は使用できません。区切りには読点(、)を使ってください"},
		{"full stop", "title。name", "バリデーションエラー: ピリオド(.・。)は使用できません。区切りには読点(、)を使ってください"},
		{"punctuation", "title!?", "バリデーションエラー: 感嘆符・疑問符(!?！？)は使用できません"},
		{"newline", "title\nname", "バリデーションエラー: 改行文字は使用できません"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePostTitle(tt.input)
			if err == nil {
				t.Fatalf("ParsePostTitle(%q) succeeded", tt.input)
			}
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("error = %T, want *ValidationError", err)
			}
			if err.Error() != tt.want {
				t.Fatalf("error = %q, want %q", err, tt.want)
			}
		})
	}
}

func TestParsePostTitlePreservesRawValue(t *testing.T) {
	const raw = "A title #1"
	got, err := ParsePostTitle(raw)
	if err != nil {
		t.Fatalf("ParsePostTitle() error = %v", err)
	}
	if got.String() != raw {
		t.Fatalf("String() = %q, want %q", got.String(), raw)
	}
}

func TestPostTitleZeroValueAndComparable(t *testing.T) {
	var zero PostTitle
	if zero.String() != "" {
		t.Fatalf("zero String() = %q", zero.String())
	}
	first, err := ParsePostTitle("first")
	if err != nil {
		t.Fatal(err)
	}
	second, err := ParsePostTitle("first")
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("equal parsed PostTitle values are not comparable")
	}
	values := map[PostTitle]string{first: "value"}
	if values[second] != "value" {
		t.Fatal("PostTitle cannot be used as an equivalent map key")
	}
}
