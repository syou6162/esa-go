# esa-go

esa.io の REST client とラクガキ帳（scratchpad）のロジックを提供する Go module です。

## パッケージ

| パッケージ | import path | 役割 |
| --- | --- | --- |
| `esa` | `github.com/syou6162/esa-go/esa` | esa.io REST client |
| `scratchpad` | `github.com/syou6162/esa-go/scratchpad` | ラクガキ帳の本文フォーマットに関する純ロジック |

設計方針は [`docs/design-guidelines.md`](docs/design-guidelines.md)、scratchpad の仕様は [`docs/scratchpad-spec.md`](docs/scratchpad-spec.md) にまとめています。

## Revision API について（beta）

`esa.Client` の `ListRevisions` / `GetRevision` / `CompareRevisions` / `RollbackRevision` が叩く esa.io の Revision API は beta 公開で、[公式の API ドキュメント](https://docs.esa.io/posts/102) には未記載です。一次情報は [esa-ruby の README](https://github.com/esaio/esa-ruby) と [esa-mcp-server #338](https://github.com/esaio/esa-mcp-server/pull/338) の API 定義で、仕様は予告なく変わる可能性があります。

`RollbackRevision` は指定リビジョンのタイトル・カテゴリ・タグ・本文へ記事を戻し、その結果を新しい最新リビジョンとして積みます（履歴は消えません）。`WIP` と `Message` は任意で、nil のときはリクエスト body に含めず esa.io の既定動作に任せます。最新リビジョン自身への rollback は esa.io が HTTP 400 を返しますが、理由は machine-readable な形では返らないため、esa-go は status code と response body を error message に載せてそのまま返し、解釈は呼び出し側に委ねます。記事・リビジョンが存在しない場合は参照系（`ListRevisions` / `GetRevision` / `CompareRevisions`）が `esa.ErrNotFound` になります。

## ローカルでの検証

```bash
gofmt -l .
go vet ./...
go test ./...
go build ./...
```

`gofmt -l .` の出力が空であること、`go vet` / `go test` / `go build` がエラーなく完了することを確認してください。

## ライセンス

MIT License（[LICENSE](./LICENSE) を参照）
