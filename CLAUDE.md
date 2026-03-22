# jid — JMESPath 対応作業メモ

## プロジェクト概要

**jid (JSON Incremental Digger)** はターミナルで JSON をインタラクティブに掘り下げるツール。
リポジトリ: `github.com/simeji/jid`

## 作業ブランチ

```
claude/add-jmespath-support-hSsUN
```

## 完了した作業

### 追加した機能

1. **JMESPath 式の評価** (`json_manager.go`)
   - `github.com/jmespath/go-jmespath v0.4.0` を導入
   - パイプ (` | `)、ワイルドカード (`[*]`)、フィルタ (`[?`)、関数呼び出し、`@` 参照などを含むクエリを自動検出して JMESPath エンジンで評価
   - 従来の `.field[0]` 形式は既存パスをそのまま使用（後方互換）

2. **関数入力中のプレビュー維持** (`json_manager.go`)
   - `.users | len` のように関数名を途中まで入力している間 (`(` が未入力) は `isFunctionTypingMode` で検出
   - パイプ前の式の結果をプレビューに表示し続ける
   - 入力済み文字に一致する関数候補を候補リストに提示

3. **関数サジェスト** (`suggestion.go`)
   - `GetFunctionCandidates(prefix)` / `GetFunctionSuggestion(prefix)` を追加
   - `GetFunctionCandidatesFiltered(prefix, SuggestionDataType)` / `GetFunctionSuggestionFiltered(prefix, SuggestionDataType)` を追加
   - 25 個の JMESPath 組み込み関数を管理。候補は `length(` のように末尾に `(` が付いた形で返される

4. **パイプ前の型による関数絞り込み** (`suggestion.go`, `json_manager.go`)
   - `jmespathFunctionsByType` マップ: ARRAY / MAP / STRING / NUMBER / BOOL ごとに有効な関数を定義
   - `GetCurrentType(json)` でパイプ前の値の型を判定し、型に合った関数のみ表示
   - 例: `.game_indices | ` → ARRAY 型対象の関数のみ表示

5. **候補確定の改善** (`engine.go`)
   - `confirmCandidate()`:
     - 関数候補 (`(` で終わる): ` | funcname(args)` 形式で挿入、カーソルを適切な位置に配置
     - ワイルドカード投影 (`[*]` 末尾): `.fieldname` を追加
     - パイプ後の完成済み式 (suffix に `(` が含まれる): フルクエリに `.fieldname` を追加
     - 単純パイプフィールド: `base.field` に変換
   - `tabAction()`: `[*` 末尾で Tab → `[*]` に閉じる; `[N]` で配列インクリメント; `[0]` 単一要素配列対応

6. **クエリバリデーションの緩和** (`query.go`)
   - `validate()` に JMESPath のパイプ・ワイルドカード・フィルタを通す例外を追加

7. **関数説明の表示と Ctrl+X トグル** (`terminal.go`, `engine.go`)
   - 関数候補選択中に候補一覧の下行に使い方を黄色で表示
   - デフォルトON、Ctrl+X でトグル
   - `FunctionDescription(name)` で説明文を取得
   - `jmespathFunctionDescriptions` マップに全関数の説明を定義

8. **関数引数テンプレートとプレースホルダ** (`suggestion.go`, `engine.go`, `terminal.go`)
   - `FunctionTemplate(name)` → args / cursorBack / placeholderLen を返す
   - `contains(@, '')` → カーソルが `''` の内側に; `sort_by(@, &field)` → `field` がプレースホルダとして薄青で表示
   - プレースホルダ文字位置で任意のキーを打つと置き換え

9. **Shift+Tab で候補を後退** (`engine.go`)
   - `\x1b[Z` の3イベント分割問題を対処: KeyEsc 時に候補状態を保存し、Z 到着時に復元してから `shiftTabAction()`
   - 候補モードでは前の候補へ; 候補モード外では配列インデックスをデクリメント

10. **Ctrl+W の JMESPath 対応** (`engine.go`)
    - `deleteWordBackward()`:
      - パイプ後の suffix を `removeLastJMESPathSegment()` で1セグメントずつ削除
      - `.id` → `[0]` → `func(@)` の順で削除
      - suffix が空になったら `base | ` を残し、次の Ctrl+W でパイプを削除
    - `removeLastJMESPathSegment(expr string)`: 末尾から `.`/`[`/`(` の深さを追いながら最後のセグメント区切りを検出

