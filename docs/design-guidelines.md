# esa-go 設計方針

この文書は、`github.com/syou6162/esa-go` を複数の利用側アプリケーションから安全に使える public module として保つための設計方針を定めます。

## 入力は境界で parse する

外部から受け取った文字列や数値は、利用側の処理へ渡す前に意味のある型へ parse します。parse に成功した値は、その型が表す不変条件を満たしているものとして扱い、下流の各関数で primitive の妥当性を繰り返し疑いません。

parse に失敗した入力を、trim、既定値、zero value などで黙って救済してはいけません。入力の前後の whitespace が仕様上意味を持つ場合は、raw input を変更せずに reject します。

誤用しやすい値には専用型を使います。たとえば `TimestampID` は単なる文字列ではなく、`ParseTimestampID` の成功後に規則を満たす値として扱います。

## parse 後の値を補正しない

parse 済みの値を下流で防御的に trim、正規化、既定値置換してはいけません。特に scratchpad の本文、Markdown、entry text では whitespace や newline が表示内容と構造の一部です。入力を保持したまま処理し、許可できない構造はエラーにします。

正規化が仕様として定義されている場合だけ、境界で明示的に行います。たとえば本文 parser が CRLF と CR を LF に統一する場合、その変換は parser の責務として一度だけ行います。

## fallback より失敗を優先する

基本動作は、reject、error propagation、または呼び出し側が観測できる failure です。unknown な入力、欠落した必須値、decode failure、外部 API の異常を、成功や既定値へ黙って変換してはいけません。

呼び出し側が処理を分岐する失敗には、sentinel error または typed error を使います。呼び出し側は error message の文字列比較ではなく、`errors.Is` や `errors.As` で判定できるようにします。

invalid input や empty input を、成功を意味する zero value に潰してはいけません。空入力が仕様上「今日」などの意味を持つ場合は、その規則を公開 API に明記し、暗黙の fallback と区別します。

## unknown な値は fail-closed に扱う

enum、形式、操作種別、許可リストなどの unknown な値を、適当な default として受理しません。解釈できない値は reject します。

allowlist は fail-closed で評価します。許可されていることを確認できない値は拒否し、許可タグやカテゴリなどのアプリ固有ポリシーは利用側から明示的に渡します。

## エラーには診断情報を含める

エラーには、可能な範囲で対象、実行した操作、具体的な理由を含めます。下位層のエラーは `%w` で wrap し、`errors.Is` / `errors.As` による判定可能性を保ちます。

一方、エラーメッセージに秘密情報を含めてはいけません。

- URL の query と fragment をそのまま表示しない
- URL の userinfo/password を表示しない
- token や `Authorization` header の値を表示しない
- 外部 upload provider などの URL が含まれる下位エラーも redaction する
- malformed URL を raw のまま error に echo しない

redaction は表示用の error message に適用し、機械判定に必要な wrapped error の関係は維持します。

## 小さな role interface を提供する

interface は利用者が必要とする機能のまとまりごとに分けます。たとえば投稿の read、write、本文だけの update、画像 upload などを別 role として提供し、利用側アプリケーションが必要な role だけを合成できるようにします。

esa-go は全部入りの `Client` interface を export しません。利用側は自分の package 内で必要なメソッドだけを持つ private interface を定義できます。

`NewClient` は具象型を返します。利用側が role interface を必要とする場合は、その具象型を代入して使います。

HTTP client のテストでは `httptest.Server` を使い、method、path、query、header、payload、response の decode を通信層のテストとして検証します。利用側アプリケーションの workflow テストは、role interface の spy を使って HTTP の詳細から分離します。

## package の責務を分ける

### `esa` package

`esa` package は esa.io API の wire mechanics を担当します。

- HTTP request/response の組み立て
- authentication header
- endpoint と pagination
- JSON encode/decode
- esa API の response model
- HTTP error と安全な URL redaction
- esa API に対する role interface

本文の意味、記事名の policy、カテゴリの組み立て、タグの allowlist、利用側の workflow は担当しません。

### `scratchpad` package

`scratchpad` package は scratchpad 本文の format と domain validation を担当します。

- `TimestampID`
- anchor と entry URL
- entry の parse/serialize
- 降順 ordering
- timestamp collision avoidance
- text/title validation
- pure な entry manipulation

team、カテゴリ prefix、記事名、許可タグ、日付の利用側 policy は package に固定しません。

### 利用側アプリケーション

利用側アプリケーションは、次の責務を担当します。

- workflow と esa API 呼び出しの組み合わせ
- team、カテゴリ、記事名、タグなどの設定
- 日付や曜日の policy
- lock による直列化などのアプリ固有の並行制御
- protocol adapter、tool result、設定ファイルの読み込み

esa-go はこれらへ依存せず、利用側が必要な順序付けや retry、revision を使った競合処理を選べるようにします。
