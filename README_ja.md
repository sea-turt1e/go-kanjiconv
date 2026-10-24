# go-kanjiconv
![Go](https://img.shields.io/badge/-Go-00ADD8.svg?logo=go&style=flat)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

[EN](README.md) | **JA**

kanjiconvのGolang版です。

漢字をひらがな、カタカナ、ローマ字に変換するツールです。
本ライブラリはPython版 [kanjiconv](https://github.com/sea-turt1e/kanjiconv) のGo移植版です。Pure Goの日本語形態素解析器 [kagome](https://github.com/ikawaha/kagome)（辞書埋め込み型）を使って、日本語の文の読みや発音を取得できます。外部サービスや辞書のダウンロードは不要です。

## Python版との違い

Python版はSudachiPy/SudachiDictをベースにしています。Go版ではSudachiDictを扱えるメンテナンスされたGoバインディングが無いため、代わりに[kagome](https://github.com/ikawaha/kagome)を使用し、**デフォルトの辞書にUniDicを採用**しています。UniDicはより大規模で新しい語彙を持つため、レガシーなIPA辞書（`mecab-ipadic`、2007年更新が最後）よりも固有名詞・新語に強いためです。そのため以下の違いがあります：

- 読みはPython版とおおむね近い結果になりますが、トークンの区切り方が微妙に異なる場合があります（例：複合固有名詞が1トークンではなく2トークンに分かれるなど）。
- 同梱されているUniDicは、Python版のオプション`unidic`パッケージとは異なる配布物（`kagome-dict/uni`）です。
- より軽量でレガシーなIPA辞書も、`WithDict(ipa.Dict())`（ライブラリ）や`--dict ipa`（CLI）で引き続き利用できます。読みの精度を多少犠牲にしてバイナリサイズを抑えたい場合はこちらを使ってください（サイズの違いは[環境](#環境)を参照）。
- SudachiPyのsplit_mode（A/B/C）に相当するものはなく、代わりにkagome自身の`TokenizeMode`（Normal/Search/Extended）を`WithTokenizeMode`で指定します。

## 環境
```
Go 1.27.3（推奨）または Go 1.26.10
```

`go.mod` は `go 1.26.10` と `toolchain go1.27.3` を指定しています。同じ系列のより新しいパッチリリース（例：1.27.4、1.26.11）でも動作します。これらのバージョンが必要な理由は[セキュリティ：eMudhra CA鍵漏洩への対応](#セキュリティemudhra-ca鍵漏洩への対応)を参照してください。

同梱のUniDic辞書は`go:embed`/kagome経由でバイナリに埋め込まれ、バイナリサイズがおよそ45MB増加します（外部ダウンロードは不要です）。バイナリサイズを抑えたく、固有名詞の精度低下を許容できる場合は、代わりにレガシーなIPA辞書（ライブラリでは`WithDict(ipa.Dict())`、CLIでは`--dict ipa`）を使用してください。こちらの増加量は約11MBです。

## インストール
```bash
go get github.com/sea-turt1e/go-kanjiconv
```

CLIをインストールする場合:
```bash
go install github.com/sea-turt1e/go-kanjiconv/cmd/kanjiconv@latest
```

## 使用方法

### CLIの使用方法
```bash
kanjiconv "幽☆遊☆白書は、最高の漫画デス。"
# yuuyuu hakusho ha ,  saikou no manga desu .
```

デフォルトでは、CLIはテキストをローマ字に変換し、各トークンの読みの間に半角スペースを1つ挿入します。

`-m`/`--mode` で出力形式を選択できます:

```bash
# ひらがなに変換
kanjiconv "幽☆遊☆白書は、最高の漫画デス。" --mode hiragana
# ゆうゆう はくしょ は 、 さいこう の まんが です 。

# カタカナに変換
kanjiconv "幽☆遊☆白書は、最高の漫画デス。" --mode katakana
# ユウユウ ハクショ ハ 、 サイコウ ノ マンガ デス 。

# ローマ字に変換（デフォルトと同じ）
kanjiconv "幽☆遊☆白書は、最高の漫画デス。" --mode roman
# yuuyuu hakusho ha ,  saikou no manga desu .
```

`-s`/`--separator` でトークンの読みの間に挿入する区切り文字を変更できます:

```bash
kanjiconv "今日は良い天気です" --mode hiragana --separator "/"
# きょう/は/よい/てんき/です

kanjiconv "今日は良い天気です" --mode hiragana --separator ""
# きょうはよいてんきです
```

フラグはテキスト引数の前後どちらに置いても構いません:

```bash
kanjiconv "今日は良い天気です" --mode hiragana --separator "/"
```

その他のオプション:

```bash
# デフォルトのUniDicの代わりに、より軽量なレガシーIPA辞書を使用
kanjiconv "東京に行く" --mode hiragana --dict ipa

# カスタム読みのフォールバックを無効化
kanjiconv "激を飛ばす" --mode hiragana --no-custom-readings
```

## CLIのフラグ/オプション

| オプション                                  | 説明                                                                    |
| ------------------------------------------ | ---------------------------------------------------------------------- |
| `-m`/`--mode {roman,hiragana,katakana}`    | 変換モード。デフォルトは `roman`。                                        |
| `-s`/`--separator SEPARATOR`               | トークンの読みの間に挿入する区切り文字。デフォルトは半角スペース1つ。         |
| `--dict {uni,ipa}`                         | 使用する辞書。デフォルトは`uni`（固有名詞の精度が高い）。`ipa`は軽量だがレガシー。 |
| `--no-custom-readings`                     | カスタム読みのフォールバックを無効化する。                                  |
| `--tokenize-mode {normal,search,extended}` | kagomeのトークナイズモード。デフォルトは`normal`。                          |
| `--version`                                | インストールされているバージョンを表示する。                                |
| `-h`/`--help`                              | ヘルプを表示する。                                                       |

### インポートとインスタンスの生成
```go
import kanjiconv "github.com/sea-turt1e/go-kanjiconv"

// 基本的な使い方（デフォルトで同梱のUniDic辞書を使用）
conv, err := kanjiconv.New(kanjiconv.WithSeparator("/"))

// バイナリサイズを抑えるためレガシーIPA辞書を使い、IPAで解決できない語だけ
// UniDicにフォールバックする場合
import (
    "github.com/ikawaha/kagome-dict/ipa"
    "github.com/ikawaha/kagome-dict/uni"
)

conv, err := kanjiconv.New(
    kanjiconv.WithSeparator("/"),
    kanjiconv.WithDict(ipa.Dict()),
    kanjiconv.WithFallbackDict(uni.Dict()),
)
```

### 読みの取得
```go
text := "今日は良い天気です"

// ひらがなへの変換
fmt.Println(conv.ToHiragana(text))
// きょう/は/よい/てんき/です

// カタカナへの変換
fmt.Println(conv.ToKatakana(text))
// キョウ/ハ/ヨイ/テンキ/デス

// ローマ字への変換
fmt.Println(conv.ToRoman(text))
// kyou/ha/yoi/tenki/desu

// 区切り文字を変える、区切り文字無し
conv, _ = kanjiconv.New(kanjiconv.WithSeparator("_"))
fmt.Println(conv.ToHiragana(text))
// きょう_は_よい_てんき_です

conv, _ = kanjiconv.New(kanjiconv.WithSeparator(""))
fmt.Println(conv.ToHiragana(text))
// きょうはよいてんきです
```

## カスタム漢字読み辞書の使用方法
go-kanjiconvは、IPA辞書やUniDicで正しく認識されない特殊な漢字の読みに対応するカスタム辞書をサポートしています。これは以下のような場合に特に有用です：

1. 特殊な表現や独自の読み方
2. 専門用語や固有名詞
3. 文脈によって複数の読み方がある曖昧な漢字

カスタム辞書はデフォルトで自動的に読み込まれますが、独自の辞書で上書きすることもできます：

```go
conv, err := kanjiconv.New(
    kanjiconv.WithSeparator("/"),
    kanjiconv.WithCustomReadings(kanjiconv.CustomReadings{
        Single: map[string][]string{
            "激": {"げき"},
            "飛": {"と", "ひ"},
        },
        Compound: map[string]string{
            "激を飛ばす": "げきをとばす",
            "飛ばす":     "とばす",
        },
    }),
)

fmt.Println(conv.ToHiragana("激を飛ばす"))
// げき/を/とばす
```

### カスタム辞書の構造

- `Single`：個々の漢字とその読み候補のマッピング。
  - 各漢字は複数の読みを持つことができ、最初の読みがデフォルトとして使用されます。
- `Compound`：複数文字の表現とその読みのマッピング。
  - トークン化前にテキスト置換として適用され、優先されます。

角括弧でテキスト中に読みを直接埋め込むこともでき、その部分はトークナイズをバイパスします： `"[テスト]です"` → `てすと/です`。

## セキュリティ：eMudhra CA鍵漏洩への対応

eMudhraの中間CA証明書8件の秘密鍵が漏洩しました（[詳細](https://emudhra.we-have-a-problem.com/)）。これらを失効扱いにしたGoはGo 1.27.3とGo 1.26.10のため、本プロジェクトではそのどちらかが必要です。

- **最低Goバージョン**：`go.mod` で `go 1.26.10` を指定しています。そのため古いツールチェーンでは、モジュールのビルドが拒否されるか、デフォルトの `GOTOOLCHAIN=auto` であれば新しいツールチェーンが自動でダウンロードされます。本リポジトリ内では `toolchain go1.27.3` により、`go` コマンドはGo 1.27.3を使います。
- **他のGoバージョンでの警告**：CLIは、1.27.3・1.26.10（およびそれぞれの系列の後続パッチリリース）以外のGoでビルドされた場合、標準エラー出力に警告を出します。たとえば1.27.0〜1.27.2や1.28.xが対象です。ライブラリ利用者は `kanjiconv.GoVersionWarning()` で同じチェックができ、対応バージョンなら `""` が返ります：

  ```go
  if w := kanjiconv.GoVersionWarning(); w != "" {
      log.Println(w)
  }
  ```

  ライブラリ自身は何も出力しません。
- **CI（GitHub Actions）**：
  - テストはGo 1.26.10と1.27.3の両方で実行します。`GOTOOLCHAIN=local` を指定しているので、各ジョブは対象のバージョンそのものを使います。
  - Actionはコミットの完全なSHAで固定しており、Dependabotが更新します（7日間のクールダウン付き）。
  - `GITHUB_TOKEN` は読み取り専用（`permissions: contents: read`）で、`actions/checkout` は認証情報を残しません。
  - ジョブにはタイムアウトを設定し、バージョン固定の `govulncheck` で既知の脆弱性を検査します。

go-kanjiconv自体はネットワーク通信を行いません。ただし、お使いの環境にあるcurl、GitHub CLI、Docker、Node.js、Pythonなどの更新が必要な場合があります。修正済みバージョンは[上記ページ](https://emudhra.we-have-a-problem.com/)を参照してください。

## ライセンス

本プロジェクトは[Apache License 2.0](LICENSE)の下でライセンスされています。

### 使用しているオープンソースソフトウェア

- [kagome](https://github.com/ikawaha/kagome): MIT License
- [kagome-dict](https://github.com/ikawaha/kagome-dict)（IPA辞書、`mecab-ipadic`ベース）: MIT License
- [kagome-dict](https://github.com/ikawaha/kagome-dict)（UniDic辞書、オプション）: MIT License

本ライブラリは形態素解析にkagomeとその同梱辞書を使用しています。詳細なライセンス・辞書の帰属表示については、各プロジェクトのLICENSE・NOTICEファイルをご確認ください。

- [kagome LICENSE](https://github.com/ikawaha/kagome/blob/master/LICENSE)
- [kagome-dict/ipa LICENSE](https://github.com/ikawaha/kagome-dict/blob/master/ipa/LICENSE) / [NOTICE](https://github.com/ikawaha/kagome-dict/blob/master/ipa/NOTICE.txt)
- [kagome-dict/uni LICENSE](https://github.com/ikawaha/kagome-dict/blob/master/uni/LICENSE)
