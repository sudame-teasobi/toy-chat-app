# Code Review: toy-chat-app

## 総合評価

CQRS + イベントソーシングの学習プロジェクトとして、全体的な設計方針は適切で、DDD のパターンに沿った構造になっている。一方で、エラーハンドリング・命名規則の一貫性・テストカバレッジにおいて改善の余地がある。

---

## 1. アーキテクチャレビュー

### 1.1 良い点

- **CQRS パターンの分離が明確**: Write 側（TiDB）と Read 側（DynamoDB）が物理的に分離されており、CQRS の本質を捉えている
- **イベント駆動アーキテクチャ**: TiCDC → Kafka → Consumer の流れにより、書き込みと読み込みが疎結合になっている
- **DDD の適用**: `internal/domain/` 配下に集約ルート（User, Room, Membership）が適切に定義され、イベントの発行まで含めたドメインモデルになっている
- **レイヤードアーキテクチャ**: handler → service → domain → repository の依存方向が一貫しており、ドメイン層がインフラ層に依存していない
- **Docker Compose の構成**: サービスのポート帯を用途別に分類（データ層: デフォルト、アプリ: 80xx、開発ツール: 81xx）しており分かりやすい

### 1.2 アーキテクチャ上の懸念

#### Write 側への同期依存（重要度: 高）

`membership` コンシューマーが `write-server` の HTTP API (`/check-room-existence`) を直接呼び出している。

```
compose.yml:179 → ROOM_SERVER: "http://write-server:8081"
internal/infrastructure/query/room.go → write-server に HTTP リクエスト
```

**問題**: Read 側のコンシューマーが Write 側サーバーに同期依存しているため、write-server がダウンするとメンバーシップ作成が全て失敗する。CQRS の分離原則に反している。

**改善案**: メンバーシップコンシューマーが TiDB に直接クエリするか、ローカルのリードモデルで存在確認を行う。

#### イベントの二重 JSON ラッピング（重要度: 中）

`internal/consumer/read_model.go:102-107` および `internal/consumer/membership.go:34-43`:

```go
var es string
err := json.Unmarshal(data.Payload, &es)  // Payload を一旦 string に
eb := []byte(es)                           // string をまた []byte に
// eb を再度 Unmarshal
```

ドメインイベントが JSON 文字列として二重にエスケープされて格納されている。これはデバッグを困難にし、パフォーマンスも低下させる。イベントの Payload を `json.RawMessage` のまま直接格納することを検討すべき。

#### イベントバージョニングの不在（重要度: 中）

`internal/events/events.go` の `EventEnvelope` に `version` フィールドがない。ドメインモデルの変更に伴いイベントスキーマが変わった場合、過去のイベントをデシリアライズできなくなる。

```go
// 現状
type EventEnvelope struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

// 推奨
type EventEnvelope struct {
    Type    string          `json:"type"`
    Version int             `json:"version"`
    Payload json.RawMessage `json:"payload"`
}
```

#### Saga / 補償トランザクションの不在（重要度: 低）

ルーム作成後のメンバーシップ作成が失敗した場合の補償処理がない。学習プロジェクトとしては許容範囲だが、プロダクションでは Saga パターンの導入が必要。

---

## 2. コードレビュー

### 2.1 バグ

#### `create_room.go:39` — `new()` と `util.ToPtr()` の不一致

```go
// internal/service/create_room.go:39
return new(r.ID()), nil

// internal/service/create_user.go:32（比較）
return util.ToPtr(u.ID()), err
```

同じパターンのコードで `new()` と `util.ToPtr()` が混在している。`util.ToPtr()` に統一すべき。

#### `httpclient.go:51` — エラーメッセージの不完全

```go
return zero, fmt.Errorf("failed to call server, status code: %d, message: ", res.StatusCode)
```

`message: ` の後にレスポンスボディの内容が含まれていない。レスポンスボディを読み取ってエラーメッセージに含めるべき。

