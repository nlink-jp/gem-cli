# gem-cli

Vertex AI 経由で Google Gemini を操作する CLI クライアント。マルチモーダル入力（画像・PDF・音声・動画）、ストリーミング、Google Search Grounding、構造化出力、プロンプトインジェクション防御に対応。

[English README is here](README.md)

## 特徴

- **マルチモーダル入力** — テキストに加えて画像・PDF・音声・動画ファイルを添付可能
- **ストリーミング** — `--stream` でトークン単位の出力
- **Google Search Grounding** — `--grounding` で Web 検索付き生成
- **構造化出力** — `--format json` / `--json-schema` で Gemini ネイティブの構造化出力
- **データ隔離** — stdin/ファイル入力を自動的にノンス付き XML でラップし、プロンプトインジェクションを防御
- **バッチモード** — 入力を1行ずつ処理（実装予定）
- **パイプ対応** — stdin から読み、stdout に書き出す Unix 的設計

## インストール

```sh
git clone https://github.com/nlink-jp/gem-cli.git
cd gem-cli
make build
# バイナリ: dist/gem-cli
```

## 設定

```sh
gcloud auth application-default login
export GOOGLE_CLOUD_PROJECT="your-project-id"
```

設定ファイル（任意）: `~/.config/gem-cli/config.toml`

```toml
[gcp]
project  = "your-project-id"
location = "us-central1"

[model]
name = "gemini-2.5-flash"
```

| 環境変数 | 説明 |
|---|---|
| `GOOGLE_CLOUD_PROJECT` | GCP プロジェクト ID（必須） |
| `GOOGLE_CLOUD_LOCATION` | Vertex AI リージョン（デフォルト: us-central1） |
| `GEM_CLI_MODEL` | モデル名（デフォルト: gemini-2.5-flash） |

## 使い方

```sh
# テキストプロンプト
gem-cli "日本の首都は？"

# 画像分析
gem-cli "この画像を説明してください" --image photo.jpg

# PDF 分析
gem-cli "この文書を要約してください" --file report.pdf

# Google Search Grounding
gem-cli --grounding "Log4j の最新 CVE"

# ストリーミング
gem-cli --stream "セキュリティについて俳句を書いて"

# JSON 出力
gem-cli --format json "Go のベストプラクティスを3つ挙げて"

# パイプ（データ隔離あり）
echo "Alice, 34, engineer" | gem-cli \
  -s "<{{DATA_TAG}}> からフィールドを抽出して JSON で返して" \
  --format json
```

## データ隔離

stdin やファイルからの入力は自動的にランダムタグ付き XML でラップされます:

```
<user_data_a3f8b2>
{あなたのデータ}
</user_data_a3f8b2>
```

システムプロンプト内で `{{DATA_TAG}}` を使うとタグ名を参照できます。`--no-safe-input` で無効化。

## ビルド

```sh
make build      # 現在のプラットフォーム → dist/gem-cli
make build-all  # 全5プラットフォーム → dist/
make test       # テスト実行
make check      # vet + test + build
```
