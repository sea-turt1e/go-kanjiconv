# go-kanjiconv
![Go](https://img.shields.io/badge/-Go-00ADD8.svg?logo=go&style=flat)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

**EN** | [JA](README_ja.md)

kanjiconv for Golang.

Kanji Converter to Hiragana, Katakana, Roman alphabet.
This is a Go port of the Python library [kanjiconv](https://github.com/sea-turt1e/kanjiconv). You can get the reading and pronunciation of Japanese sentences using [kagome](https://github.com/ikawaha/kagome), a pure-Go Japanese morphological analyzer with an embedded dictionary — no external services or dictionary downloads required.

## Differences from the Python version

The Python library is built on SudachiPy/SudachiDict. This Go port uses [kagome](https://github.com/ikawaha/kagome) with the **UniDic dictionary as its default**, since there is no actively maintained Go binding for SudachiDict, and UniDic's larger, more current vocabulary handles proper nouns and new words noticeably better than the legacy IPA dictionary (`mecab-ipadic`, last updated in 2007). As a result:

- Readings generally match the Python version closely, but tokenization boundaries can differ slightly (e.g. a compound proper noun may be split into two tokens instead of one).
- The UniDic dictionary bundled here is a different distribution (`kagome-dict/uni`) than the Python version's optional `unidic` package.
- The lighter, legacy IPA dictionary is still available via `WithDict(ipa.Dict())` (library) or `--dict ipa` (CLI) if you'd rather trade some reading accuracy for a smaller binary — see [Environment](#environment) for the size difference.
- There is no equivalent to SudachiPy's split mode (A/B/C granularity); instead, kagome's own `TokenizeMode` (Normal/Search/Extended) is exposed via `WithTokenizeMode`.

## Environment
```
Go 1.27.3 (recommended) or Go 1.26.10
```

`go.mod` declares `go 1.26.10` and `toolchain go1.27.3`. Later patch releases of the same lines (e.g. 1.27.4, 1.26.11) also work. See [Security: eMudhra CA compromise](#security-emudhra-ca-compromise) for why these versions are required.

The bundled UniDic dictionary is embedded in the binary via `go:embed`/kagome and adds roughly 45MB to your binary size (no external downloads needed). If you need a smaller binary and can tolerate weaker proper-noun coverage, use the legacy IPA dictionary instead (`WithDict(ipa.Dict())` in the library, or `--dict ipa` on the CLI), which adds about 11MB.

## Install
```bash
go get github.com/sea-turt1e/go-kanjiconv
```

To install the CLI:
```bash
go install github.com/sea-turt1e/go-kanjiconv/cmd/kanjiconv@latest
```

## How to use

### CLI usage
```bash
kanjiconv "幽☆遊☆白書は、最高の漫画デス。"
# yuuyuu hakusho ha ,  saikou no manga desu .
```

By default, the CLI converts text to roman alphabet and inserts a single space between token readings.

Use `-m`/`--mode` to choose the output format:

```bash
# Convert to hiragana
kanjiconv "幽☆遊☆白書は、最高の漫画デス。" --mode hiragana
# ゆうゆう はくしょ は 、 さいこう の まんが です 。

# Convert to katakana
kanjiconv "幽☆遊☆白書は、最高の漫画デス。" --mode katakana
# ユウユウ ハクショ ハ 、 サイコウ ノ マンガ デス 。

# Convert to roman alphabet (which is already the default)
kanjiconv "幽☆遊☆白書は、最高の漫画デス。" --mode roman
# yuuyuu hakusho ha ,  saikou no manga desu .
```

Use `-s`/`--separator` to change the separator inserted between token readings:

```bash
kanjiconv "今日は良い天気です" --mode hiragana --separator "/"
# きょう/は/よい/てんき/です

kanjiconv "今日は良い天気です" --mode hiragana --separator ""
# きょうはよいてんきです
```

Flags may be placed before or after the text argument:

```bash
kanjiconv "今日は良い天気です" --mode hiragana --separator "/"
```

Additional options:

```bash
# Use the smaller, legacy IPA dictionary instead of the UniDic default
kanjiconv "東京に行く" --mode hiragana --dict ipa

# Disable the custom reading fallback
kanjiconv "激を飛ばす" --mode hiragana --no-custom-readings
```

## CLI flags/options

| Option                                    | Description                                                            |
| ----------------------------------------- | ------------------------------------------------------------------------ |
| `-m`/`--mode {roman,hiragana,katakana}`   | Conversion mode. Defaults to `roman`.                                    |
| `-s`/`--separator SEPARATOR`              | Separator inserted between token readings. Defaults to a single space.  |
| `--dict {uni,ipa}`                        | Dictionary to use. Defaults to `uni` (better proper noun coverage); `ipa` is smaller but legacy. |
| `--no-custom-readings`                    | Disable custom reading fallback.                                        |
| `--tokenize-mode {normal,search,extended}`| kagome tokenize mode. Defaults to `normal`.                              |
| `--version`                               | Show the installed version.                                             |
| `-h`/`--help`                             | Show help.                                                              |

### Import & Create Instance
```go
import kanjiconv "github.com/sea-turt1e/go-kanjiconv"

// Basic usage (uses the bundled UniDic dictionary by default)
conv, err := kanjiconv.New(kanjiconv.WithSeparator("/"))

// Trade reading accuracy for a smaller binary by using the legacy IPA
// dictionary instead, optionally with UniDic as a fallback for words IPA
// can't resolve
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

### Get Reading
```go
text := "今日は良い天気です"

// convert to hiragana
fmt.Println(conv.ToHiragana(text))
// きょう/は/よい/てんき/です

// convert to katakana
fmt.Println(conv.ToKatakana(text))
// キョウ/ハ/ヨイ/テンキ/デス

// convert to Roman alphabet
fmt.Println(conv.ToRoman(text))
// kyou/ha/yoi/tenki/desu

// You can change the separator to another character or an empty string
conv, _ = kanjiconv.New(kanjiconv.WithSeparator("_"))
fmt.Println(conv.ToHiragana(text))
// きょう_は_よい_てんき_です

conv, _ = kanjiconv.New(kanjiconv.WithSeparator(""))
fmt.Println(conv.ToHiragana(text))
// きょうはよいてんきです
```

## Using Custom Kanji Reading Dictionary
go-kanjiconv supports a custom dictionary for handling special kanji readings that are not properly recognized by the IPA dictionary or UniDic. This is particularly useful for:

1. Special expressions with unique readings
2. Technical terms or proper nouns
3. Ambiguous kanji with multiple readings based on context

A bundled custom dictionary is loaded by default, but you can override it with your own:

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

### Custom Dictionary Structure

- `Single`: maps an individual kanji to its candidate readings.
  - Each kanji can have multiple readings; the first one is used as the default.
- `Compound`: maps multi-character expressions to their reading.
  - These are applied as a text substitution before tokenization, and take priority.

Literal readings can also be inlined in the input text using square brackets, which bypass tokenization entirely: `"[テスト]です"` → `てすと/です`.

## Security: eMudhra CA compromise

The private keys of eight eMudhra intermediate CA certificates were compromised ([details](https://emudhra.we-have-a-problem.com/)). Go 1.27.3 and Go 1.26.10 are the releases that distrust them, so this project requires one of them:

- **Minimum Go version**: `go.mod` declares `go 1.26.10`, so older toolchains either refuse to build the module or (with the default `GOTOOLCHAIN=auto`) download a newer one. Inside this repository, `toolchain go1.27.3` makes the `go` command use Go 1.27.3.
- **Warning for other Go versions**: the CLI prints a warning to stderr when it was built with a Go version other than 1.27.3 or 1.26.10 (or a later patch release of those lines), such as 1.27.0–1.27.2 or 1.28.x. Library users can run the same check with `kanjiconv.GoVersionWarning()`, which returns `""` for supported versions:

  ```go
  if w := kanjiconv.GoVersionWarning(); w != "" {
      log.Println(w)
  }
  ```

  The library itself never prints anything.
- **CI (GitHub Actions)**:
  - Tests run on Go 1.26.10 and 1.27.3 with `GOTOOLCHAIN=local`, so each job uses exactly the version under test.
  - Actions are pinned to full commit SHAs, and Dependabot keeps them up to date (with a 7-day cooldown).
  - The `GITHUB_TOKEN` is read-only (`permissions: contents: read`), and `actions/checkout` does not persist credentials.
  - Jobs have a timeout, and `govulncheck` (pinned version) checks for known vulnerabilities.

go-kanjiconv makes no network connections itself. Other software on your machine may still need updating: curl, the GitHub CLI, Docker, Node.js, Python and so on. See the [page above](https://emudhra.we-have-a-problem.com/) for patched versions.

## License

This project is licensed under the [Apache License 2.0](LICENSE).

### Open Source Software Used

- [kagome](https://github.com/ikawaha/kagome): MIT License
- [kagome-dict](https://github.com/ikawaha/kagome-dict) (IPA dictionary, based on `mecab-ipadic`): MIT License
- [kagome-dict](https://github.com/ikawaha/kagome-dict) (UniDic dictionary, optional): MIT License

This library uses kagome and its bundled dictionaries for morphological analysis. For detailed license and dictionary attribution information, see the LICENSE and NOTICE files of each project:

- [kagome LICENSE](https://github.com/ikawaha/kagome/blob/master/LICENSE)
- [kagome-dict/ipa LICENSE](https://github.com/ikawaha/kagome-dict/blob/master/ipa/LICENSE) / [NOTICE](https://github.com/ikawaha/kagome-dict/blob/master/ipa/NOTICE.txt)
- [kagome-dict/uni LICENSE](https://github.com/ikawaha/kagome-dict/blob/master/uni/LICENSE)