#### `create_user.go:41` — "errror" のタイプミス

```go
return c.JSON(http.StatusInternalServerError,
    map[string]string{"error": fmt.Sprintf("internal server errror: %s", err.Error())})
//                                                          ^^^^^^ r が3つ
```

### 2.2 エラーハンドリングの問題

#### トランザクション Rollback エラーの無視（重要度: 高）

3つのリポジトリ全てで同じパターンが使われている:

```go
// internal/infrastructure/repository/user_repository.go:37
// internal/infrastructure/repository/chatroom_repository.go:36
// internal/infrastructure/repository/membership_repository.go:54
defer func() { _ = tx.Rollback() }()
```

`tx.Commit()` の後に呼ばれる `Rollback()` は no-op になるため通常は問題ないが、Commit 前にエラーで抜けた場合の Rollback 失敗が完全に無視される。最低限ログ出力すべき。

#### Kafka コンシューマーのエラー後の処理（重要度: 高）

`cmd/read/main.go` と `cmd/membership/main.go` の Kafka 消費ループ:

```go
if err != nil {
    slog.ErrorContext(ctx, "failed to consume event", "err", err)
    // → ログだけ出してスキップ。リトライ・DLQ・サーキットブレーカーなし
}
```

メッセージ処理失敗時にログ出力のみでスキップしている。at-least-once セマンティクスなのに、失敗したメッセージは永久に失われる。

#### `env.go:34` — panic によるエラーハンドリング

```go
func (e *env) Value() string {
    value, err := e.SafeValue()
    if err != nil {
        panic(fmt.Sprintf("failed to read environment: %s", err.Error()))
    }
    return value
}
```

環境変数未設定時に panic する。起動時に必須環境変数をまとめてバリデーションし、不足分を明示的にエラー出力してから終了するほうが運用しやすい。

### 2.3 重複イベント処理の TODO 未解決（重要度: 高）

```go
// internal/consumer/read_model.go:98
// internal/consumer/membership.go:24
// TODO: kafka は at-least-once な保証スタイルなので、重複したイベントが飛んできたときの処理を検討すべき
```

CQRS + イベントソーシングの核心的な課題が未解決のまま。DynamoDB への `PutItem` は上書きされるのでリードモデルは問題ないが、メンバーシップの重複作成は実害がある。以下のいずれかで対処すべき:

- 処理済みイベント ID の記録（冪等性テーブル）
- DB 側のユニーク制約 + UPSERT 操作
- コンシューマーでのイベント ID ベースの重複チェック

### 2.4 命名規則の不一致

#### リポジトリインターフェースのメソッド名

```go
// internal/domain/user/repository.go
FindByID(ctx context.Context, id string) (*User, error)  // "ID" (大文字)

// internal/domain/membership/repository.go
FindById(ctx context.Context, id string) (*Membership, error)  // "Id" (小文字d)
```

