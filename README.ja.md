# gem-cli

Vertex AI 経由で Google Gemini を操作する CLI クライアント。マルチモーダル入力（画像・PDF・音声・動画）、ストリーミング、バッチ処理、Google Search Grounding、構造化出力、プロンプトインジェクション防御に対応。

[lite-llm](https://github.com/nlink-jp/lite-llm)（OpenAI 互換）の Gemini ネイティブ版として設計。Gemini 固有機能にフルアクセス可能。

[English README is here](README.md)

## 特徴

- **マルチモーダル入力** — テキストに加えて画像・PDF・音声・動画ファイルを添付可能
- **ストリーミング** — `--stream` でトークン単位の出力
- **Google Search Grounding** — `--grounding` で Web 検索付き生成
- **構造化出力** — `--format json` / `--json-schema` で Gemini ネイティブの構造化出力
- **データ隔離** — stdin/ファイル入力を自動的にノンス付き XML でラップし、プロンプトインジェクションを防御
- **バッチモード** — 入力を1行ずつ処理、1行1リクエスト
- **パイプ対応** — stdin から読み、stdout に書き出す Unix 的設計
- **quiet モード** — `--quiet` / `-q` で警告を抑制（パイプライン向け）

## インストール

```sh
git clone https://github.com/nlink-jp/gem-cli.git
cd gem-cli
make build
# バイナリ: dist/gem-cli
```

または [リリースページ](https://github.com/nlink-jp/gem-cli/releases) からビルド済みバイナリをダウンロード。

## クイックスタート

```sh
# Google Cloud 認証
gcloud auth application-default login
export GOOGLE_CLOUD_PROJECT="your-project-id"

# 質問する
gem-cli "日本の首都は？"

# パイプでデータを渡す（自動的にデータ隔離）
echo "2026-03-29: 売上 ¥4,520,000" | gem-cli "日付と金額を JSON で抽出して" --format json

# 画像を分析
gem-cli "この写真に何が写っている？" --image photo.jpg

# Web 検索 Grounding
gem-cli --grounding "Log4j の最新 CVE は？"

# バッチ処理
cat questions.txt | gem-cli --batch --format jsonl \
  --system-prompt "1文で回答してください。"

# ストリーミング
gem-cli --stream "サイバーセキュリティについて俳句を書いて"
```

## 設定

設定例ファイルをコピー:

```sh
mkdir -p ~/.config/gem-cli
cp config.example.toml ~/.config/gem-cli/config.toml
```

```toml
# ~/.config/gem-cli/config.toml
[gcp]
project  = "your-project-id"
location = "us-central1"

[model]
name = "gemini-2.5-flash"
```

**優先順位（高い順）:** CLI フラグ → 環境変数 → 設定ファイル → デフォルト値

| 環境変数 | 説明 |
|---|---|
| `GOOGLE_CLOUD_PROJECT` | GCP プロジェクト ID（必須） |
| `GOOGLE_CLOUD_LOCATION` | Vertex AI リージョン（デフォルト: us-central1） |
| `GEM_CLI_MODEL` | モデル名（デフォルト: gemini-2.5-flash） |

## 使い方

```
gem-cli [flags] [prompt]

入力フラグ:
  -p, --prompt string              ユーザープロンプトテキスト
  -f, --file string                入力ファイルパス（テキストとして読み込み、- で stdin）
  -s, --system-prompt string       システムプロンプトテキスト
  -S, --system-prompt-file string  システムプロンプトファイルパス

マルチモーダル:
      --image strings              画像ファイルパス（複数指定可）
      --attach strings             添付ファイル: PDF, 音声, 動画（複数指定可）

モデル:
  -m, --model string               モデル名（設定を上書き）

実行モード:
      --stream                     ストリーミング出力を有効化
      --batch                      バッチモード: 入力1行ごとに1リクエスト
      --grounding                  Google Search Grounding を有効化

出力形式:
      --format string              出力形式: text（デフォルト）, json, jsonl
      --json-schema string         JSON Schema ファイル（--format json を含意）

セキュリティ:
      --no-safe-input              自動データ隔離を無効化
  -q, --quiet                      stderr の警告を抑制
      --debug                      API リクエスト詳細を stderr に出力

設定:
  -c, --config string              設定ファイルパス
```

## マルチモーダル入力

### 画像

```sh
# 1枚の画像
gem-cli "この画像を詳しく説明してください" --image photo.jpg

# 複数の画像
gem-cli "この2つのスクリーンショットの違いは？" \
  --image before.png --image after.png
```

対応形式: JPEG, PNG, GIF, WebP

### ドキュメント・音声・動画

```sh
# PDF 分析
gem-cli "主要な調査結果を要約して" --attach report.pdf

# 音声文字起こし
gem-cli "この音声を文字起こしして" --attach recording.mp3

# 動画分析
gem-cli "この動画で何が起きているか説明して" --attach clip.mp4

# テキスト + マルチモーダルの組み合わせ
gem-cli "この図は仕様書と整合性がありますか？" \
  --image architecture.png --attach spec.pdf
```

対応形式: PDF, MP3, WAV, MP4, MOV, TXT, CSV, Markdown

## データ隔離

stdin やファイル（`-f`）からの入力は自動的にランダムタグ付き XML でラップされます:

```
<user_data_a3f8b2>
{あなたのデータ}
</user_data_a3f8b2>
```

**システムプロンプト**内で `{{DATA_TAG}}` を使うとタグ名を参照できます:

```sh
echo "Alice, 34, engineer" | gem-cli \
  --system-prompt "<{{DATA_TAG}}> からフィールドを抽出して JSON で返して。キー: name, age, role" \
  --format json
```

> `{{DATA_TAG}}` は**システムプロンプト内でのみ**展開され、ユーザー入力では展開されません。

信頼できる入力の場合は `--no-safe-input` で無効化。

## 構造化出力

```sh
# JSON オブジェクト
gem-cli --format json "サイバーセキュリティのベストプラクティスを3つ挙げて"

# JSON Schema — 厳密な構造指定
gem-cli --json-schema person.json "架空のセキュリティアナリストを生成して"

# バッチ + JSONL
cat items.txt | gem-cli --batch --format jsonl \
  --system-prompt "food, vehicle, animal, other のいずれかに分類して。"
```

### JSON Schema の例

```json
{
  "type": "OBJECT",
  "properties": {
    "name": {"type": "STRING"},
    "age": {"type": "INTEGER"},
    "occupation": {"type": "STRING"}
  },
  "required": ["name", "age", "occupation"]
}
```

> Gemini は独自のスキーマ形式を使用します（`"type": "string"` ではなく `"type": "STRING"`）。

## Google Search Grounding

```sh
# Web 検索で回答を補強
gem-cli --grounding "Go の最新バージョンは？"

# 構造化出力と組み合わせ
gem-cli --grounding --format json "現在のビットコイン価格（USD と JPY）"
```

Grounding は Vertex AI 経由の Google Search を使用します。プロジェクトで Vertex AI API が有効になっている必要があります。

## ビルド

```sh
make build      # 現在のプラットフォーム → dist/gem-cli
make build-all  # 全5プラットフォーム → dist/
make test       # テスト実行（6パッケージ31テスト）
make check      # vet + test + build
```

> **注（サンドボックス環境）:** デフォルトの Go キャッシュパスに書き込めない場合:
>
> ```sh
> GOCACHE=/tmp/go-cache GOMODCACHE=/tmp/gopath/pkg/mod make build
> ```
