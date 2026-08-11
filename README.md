# esa-go

esa.io REST client とラクガキ帳（scratchpad）の共通 Go module です。

[`syou6162/times-agent-talk`](https://github.com/syou6162/times-agent-talk)（Slack bot）と [`syou6162/stackchan-mcp`](https://github.com/syou6162/stackchan-mcp)（MCP server）の両方で重複していた esa.io 連携とラクガキ帳ロジックを一本化するために作成しました。両アプリから import される public module です。

将来 `esa-rs` / `esa-py` など他言語向けの実装が増える想定で、リポジトリ名に `-go` が付いています。

## Module path

```
github.com/syou6162/esa-go
```

## パッケージ

| パッケージ | import path | 役割 |
| --- | --- | --- |
| `esa` | `github.com/syou6162/esa-go/esa` | esa.io REST client（posts の検索・取得・作成・更新、タグ更新、画像アップロード、`PostReader` / `PostWriter` などの role interface） |
| `scratchpad` | `github.com/syou6162/esa-go/scratchpad` | ラクガキ帳の純ロジック（TimestampID、Entry/Entries の parse/serialize、validation、permalink 生成など） |

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
