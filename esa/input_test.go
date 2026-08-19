package esa

import (
	"errors"
	"strconv"
	"testing"
)

func TestParseTag(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{"empty", "", "バリデーションエラー: タグが空です"},
		{"leading space", " tag", "バリデーションエラー: タグの前後に空白は使用できません"},
		{"trailing space", "tag ", "バリデーションエラー: タグの前後に空白は使用できません"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTag(tt.input)
			if err == nil {
				t.Fatalf("ParseTag(%q) succeeded", tt.input)
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

func TestTagZeroValueAndComparable(t *testing.T) {
	var zero Tag
	if zero.String() != "" {
		t.Fatalf("zero String() = %q", zero.String())
	}
	first, err := ParseTag("tag")
	if err != nil {
		t.Fatal(err)
	}
	second, err := ParseTag("tag")
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("equal parsed Tag values are not comparable")
	}
	values := map[Tag]string{first: "value"}
	if values[second] != "value" {
		t.Fatal("Tag cannot be used as an equivalent map key")
	}
}

func TestNewValidationErrorCopiesIssues(t *testing.T) {
	issues := []string{"タグが空です", "別の問題"}
	err := NewValidationError(issues)
	issues[0] = "変更された問題"

	if err.Message != "バリデーションエラー: タグが空です; 別の問題" {
		t.Fatalf("Message = %q", err.Message)
	}
	if err.Error() != err.Message {
		t.Fatalf("Error() = %q, want Message %q", err.Error(), err.Message)
	}
	if err.Issues[0] != "タグが空です" {
		t.Fatalf("Issues = %#v", err.Issues)
	}
}

func TestParsePostNumber(t *testing.T) {
	for _, raw := range []int{0, -1} {
		_, err := ParsePostNumber(raw)
		if err == nil {
			t.Fatalf("ParsePostNumber(%d) succeeded", raw)
		}
		want := "バリデーションエラー: 記事番号は正の整数である必要があります (post_number=" +
			strconv.Itoa(raw) + ")"
		if err.Error() != want {
			t.Fatalf("error = %q, want %q", err, want)
		}
		var validationErr *ValidationError
		if !errors.As(err, &validationErr) {
			t.Fatalf("error = %T, want *ValidationError", err)
		}
	}

	got, err := ParsePostNumber(42)
	if err != nil {
		t.Fatalf("ParsePostNumber(42) error = %v", err)
	}
	if got.Int() != 42 {
		t.Fatalf("Int() = %d, want 42", got.Int())
	}
}

func TestParseRevisionNumber(t *testing.T) {
	for _, raw := range []int{0, -1} {
		_, err := ParseRevisionNumber(raw)
		if err == nil {
			t.Fatalf("ParseRevisionNumber(%d) succeeded", raw)
		}
		want := "バリデーションエラー: リビジョン番号は正の整数である必要があります (revision_number=" +
			strconv.Itoa(raw) + ")"
		if err.Error() != want {
			t.Fatalf("error = %q, want %q", err, want)
		}
		var validationErr *ValidationError
		if !errors.As(err, &validationErr) {
			t.Fatalf("error = %T, want *ValidationError", err)
		}
	}

	got, err := ParseRevisionNumber(5)
	if err != nil {
		t.Fatalf("ParseRevisionNumber(5) error = %v", err)
	}
	if got.Int() != 5 {
		t.Fatalf("Int() = %d, want 5", got.Int())
	}
}

func TestPostNumberZeroValueAndComparable(t *testing.T) {
	var zero PostNumber
	if zero.Int() != 0 {
		t.Fatalf("zero Int() = %d", zero.Int())
	}
	first, err := ParsePostNumber(42)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ParsePostNumber(42)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("equal parsed PostNumber values are not comparable")
	}
	values := map[PostNumber]string{first: "value"}
	if values[second] != "value" {
		t.Fatal("PostNumber cannot be used as an equivalent map key")
	}
}

func TestRevisionNumberZeroValueAndComparable(t *testing.T) {
	var zero RevisionNumber
	if zero.Int() != 0 {
		t.Fatalf("zero Int() = %d", zero.Int())
	}
	first, err := ParseRevisionNumber(5)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ParseRevisionNumber(5)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("equal parsed RevisionNumber values are not comparable")
	}
	values := map[RevisionNumber]string{first: "value"}
	if values[second] != "value" {
		t.Fatal("RevisionNumber cannot be used as an equivalent map key")
	}
}
