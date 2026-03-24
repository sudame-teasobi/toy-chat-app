# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

CQRS（Command Query Responsibility Segregation）+ イベントソーシングアーキテクチャを実装するチャットアプリケーションの学習プロジェクト。

**アーキテクチャフロー:**
- 書き込み側: GoサーバーがTiDBに状態を書き込む
- イベント: TiCDCが変更をキャプチャしてKafkaにストリーミング
- 読み込み側: Kafkaからイベントを投影してリードモデルを構築

## ビルドコマンド

```bash
# インフラ起動（TiDB, Kafka, TiCDC, DynamoDB）
docker compose up -d

# データベースマイグレーション実行
go run ./cmd/migrate/main.go -cmd up

# マイグレーションロールバック
go run ./cmd/migrate/main.go -cmd down -steps 1

# 書き込みサーバー起動
go run ./cmd/write-server/main.go

# SQLクエリからコード生成
sqlc generate

# 依存関係整理
go mod tidy
```

## テスト

```bash
# 全テスト実行
go test ./...

# 特定パッケージのテスト
go test ./internal/applicationservice/...

# 詳細出力付き
go test -v ./...
```

## 環境変数

write-serverに必要:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` - TiDB接続情報
- `KAFKA_BROKER` - Kafkaブローカーアドレス

ローカルデフォルト値: `localhost:4000`（TiDB）, `localhost:9092`（Kafka）

## データベース

- **エンジン**: TiDB（MySQL互換の分散データベース）
- **マイグレーション**: golang-migrateによる番号付きup/downファイル
- **コード生成**: SQLCが`sql/queries/*.sql`から型付きGoコードを生成

テーブル: `users`, `chat_rooms`, `chat_room_members`, `messages`, `event_records`

## インフラストラクチャ

Docker Composeサービス:

**TiDBクラスタ:**
- **tidb-pd**（ポート2379） - Placement Driver（クラスタメタデータ管理）
- **tidb-tikv**（ポート20160） - 分散KVストレージ
- **tidb-tidb**（ポート4000） - MySQL互換SQLレイヤー
- **tidb-ticdc**（ポート8300） - Kafkaへの変更データキャプチャ
- **tidb-ticdc-init** - TiCDC changefeed初期化用

**その他:**
- **Kafka**（ポート9092） - イベントストリーミング
- **Kafka-UI**（ポート8181） - Kafka監視用Webインターフェース
- **DynamoDB**（ポート8000） - ローカル開発用（リードモデル用）

## 開発メモ

- TiDBはgolang-migrate互換のため`skip-isolation-level-check = true`で設定されている
- SQLC設定: クエリは`sql/queries/`、出力先は`internal/db/`
- このプロジェクトではTDDスキルが有効

## Go バージョン: 1.26（2026年2月リリース）

このプロジェクトは Go 1.26 を使っています。

### 言語仕様の変更

- **`new(expr)` の拡張**: `new` 組み込み関数の引数に式を渡せるようになった。
  型だけでなく初期値も同時に指定できる。

  ```go
  // 旧: 2ステップ必要だった
  age := 30
  p.Age = &age

  // 新: 1ステップで書ける
  p.Age = new(yearsSince(born))
  ```

  JSON/Protobuf のオプショナルポインタフィールドの初期化が簡潔になる。

- **ジェネリック型の自己参照**: ジェネリック型が自身の型パラメータリスト内で
  自己参照できるようになった。複雑なデータ構造やインターフェースの実装が簡略化される。

### 標準ライブラリ（新規・変更）

- **`crypto/hpke`（新パッケージ）**: RFC 9180 準拠の Hybrid Public Key Encryption。
  ポスト量子ハイブリッド KEM もサポート。

- **`crypto/mlkem/mlkemtest`（新パッケージ）**: ML-KEM の既知答えテスト用。

- **`testing/cryptotest`（新パッケージ）**: 暗号実装のテストユーティリティ。

- **`crypto/fips140`**: `WithoutEnforcement` / `Enforced` 関数が追加。
  `GODEBUG=fips140=only` モード下でも FIPS 140-3 厳格チェックを選択的に無効化可能。

- **`crypto/rand.Prime`**: `random` 引数が無視されるようになった（破壊的変更なし）。

### ツール

- **`go fix` の刷新**: `go vet` と同じ analysis フレームワーク採用。
  古い歴史的 fixer は削除され、新しい言語・ライブラリの新機能を使うよう提案する
  アナライザに置き換えられた。

- **`go tool doc` / `cmd/doc` の削除**: 代わりに `go doc` を使う（フラグ・引数は同じ）。

- **pprof Web UI**: `-http` フラグ使用時のデフォルトビューがフレームグラフに変更。
  旧グラフビューは "View -> Graph" メニューで引き続き利用可能。

