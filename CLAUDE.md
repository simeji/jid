# jid — JMESPath 対応作業メモ

## プロジェクト概要

**jid (JSON Incremental Digger)** はターミナルで JSON をインタラクティブに掘り下げるツール。
リポジトリ: `github.com/simeji/jid`

## 作業ブランチ

```
claude/add-jmespath-support-hSsUN
```

## 完了した作業（コミット済み・プッシュ済み）

### 追加した機能

1. **JMESPath 式の評価** (`json_manager.go`)
   - `github.com/jmespath/go-jmespath v0.4.0` を導入
   - パイプ (` | `)、ワイルドカード (`[*]`)、フィルタ (`[?`)、関数呼び出し、`@` 参照などを含むクエリを自動検出して JMESPath エンジンで評価
   - 従来の `.field[0]` 形式は既存パスをそのまま使用（後方互換）

2. **関数入力中のプレビュー維持** (`json_manager.go`)
   - `.users | len` のように関数名を途中まで入力している間 (`(` が未入力) は `isFunctionTypingMode` で検出
   - パイプ前の式の結果をプレビューに表示し続ける（空白や null にならない）
   - 入力済み文字に一致する関数候補を候補リストに提示

3. **関数サジェスト** (`suggestion.go`)
   - `SuggestionInterface` に `GetFunctionCandidates(prefix)` と `GetFunctionSuggestion(prefix)` を追加
   - 25 個の JMESPath 組み込み関数 (`abs`, `avg`, `ceil`, `contains`, `ends_with`, `floor`, `join`, `keys`, `length`, `map`, `max`, `max_by`, `merge`, `min`, `min_by`, `not_null`, `reverse`, `sort`, `sort_by`, `starts_with`, `sum`, `to_array`, `to_number`, `to_string`, `type`, `values`) を管理
   - 候補は `length(` のように末尾に `(` が付いた形で返される

4. **候補確定の改善** (`engine.go`)
   - `confirmCandidate()`: 関数候補 (`(` で終わる) は ` | funcname(@)` 形式で挿入し、カーソルを `)` の直前に配置
   - `tabAction()`: 関数補完時は先頭の `.` を付けないよう分岐

5. **クエリバリデーションの緩和** (`query.go`)
   - `validate()` に JMESPath のパイプ・ワイルドカード・フィルタを通す例外を追加

## 主要ファイル構成

```
engine.go          # メインループ・キー操作・候補確定ロジック
json_manager.go    # JSON フィルタリング（JMESPath / Legacy 二系統）
query.go           # クエリ文字列のパース・バリデーション
suggestion.go      # フィールド補完 + JMESPath 関数補完
terminal.go        # TUI レンダリング（変更なし）
cmd/jid/jid.go    # エントリポイント（変更なし）
```

## 使い方（実装後）

```
.                    → ルート JSON を表示
.users               → users フィールドへ絞り込み
.users[0].name       → 配列インデックス + フィールドアクセス（従来通り）
.users[*].name       → JMESPath ワイルドカード：全ユーザーの name を抽出
.users | <TAB>       → 全関数候補を表示しながら users の内容をプレビュー
.users | len<TAB>    → length( などにフィルタリング → 確定で .users | length(@) に
. | keys(@)          → ルートのキー一覧を取得
.users | sort_by(@, &name)  → name でソート
```

## ビルド・テスト

```bash
go build ./...
go test ./...
```

## 残課題・改善余地

- `sort_by(@, &field)` のような引数付き関数を確定したとき、カーソル位置をさらに使いやすくする
- `[?age > `20`]` のようなフィルタ式入力中のサジェストは未対応
- `deleteWordBackward` (Ctrl+W) が JMESPath 式の ` | ` 単位で消せるとより使いやすい
- `validate()` の JMESPath 用パスをより厳密にする（現状はパイプがあれば素通し）
