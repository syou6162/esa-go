# esa-go

esa.io の REST client とラクガキ帳（scratchpad）のロジックを提供する Go module です。

## パッケージ

| パッケージ | import path | 役割 |
| --- | --- | --- |
| `esa` | `github.com/syou6162/esa-go/esa` | esa.io REST client |
| `scratchpad` | `github.com/syou6162/esa-go/scratchpad` | ラクガキ帳の本文フォーマットに関する純ロジック |

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
