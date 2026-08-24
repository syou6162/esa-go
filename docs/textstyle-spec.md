# textstyle 仕様

## 目的と背景

`textstyle` は、esa.io に書く記事本文の記法ルールを定義します。ラクガキ帳の entry、ストーリー記事の自由記述など、記事の format が違っても共通して守りたい記法だけを持ちます。

記法ルールを format ごとに書き写すと、同じ意図のルールが別の実装・別の文言に分かれます。format をまたいで同じ判定と同じ error message を使うため、共通ルールはこの package に置き、format 固有のルールはその format を所有する package に残します。

## 対象とするルール

`Validate` は次の構造を issue として報告します。

- Markdown の水平線として解釈される separator
- Markdown heading
- Markdown bold syntax
- 全角 colon
- 全角 parentheses
- 各行の行頭にある中黒 `・`

ただし、table row の一部として使われる separator は例外です。table の cell 区切りを含む行は、水平線だけを意図した入力と同じように reject しません。

中黒の検証対象は各行の行頭だけです。行中の中黒は日本語の並列表記として使えるため reject しません。

## この package が持たないルール

特定の format にだけ意味があるルールは持ちません。たとえばラクガキ帳の entry は本文の先頭へ system が anchor を連結するため、本文全体の先頭の時刻表記と list marker を reject しますが、これは entry format の制約であり `scratchpad` package が持ちます。

## API と挙動

`Validate(text string) []string` は、見つかった issue の日本語 message を返します。issue が無い場合は nil を返します。

- 入力を trim・正規化・補正しません。判定は受け取った text に対してそのまま行います。
- 最初の違反で打ち切らず、該当するルールの issue をすべて返します。呼び出し側が一度に全部提示できるようにするためです。
- issue の順序は上記ルールの並び順で安定しています。
- 空文字列は issue なしです。本文が必須かどうかは format 側の判断なので、この package では空入力を error にしません。

error 型は持ちません。呼び出し側が自分の format 固有の issue と結合し、その package の validation error として返します。
