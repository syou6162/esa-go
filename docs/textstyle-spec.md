# textstyle 仕様

## 目的と背景

`textstyle` は、esa.io に書く記事本文の記法ルールを定義します。ラクガキ帳の entry、ストーリー記事の自由記述など、記事の format が違っても使い回せる判定だけを持ちます。

記法ルールを format ごとに書き写すと、同じ意図のルールが別の実装・別の文言に分かれます。format をまたいで同じ判定と同じ error message を使うため、判定はこの package に置きます。

## 全 format 共通のルールと、選んで使うルール

どの format でも守りたいルールは `Validate` にまとめます。

- 全角 colon
- 全角 parentheses
- 各行の行頭にある中黒 `・`

一方、format によって許容したいかどうかが変わるルールは、個別の function として export します。呼び出し側は必要なものだけを組み合わせます。

- `ValidateHeading`: Markdown heading
- `ValidateBold`: Markdown bold syntax

Markdown heading は、見出しで構造を作ってよい本文では許容したい一方、テンプレートで項目が決まっている本文では勝手な項目追加になります。この判断は format 側が持ちます。

中黒の検証対象は各行の行頭だけです。行中の中黒は日本語の並列表記として使えるため reject しません。

## この package が持たないルール

特定の format にだけ意味があるルールは持ちません。たとえばラクガキ帳の entry は水平線 `---` を entry の separator として使い、本文の先頭へ system が anchor を連結します。したがって `---` と、本文全体の先頭の時刻表記・list marker の reject は entry format の制約であり、`scratchpad` package が持ちます。

## API と挙動

各 function は `func(text string) []string` の形で、見つかった issue の日本語 message を返します。issue が無い場合は nil を返します。

- 入力を trim・正規化・補正しません。判定は受け取った text に対してそのまま行います。
- `Validate` は最初の違反で打ち切らず、共通ルールの issue をすべて返します。呼び出し側が一度に全部提示できるようにするためです。
- issue の順序は安定しています。
- 空文字列は issue なしです。本文が必須かどうかは format 側の判断なので、この package では空入力を error にしません。

error 型は持ちません。呼び出し側が自分の format 固有の issue と結合し、その package の validation error として返します。
