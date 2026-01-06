# toy-chat-app

## コマンド

### TiCDC Changefeed 管理

TiCDC の changefeed を管理するコマンドです。

```bash
# changefeed 一覧を取得
go run ./cmd/ticdc/main.go -cmd list

# changefeed を作成
go run ./cmd/ticdc/main.go -cmd create

# changefeed を削除
go run ./cmd/ticdc/main.go -cmd delete
```
