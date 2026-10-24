// Command kanjiconv converts Japanese text containing kanji to hiragana,
// katakana, or romaji.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	kanjiconv "github.com/sea-turt1e/go-kanjiconv"
)

// version is set via -ldflags "-X main.version=..." at release build time.
var version = "0.0.0"

// boolFlagNames lists flags that take no value, so reorderArgs knows not to
// consume the following token as their argument.
var boolFlagNames = map[string]bool{
	"no-custom-readings": true,
	"version":            true,
	"h":                  true,
	"help":               true,
}

// reorderArgs moves flag tokens (and their values) before positional
// arguments, so that flags may appear after the text, e.g.
// `kanjiconv "text" --mode hiragana`. Go's flag package otherwise stops
// parsing flags at the first positional argument.
func reorderArgs(args []string) []string {
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if len(arg) < 2 || arg[0] != '-' {
			positional = append(positional, arg)
			continue
		}
		flags = append(flags, arg)
		name := strings.TrimLeft(arg, "-")
		if strings.Contains(name, "=") {
			continue // value is embedded, nothing more to consume
		}
		if boolFlagNames[name] {
			continue
		}
		if i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}
	return append(flags, positional...)
}

func run(args []string, stdout, stderr *os.File) int {
	if w := kanjiconv.GoVersionWarning(); w != "" {
		fmt.Fprintln(stderr, "kanjiconv: "+w)
	}

	args = reorderArgs(args)
	fs := flag.NewFlagSet("kanjiconv", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: kanjiconv [flags] TEXT...\n\n")
		fmt.Fprintf(stderr, "Convert Japanese text containing kanji to hiragana, katakana, or romaji.\n\n")
		fmt.Fprintf(stderr, "Flags:\n")
		fs.PrintDefaults()
	}

	var mode string
	fs.StringVar(&mode, "mode", "roman", "Conversion mode: roman, hiragana, or katakana.")
	fs.StringVar(&mode, "m", "roman", "Shorthand for --mode.")

	var separator string
	fs.StringVar(&separator, "separator", " ", "Separator inserted between token readings.")
	fs.StringVar(&separator, "s", " ", "Shorthand for --separator.")

	var dictName string
	fs.StringVar(&dictName, "dict", "uni", "Dictionary to use: uni (default, better proper noun coverage) or ipa (smaller, legacy).")

	var noCustomReadings bool
	fs.BoolVar(&noCustomReadings, "no-custom-readings", false, "Disable custom reading fallback.")

	var tokenizeModeFlag string
	fs.StringVar(&tokenizeModeFlag, "tokenize-mode", "normal", "Tokenize mode: normal, search, or extended.")

	var showVersion bool
	fs.BoolVar(&showVersion, "version", false, "Show the installed version.")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	if showVersion {
		fmt.Fprintln(stdout, version)
		return 0
	}

	text := strings.Join(fs.Args(), " ")
	if text == "" {
		fmt.Fprintln(stderr, "kanjiconv: no text provided")
		fs.Usage()
		return 2
	}

	switch mode {
	case "roman", "hiragana", "katakana":
	default:
		fmt.Fprintf(stderr, "kanjiconv: invalid mode %q (must be roman, hiragana, or katakana)\n", mode)
		return 2
	}

	var tokMode tokenizer.TokenizeMode
	switch tokenizeModeFlag {
	case "normal":
		tokMode = tokenizer.Normal
	case "search":
		tokMode = tokenizer.Search
	case "extended":
		tokMode = tokenizer.Extended
	default:
		fmt.Fprintf(stderr, "kanjiconv: invalid tokenize-mode %q (must be normal, search, or extended)\n", tokenizeModeFlag)
		return 2
	}

	opts := []kanjiconv.Option{
		kanjiconv.WithSeparator(separator),
		kanjiconv.WithTokenizeMode(tokMode),
	}
	if noCustomReadings {
		opts = append(opts, kanjiconv.WithoutCustomReadings())
	}
	switch dictName {
	case "uni":
		// UniDic is the library's default; nothing to add.
	case "ipa":
		opts = append(opts, kanjiconv.WithDict(ipa.Dict()))
	default:
		fmt.Fprintf(stderr, "kanjiconv: invalid dict %q (must be uni or ipa)\n", dictName)
		return 2
	}

	converter, err := kanjiconv.New(opts...)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	var result string
	switch mode {
	case "hiragana":
		result = converter.ToHiragana(text)
	case "katakana":
		result = converter.ToKatakana(text)
	default:
		result = converter.ToRoman(text)
	}

	fmt.Fprintln(stdout, result)
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
