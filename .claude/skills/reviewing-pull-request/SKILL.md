---
name: reviewing-pull-request
description: Use when asked to review a PR, give feedback on code changes, or post review comments to GitHub in this repository
allowed_tools:
  - Bash(gh pr view:*)
  - Bash(gh repo view:*)
  - Bash(gh pr list:*)
  - Bash(gh api:*)
  - Bash(git diff:*)
  - Bash(git log:*)
  - Bash(git show:*)
  - Bash(git branch:*)
---

# PR レビュースキル

## Overview

コードレビューの結果を GitHub に **インラインコメント + APPROVE/REQUEST_CHANGES** として投稿する手順。

## 手順

### Step 1: PR番号とリポジトリ情報を取得

```bash
# 現在のブランチのPR番号を取得
gh pr view --json number,headRefName,baseRefName

# オーナー/リポジトリ名を取得
gh repo view --json owner,name
```

### Step 2: 既存のレビューコメントと議論を確認する

```bash
# 過去のレビュー（総評）を確認
gh api repos/{owner}/{repo}/pulls/{PR番号}/reviews

# 過去のインラインコメントを確認（position: null のものは解決済みまたは outdated）
gh api repos/{owner}/{repo}/pulls/{PR番号}/comments

# PR の一般コメント（issue comments）を確認 — 開発者の意見・反論・補足が含まれる
gh api repos/{owner}/{repo}/issues/{PR番号}/comments
```

**返信スレッドの確認:**
- インラインコメント（`/pulls/{PR番号}/comments`）の `in_reply_to_id` フィールドに注目する
- `in_reply_to_id` が設定されているコメントは、元のコメントへの返信である
- スレッドの文脈（元の指摘 → 開発者の返信 → 再返信）を把握した上でレビューする

**重複指摘のルール:**
- `position` が値を持つコメント（まだ有効なスレッド）→ **同じ内容は指摘しない**
- `position: null` のコメント（コードが変更されて outdated になったもの）→ 問題が再発していれば**再度指摘してよい**

**過去の議論を尊重するルール:**
- 過去のレビューで開発者が反論・説明している場合、その内容を理解した上でレビューする
- 以前の指摘に対して開発者が納得できる反論をしている場合、同じ指摘を繰り返さない
- 議論が未解決のスレッドがある場合、その議論を踏まえて判断する

### Step 3: コードを読んでレビュー内容を決める

- `git diff main..HEAD` で変更差分を確認
- 問題を見つけたらファイルパスと行番号を記録しておく

### Step 4: `gh api` でレビューを投稿

レビューは **総評（`body`）** と **インラインコメント（`comments`）** の2種類に分けて投稿する。

- **`body`（総評）**: PR全体の評価、良い点・悪い点のまとめ、マージ判定の根拠など。個別のコードに紐付かない内容はここに書く。
- **`comments`（インラインコメント）**: 特定のファイル・行に対する指摘。バグの指摘、改善提案など。インラインコメントの冒頭には必ず以下の接頭辞をつける：

| 接頭辞 | 意味 | `event` との関係 |
|--------|------|-----------------|
| `[must]` | マージ前に必ず修正が必要 | `REQUEST_CHANGES` の根拠になる |
| `[should]` | 修正を強く推奨するが、マージはブロックしない | 総評で言及する |
| `[nits]` | 軽微な提案・スタイル・好み | なくてもよい |

`event` は `[must]` が1つでもあれば `REQUEST_CHANGES`、なければ `APPROVE` または `COMMENT`。

```bash
gh api repos/{owner}/{repo}/pulls/{PR番号}/reviews \
  --method POST \
  --input - << 'EOF'
{
  "body": "## 総評\n\nCQRSの基本的な構造は正しく実装できています。\nただし、Kafkaコンシューマーのエラーハンドリングに重大なバグがあります。\n\n**良い点:**\n- ドメイン集約でイベントをカプセル化するパターンが正しい\n- コンパイル時インターフェースチェックを使っている\n\n**要修正:**\n- `MembershipConsumer` のループ制御バグ（インラインコメント参照）\n- `ReadModelConsumer` のエラー後の処理続行（インラインコメント参照）",
  "event": "REQUEST_CHANGES",
  "comments": [
    {
      "path": "internal/consumer/membership.go",
      "line": 25,
      "body": "[must] `return nil` ではなく `continue` にしてください。\n\n`return nil` だとループを抜けてしまい、残りのエントリが処理されなくなります。"
    },
    {
      "path": "internal/consumer/read_model.go",
      "line": 40,
      "body": "[must] エラー処理後に `continue` がないため、ゼロ値のまま後続処理が実行されます。\n\nエラーを検出したら `continue` でそのイベントをスキップしてください。"
    }
  ]
}
EOF
```

**`event` の選択肢:**

| 値 | 意味 |
|----|------|
| `"APPROVE"` | 問題なし、マージ可能 |
| `"REQUEST_CHANGES"` | 修正が必要（マージブロック） |
| `"COMMENT"` | 参考意見のみ（判定なし） |

### Step 5: 確認

```bash
# 投稿されたレビューを確認
gh pr view {PR番号} --comments
```

## インラインコメントのフィールド

| フィールド | 必須 | 説明 |
|-----------|------|------|
| `path` | ○ | リポジトリルートからの相対ファイルパス |
| `line` | ○ | コメントする行番号（変更後のファイル基準） |
| `body` | ○ | コメント本文（Markdown使用可） |
| `start_line` | - | 複数行にまたがる場合の開始行 |
| `side` | - | `"RIGHT"`（デフォルト、変更後）または `"LEFT"`（変更前） |

## 注意点

- `line` は **変更後のファイルの行番号**（`git diff` で `+` が付いている行、または変更されていない文脈行）
- 削除された行にコメントするには `side: "LEFT"` と削除前の行番号を使う
- `body` には Markdown が使える（コードブロック、箇条書き等）
- インラインコメントなしで全体コメントだけ投稿するなら `comments` フィールドは省略可

## トラブルシューティング

**「Unprocessable Entity」エラーが出る場合:**
- `line` が diff に含まれていない行を指している可能性がある
- `git diff main..HEAD -- {ファイルパス}` で実際の差分を確認して行番号を調整する

**PR番号が分からない場合:**
```bash
gh pr list --head $(git branch --show-current)
```