Go の慣例では `ID` が正しい（[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments#initialisms)）。`FindByID` に統一すべき。

#### ドメインモデルのフィールド公開範囲の不一致

```go
// internal/domain/user/user.go — フィールド非公開、ゲッターあり
type User struct {
    id     string
    name   string
    events []events.Event
}

// internal/domain/membership/membership.go — フィールド公開、JSON タグ付き
type Membership struct {
    Id         string         `json:"id"`
    ChatRoomId string         `json:"chat_room_id"`
    UserId     string         `json:"user_id"`
    Events     []events.Event `json:"-"`
}
```

`User` と `Room` はフィールドが非公開でゲッター経由のアクセスを強制しているが、`Membership` はフィールドが公開されている。カプセル化のアプローチが一貫していない。

#### ID プレフィックスの形式不一致

```go
// user.go:15
id := "user:" + ulid.Make().String()        // コロン区切り

// room.go:19
id := "chat-room:" + ulid.Make().String()   // コロン区切り

// membership.go:16
id := "membership-" + ulid.Make().String()  // ハイフン区切り
```

`membership` だけ区切り文字がハイフン(`-`)になっている。コロン(`:`)に統一すべき。

### 2.5 セキュリティ

#### 入力バリデーションの不足

ハンドラー層でリクエストボディのバインドのみ行い、フィールドレベルのバリデーション（空文字チェック、長さ制限など）を行っていない。ドメイン層で空文字は弾いているが、長さ制限やインジェクション対策はない。

```go
// internal/handler/create_user.go — name の長さチェックなし
// internal/handler/create_chat_room.go — name, creator_id のチェックなし
// internal/handler/check_room_existence.go — room_id のチェックなし
```

#### CORS / レート制限なし

`cmd/write-server/main.go` に CORS ミドルウェアもレート制限も設定されていない。

#### DB 接続プールの未設定

```go
// cmd/write-server/main.go:33
db, err := sql.Open("mysql", dsn)
// SetMaxOpenConns, SetMaxIdleConns, SetConnMaxLifetime が未設定
```

デフォルト値（無制限）のままだと、負荷時に接続が枯渇する可能性がある。

### 2.6 運用上の問題

#### グレースフルシャットダウンの未実装

全ての `main.go` で `signal.NotifyContext()` を使ったシグナルハンドリングがない。`SIGTERM` 受信時にリソースのクリーンアップなしでプロセスが強制終了する。

```go
// 推奨パターン
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()
```

#### ヘルスチェックエンドポイントの不在

write-server に `/health` や `/readiness` エンドポイントがない。Kubernetes や Docker ヘルスチェックでの監視ができない。

#### ロギングの不一致

```go
// cmd/write-server/main.go — log.Printf（標準ライブラリ）
log.Printf("[ERROR] /create-user: failed to bind request: %v", err)

// cmd/read/main.go — slog（構造化ログ）
slog.ErrorContext(ctx, "failed to consume event", "err", err)
```

`slog` で統一すべき。

### 2.7 HTTP API 設計

#### RESTful でないエンドポイント設計

```go
e.POST("/create-user", ...)            // → POST /users
e.POST("/create-chat-room", ...)       // → POST /rooms
e.POST("/check-room-existence", ...)   // → GET /rooms/:id
```

RPC スタイルの命名になっている。REST のリソース指向に寄せたほうが一般的だが、学習プロジェクトなのでコマンド（CQRS の C）を明示する意図であれば許容範囲。

#### エラーレスポンスの不統一

```go
// create_user.go:41 — エラー詳細をクライアントに返却
fmt.Sprintf("internal server errror: %s", err.Error())

// create_chat_room.go:48 — 同様にエラー詳細を返却
fmt.Sprintf("internal server error: %s", err.Error())
```

内部エラーの詳細をクライアントに返すのはセキュリティ上好ましくない。クライアントにはジェネリックなメッセージを返し、詳細はサーバーログに記録すべき。

### 2.8 その他のコード品質

#### `httpclient.go` — コンテキスト未伝播

```go
// pkg/httpclient/httpclient.go:33
req, err := http.NewRequest(http.MethodPost, client.baseURL+path, bytes.NewReader(reqBodyBytes))
```

`http.NewRequestWithContext` を使って `context.Context` を伝播すべき。現状、呼び出し元のキャンセルやタイムアウトが HTTP リクエストに伝わらない。

#### `DynamoDB` テーブル名のハードコード

```go
// internal/consumer/read_model.go
TableName: new("Users")       // L60
TableName: new("Rooms")       // L75
TableName: new("Memberships") // L90

// cmd/read/initialize_tables.go
TableName: new("Users")       // L36
TableName: new("Rooms")       // L50
TableName: new("Memberships") // L64
```

テーブル名が2箇所に分散してハードコードされている。定数として一元管理すべき。

#### `create_chat_room.go:31-36` — recover による panic キャッチ

```go
func (h *CreateRoomHandler) Handle(c echo.Context) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = c.JSON(http.StatusInternalServerError,
                map[string]string{"error": fmt.Sprintf("internal server error: %s", r)})
        }
    }()
```

Echo の `middleware.Recover()` が既に設定されているため、ハンドラー内での recover は冗長。また、`create_user.go` にはこの recover がなく、一貫性がない。

---

## 3. テストカバレッジ

### 現状のテスト

| パッケージ | テストあり | カバー範囲 |
|-----------|:--------:|---------|
| `internal/domain/user` | Yes | NewUser, ReconstructUser, イベント検証 |
| `internal/domain/room` | Yes | NewRoom, ReconstructRoom, イベント検証 |
| `internal/domain/membership` | No | - |
| `internal/service` | No | - |
| `internal/handler` | No | - |
| `internal/infrastructure/repository` | No | - |
| `internal/consumer` | No | - |
| `pkg/httpclient` | No | - |
| `pkg/env` | No | - |
| `cmd/membership` | Yes | プレースホルダーのみ（テストケースなし） |

### 不足しているテスト

- **Membership ドメインモデル**: `CreateMembership`, `ReconstructMembership` のテストがない
- **サービス層**: 全サービスにユニットテストがない。リポジトリのモック化によるテストが必要
- **ハンドラー層**: HTTP リクエスト/レスポンスのテストがない
- **コンシューマー**: イベント処理のテストがない
- **インフラ層**: リポジトリの統合テストがない

---

## 4. インフラ / CI 構成

### Docker Compose

- **良い点**: ヘルスチェックと depends_on の条件指定が適切、リソース制限の設定あり
- **懸念**: `dynamodb` サービスに `latest` タグを使用している。バージョンを固定すべき

### Dockerfile

- **良い点**: マルチステージビルド、非 root ユーザー、CGO_ENABLED=0 の静的ビルド
- **懸念**: イメージタグにダイジェストハッシュがなく、再現性が完全ではない

### CI

- 基本的な Go テストの実行は構成されている
- リント（golangci-lint）の設定ファイルは存在するが、CI で実行されているか要確認

---

## 5. 改善提案の優先度

### 即座に対応すべき（バグ修正）

1. `create_room.go:39` の `new(r.ID())` を `util.ToPtr(r.ID())` に修正（一貫性）
2. `httpclient.go:51` のエラーメッセージにレスポンスボディを含める
3. `create_user.go:41` の "errror" タイプミスを修正

### 短期的に対応すべき（信頼性向上）

4. Kafka コンシューマーの重複イベント処理を実装
5. コンシューマーのエラーハンドリング改善（リトライ / DLQ）
6. グレースフルシャットダウンの実装
7. `httpclient.go` で `http.NewRequestWithContext` を使用
8. ロギングを `slog` に統一

### 中期的に対応すべき（品質向上）

9. 命名規則の統一（`FindByID`, ID プレフィックス区切り文字, Membership のフィールド公開範囲）
10. テストカバレッジの向上（特にサービス層とコンシューマー）
11. イベントバージョニングの導入
12. ヘルスチェックエンドポイントの追加
13. DB 接続プールの設定
14. 入力バリデーションの強化

### 長期的に検討すべき（アーキテクチャ改善）

15. membership コンシューマーの write-server 依存の解消
16. イベントの二重 JSON ラッピングの解消
17. Saga パターンの導入
18. エラーレスポンス形式の統一と内部エラー詳細の隠蔽

---

## 6. 総括

学習プロジェクトとして、CQRS + イベントソーシング + DDD の概念を実際のコードに落とし込む点でよくできている。特に以下の点は評価できる:

- ドメイン層がインフラ層に依存しないクリーンな依存関係
- イベントの発行と消費の仕組みが一通り動作する構成
- Docker Compose による開発環境の再現性

一方で、プロダクション品質に近づけるためには、エラーハンドリングの体系的な改善、重複イベントへの対応、テストカバレッジの向上が最も重要な課題となる。
