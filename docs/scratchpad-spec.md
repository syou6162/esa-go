# scratchpad 仕様

## 目的と背景

scratchpad は、esa.io 上の一つの記事へ、メモ・出来事・会話ログなどを時系列の entry として追記していく仕組みです。日単位で一つの記事を扱い、その記事の中に複数の entry を保存します。

この形式により、記事を通常の Markdown として読める状態を保ちながら、各 entry を timestamp で参照できます。新しい entry は記事の先頭側に置き、既存 entry を失わずに追加・更新・削除できることを不変条件とします。

記事の検索、作成、更新をどの順序で行うか、記事を誰が所有するか、どのカテゴリ・記事名・タグを使うかは利用側アプリケーションの責務です。

## 用語

- **post**: esa.io の一つの記事。日単位の scratchpad の保存先です。
- **entry**: post の本文中に保存される一件のメモ。`TimestampID` と本文 text から構成されます。
- **TimestampID**: entry を識別する 12 桁の timestamp ID。
- **anchor**: entry の先頭に置く HTML anchor。ID への fragment link としても機能します。
- **separator**: entry 同士を区切る Markdown の水平線。

## 本文フォーマット

各 entry は次の形で表します。

```text
<a id="HHMMSSffffff" href="#HHMMSSffffff">HH:MM</a> text
```

anchor の後には半角スペースを一つ置き、その後に entry の text を続けます。text 内の newline は保持します。

entry 同士は separator line で区切ります。serialize された各 entry の末尾には `\n\n---` を置き、複数 entry の間にはさらに `\n\n` を置きます。そのため、separator は本文上では水平線の行として現れ、parser は `\n---` を entry boundary として認識します。末尾には serialize によって空の block が生じますが、parser はそれを許容します。

実装上の parser は CRLF と CR を LF に正規化してから entry を分割します。anchor がない block、anchor の ID が不正な block、許可されない構造は parse error になります。

## TimestampID

TimestampID は次の 12 桁です。

```text
HHMMSSffffff
```

- `HH`: 00 から 23
- `MM`: 00 から 59
- `SS`: 00 から 59
- `ffffff`: microsecond の 6 桁

`time.Time` から生成する場合、時・分・秒はその値から取り、fractional part は nanosecond を microsecond に変換して 6 桁で表します。

入力はちょうど 12 桁の ASCII 数字でなければなりません。時・分・秒の範囲外は reject します。TimestampID は opaque な値型で、`ParseTimestampID` または `NewTimestampIDFromTime` を経由してのみ生成できます。zero value は unset を表し、`String`、`DisplayTime`、`AnchorHTML` は空文字列を返します。TimestampID は parse 成功後に専用型として扱い、下流の処理で raw string を再検証しません。

### 一意性と衝突回避

一つの post 内で TimestampID は一意でなければなりません。同じ ID が既に存在する場合、既存 entry を上書きするのではなく、numeric ID を最小限だけ増分して、parse 可能で未使用の ID を選びます。

候補を探している間に時・分・秒の範囲を超えるなど parse 不可能になった候補はスキップします。有限回数の候補を調べても未使用 ID が見つからない場合は error とします。

この collision avoidance は entry data を壊さないための pure logic です。同じ post を複数の処理が同時に更新する場合の lock や cross-process coordination は含みません。

## entry の順序

本文は TimestampID の降順、つまり新しい entry から古い entry の順に serialize します。

`Sorted` は元の slice を変更せず、降順の copy を返します。`Update` と `Delete` も、元の collection を意図せず変更しないように扱います。

`Add` は receiver がすでに降順であることを前提に、該当位置へ entry を追加します。同じ TimestampID があれば entry を置き換えます。利用側は、未整列の入力を `Add` へ渡す前に `Sorted` を使う必要があります。

## 本文の validation

本文 text では、次の構造を reject します。

- Markdown の水平線として解釈される separator
- Markdown heading
- Markdown bold syntax
- 全角 colon
- 全角 parentheses
- 本文全体の先頭（システムが anchor を連結する位置）の `HH:MM` のような時刻
- 本文全体の先頭（システムが anchor を連結する位置）の `- ` または `* ` の list marker