11. **ワイルドカード投影後のフィールドサジェスト** (`json_manager.go`, `engine.go`)
    - `.game_indices[*]` → 配列要素オブジェクトのフィールドキーを候補に表示
    - `.game_indices[*].` (末尾ドット) → 式を末尾ドットなしで再評価してフィールド候補を表示
    - `.game_indices[*].v` (部分フィールド名) → JMESPath が空配列を返す場合、`reWildcardFieldTyping` で検出して一致するフィールド候補を表示
      - 一致なしの場合は実際の空配列結果をそのまま表示
    - `.game_indices[*].version` (配列of objects) → `[0]` ではなく `name`/`url` フィールドを候補に表示
    - `setCandidateData` / `tabAction` で `strings.Contains(qs, "[*]")` を判定に使用

12. **パイプ後の完成済み式のフィールドサジェスト** (`json_manager.go`)
    - `.[1] | to_array(@)[0]` → 評価結果がオブジェクトなら `body`, `id` 等を候補に表示
    - `.[1] | to_array(@)[0].` → パイプ error path でsuffix末尾ドット検出 → 式を再評価してフィールド候補

13. **`.[3]` 後の文字入力ドリフト修正** (`engine.go`)
    - `validate(".[3]a")` が false → `Insert` が no-op → `queryCursorIdx++` だけ進む問題
    - `inputChar` で Insert 前後の `query.Length()` を比較し、実際に挿入された場合のみインクリメント

## 主要ファイル構成

```
engine.go          # メインループ・キー操作・候補確定ロジック
json_manager.go    # JSON フィルタリング（JMESPath / Legacy 二系統）
query.go           # クエリ文字列のパース・バリデーション
suggestion.go      # フィールド補完 + JMESPath 関数補完
terminal.go        # TUI レンダリング
cmd/jid/jid.go    # エントリポイント（変更なし）
```

## テストファイル

```
suggestion_test.go     # GetFunctionCandidates / GetFunctionCandidatesFiltered / FunctionDescription / FunctionTemplate
json_manager_test.go   # GetFilteredDataJMESPath* (wildcard / type-filter / object-result)
engine_test.go         # removeLastJMESPathSegment / DeleteWordBackwardJMESPath / ConfirmCandidateJMESPath / TabActionJMESPath
```

## 使い方（実装後）

```
.                              → ルート JSON を表示
.users                         → users フィールドへ絞り込み
.users[0].name                 → 配列インデックス + フィールドアクセス（従来通り）
.users[*].name                 → JMESPath ワイルドカード：全ユーザーの name を抽出
.users[*].<Tab>                → 要素オブジェクトのフィールド候補を表示
.game_indices[*].version.<Tab> → ネストしたオブジェクト投影のフィールド候補
.users | <Tab>                 → 型に応じた関数候補を表示（ARRAY なら avg, sort 等）
.users | len<Tab>              → length( に絞り込み → 確定で .users | length(@)
. | keys(@)                    → ルートのキー一覧を取得
.users | sort_by(@, &name)     → name でソート
.[1] | to_array(@)[0].title    → パイプ連鎖 + インデックス + フィールド
```

## ビルド・テスト

```bash
go build ./...
go test ./...
go build -o /tmp/jid ./cmd/jid/   # 動作確認用バイナリ
```

## キー操作（追加分）

| キー | 動作 |
|------|------|
| `Tab` | `[*` → `[*]` 閉じ / 候補選択 / フィールド・関数補完 |
| `Shift+Tab` | 候補を前に戻る / 配列インデックスをデクリメント |
| `Ctrl+W` | JMESPath パイプ対応の単語削除（セグメント単位） |
| `Ctrl+X` | 関数説明表示のトグル（関数候補表示中のみ） |

## 残課題・改善余地

- `[?age > '20']` のようなフィルタ式入力中のサジェストは未対応
- `validate()` の JMESPath 用パスをより厳密にする（現状はパイプがあれば素通し）
- `.[*]` と `.*` の混在ケースの網羅的テスト
