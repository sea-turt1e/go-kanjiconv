// Package kanjiconv converts Japanese text containing kanji into hiragana,
// katakana, or romaji readings.
//
// It uses kagome (https://github.com/ikawaha/kagome) with the bundled UniDic
// dictionary for morphological analysis by default, since UniDic's larger
// vocabulary handles proper nouns and new words noticeably better than the
// legacy IPA dictionary (mecab-ipadic, last updated in 2007). This mirrors
// the design of the Python kanjiconv library
// (https://github.com/sea-turt1e/kanjiconv), which is built on
// SudachiPy/SudachiDict; because the underlying dictionaries differ,
// readings for some proper nouns or new words may not match exactly.
package kanjiconv

import (
	"fmt"

	"github.com/ikawaha/kagome-dict/dict"
	"github.com/ikawaha/kagome-dict/uni"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// CustomReadings holds fallback readings used when the morphological
// analyzer cannot resolve a reading for a token.
type CustomReadings struct {
	// Single maps an individual kanji to its candidate readings; the first
	// entry is used as the default.
	Single map[string][]string
	// Compound maps multi-character expressions to their reading. These are
	// applied as a text substitution before tokenization.
	Compound map[string]string
}

// KanjiConv converts Japanese text to hiragana, katakana, or romaji.
type KanjiConv struct {
	tokenizer    *tokenizer.Tokenizer
	kana         kanaData
	separator    string
	tokenizeMode tokenizer.TokenizeMode

	useCustomReadings bool
	customReadings    CustomReadings

	mainDict          *dict.Dict
	mainDictIsUni     bool
	fallbackDict      *dict.Dict
	fallbackDictIsUni bool
	fallbackTokenizer *tokenizer.Tokenizer
}

// Option configures a KanjiConv instance.
type Option func(*KanjiConv)

// WithSeparator sets the separator inserted between token readings.
// The default is a single space.
func WithSeparator(sep string) Option {
	return func(k *KanjiConv) {
		k.separator = sep
	}
}

// WithoutCustomReadings disables the bundled custom reading fallback.
func WithoutCustomReadings() Option {
	return func(k *KanjiConv) {
		k.useCustomReadings = false
	}
}

// WithCustomReadings replaces the custom reading dictionary used as a
// fallback when the tokenizer cannot resolve a reading. Passing this option
// also enables the custom reading fallback.
func WithCustomReadings(cr CustomReadings) Option {
	return func(k *KanjiConv) {
		k.useCustomReadings = true
		k.customReadings = cr
	}
}

// WithDict overrides the main dictionary used for tokenization. The default
// is the bundled UniDic dictionary (uni.Dict()). Passing a smaller dictionary
// such as the IPA dictionary (github.com/ikawaha/kagome-dict/ipa) trades
// reading accuracy on proper nouns and new words for a smaller binary size.
func WithDict(d *dict.Dict) Option {
	return func(k *KanjiConv) {
		k.mainDict = d
	}
}

// WithFallbackDict sets a secondary dictionary used to resolve readings when
// the main dictionary has no reading for a token. This is useful when
// overriding the main dictionary with WithDict and still wanting a second
// opinion from another dictionary for words the main one misses.
func WithFallbackDict(d *dict.Dict) Option {
	return func(k *KanjiConv) {
		k.fallbackDict = d
	}
}

// WithTokenizeMode sets the kagome tokenize mode (Normal, Search, or
// Extended). The default is tokenizer.Normal.
func WithTokenizeMode(m tokenizer.TokenizeMode) Option {
	return func(k *KanjiConv) {
		k.tokenizeMode = m
	}
}

// New creates a KanjiConv instance.
func New(opts ...Option) (*KanjiConv, error) {
	k := &KanjiConv{
		kana:              loadKana(),
		separator:         " ",
		tokenizeMode:      tokenizer.Normal,
		useCustomReadings: true,
		customReadings:    defaultCustomReadings(),
	}

	for _, opt := range opts {
		opt(k)
	}

	if k.mainDict == nil {
		k.mainDict = uni.Dict()
	}
	k.mainDictIsUni = isUniDict(k.mainDict)

	t, err := tokenizer.New(k.mainDict)
	if err != nil {
		return nil, fmt.Errorf("kanjiconv: failed to create tokenizer: %w", err)
	}
	k.tokenizer = t

	if k.fallbackDict != nil {
		k.fallbackDictIsUni = isUniDict(k.fallbackDict)
		ft, err := tokenizer.New(k.fallbackDict)
		if err != nil {
			return nil, fmt.Errorf("kanjiconv: failed to create fallback tokenizer: %w", err)
		}
		k.fallbackTokenizer = ft
	}

	return k, nil
}

// isUniDict reports whether d is the UniDic dictionary, which stores its
// reading (LForm) at a fixed feature offset rather than in the generic
// "_reading" contents metadata that tokenizer.Token.Reading relies on.
func isUniDict(d *dict.Dict) bool {
	info := d.Info()
	return info != nil && info.Name == uni.DictName
}

// ToHiragana converts the given text to its hiragana reading.
func (k *KanjiConv) ToHiragana(text string) string {
	joined := k.resolveReadings(text)
	return katakanaToHiragana(joined)
}

// ToKatakana converts the given text to its katakana reading.
func (k *KanjiConv) ToKatakana(text string) string {
	return k.resolveReadings(text)
}

// ToRoman converts the given text to its romaji reading.
func (k *KanjiConv) ToRoman(text string) string {
	joined := k.resolveReadings(text)
	return katakanaToRoman(joined, k.kana)
}