時刻と list marker の検証対象は本文全体の先頭だけです。2 行目以降に
時刻表記や list marker があっても reject しません。

ただし、table row の一部として使われる separator は例外です。table として解釈できる行を、水平線だけを意図した入力と同じように reject してはいけません。

validation は入力を別の本文へ変換するものではありません。禁止される構造を見つけたら具体的な validation error を返し、入力 text 自体は保持します。

## タイトルの validation

scratchpad post のタイトルでは、次を reject します。

- `/`
- 全角 space
- 許可された文字種ではない先頭文字
- `.` と日本語の full stop
- ASCII または全角の `!` / `?`
- newline と carriage return

タイトルの文字種や固定の名称は利用側アプリケーションが決めます。esa-go は特定の team、カテゴリ prefix、記事名を既定値として持ちません。

## 入力型と parse 経路

本文とタイトルは、validation 済みの opaque な値型として扱います。

- `PostText` は `ParsePostText` の成功後にだけ生成できます。
- `PostTitle` は `ParsePostTitle` の成功後にだけ生成できます。
- raw string が必要な場合は、それぞれの `String()` を使います。
- zero value は unset を表し、`String()` は空文字列を返します。空入力を受理した成功値と混同してはいけません。
- `PostText` と `PostTitle` は比較可能な値型で、`==` 比較と map key に使えます。

`ParsePostText` は空文字列を reject し、本文 validation の全ての問題を
`*ValidationError` として返します。`ParsePostTitle` は空文字列と前後の空白を
reject した上で、タイトル validation を行います。parse 済みの値を下流で raw
string に戻して再検証することはしません。

タグと記事番号も、esa package の opaque な parse 済み型として扱います。

- `Tag` は `ParseTag` の成功後にだけ生成でき、空文字列と前後の空白を reject します。
- `PostNumber` は `ParsePostNumber` の成功後にだけ生成でき、正の整数だけを受理します。
- raw tag は `Tag.String()`、記事番号は `PostNumber.Int()` で取得します。
- いずれも opaque な比較可能な値型で、`==` 比較と map key に使えます。

これらの parser は空入力や不正入力を zero value に変換せず、型付き
validation error として返します。

## 文字数制限

現在の本文・タイトル validation の実装と仕様には、文字数の上限・下限を設けていません。長さ制限を追加する場合は、既存の validation と別の仕様変更として明示してください。

## anchor と entry URL

anchor は TimestampID を `id` 属性に使い、表示部分は TimestampID の
`HH:MM` 表現です。parser は `id` から TimestampID を取り出し、`href` の
fragment 値は検証しません。

post URL と TimestampID から entry URL を作る場合は、次のように fragment を追加します。

```text
post URL + "#" + TimestampID
```

post URL または TimestampID が空の場合、entry URL は空文字とします。URL を作る処理は、post URL に既存の query や fragment がある場合の扱いを利用側と API 仕様で明確にしてください。

## Error handling

呼び出し側が処理を分岐できるよう、失敗の種類を sentinel error または typed error で区別します。

- 入力不正: validation error
- TimestampID や本文 format の不正: parse error
- 指定された entry が存在しない: entry-not-found error
- collection の collision candidate を使い切った: collision error

error には対象と具体的な理由を含めます。error message の文字列比較を利用せず、`errors.Is` と `errors.As` を使って判定します。

post が存在しない、カテゴリが所有対象か、タグが allowlist に含まれるかといった API/workflow の失敗は、利用側アプリケーションが esa client と scratchpad core を組み合わせて扱います。

## Scope 外

次の機能は scratchpad の pure core および public module の共通仕様には含みません。

- 同一 TimestampID を跨いだ複数処理の競合制御
- cross-process concurrency
- esa revision を使った optimistic locking
- 日付の選択や日付文字列の利用側 policy
- 曜日やタグの policy
- team、カテゴリ prefix、記事名、許可タグなどのアプリ固有値
- 記事の作成・検索・更新を組み合わせた workflow
- protocol adapter や利用側の result schema

これらは利用側アプリケーションが、esa の role interface と scratchpad の pure logic を組み合わせて実装します。
