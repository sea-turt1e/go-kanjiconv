package kanjiconv

import (
	"strings"

	"github.com/ikawaha/kagome-dict/uni"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// tokenReading extracts a token's reading, accounting for the fact that
// UniDic does not populate the generic "_reading" contents metadata that
// tokenizer.Token.Reading relies on; its reading is instead the LForm
// feature at a fixed offset. See kagome-dict/uni's dict.go for the layout.
func tokenReading(tok tokenizer.Token, dictIsUni bool) (string, bool) {
	if dictIsUni {
		if reading, ok := tok.FeatureAt(uni.LForm); ok && reading != "" {
			return reading, true
		}
		return "", false
	}
	if reading, ok := tok.Reading(); ok && reading != "" {
		return reading, true
	}
	return "", false
}

// resolveReadings tokenizes text and returns the separator-joined katakana
// readings, applying compound/single custom-reading fallbacks and honoring
// literal [reading] segments.
func (k *KanjiConv) resolveReadings(text string) string {
	if k.useCustomReadings {
		for compound, reading := range k.customReadings.Compound {
			text = strings.ReplaceAll(text, compound, reading)
		}
	}

	segments := splitLiteralSegments(text)

	var readings []string
	for _, seg := range segments {
		if seg.literal {
			readings = append(readings, seg.text)
			continue
		}
		readings = append(readings, k.readSegment(seg.text)...)
	}

	return strings.Join(readings, k.separator)
}

// readSegment tokenizes a non-literal chunk of text and resolves the
// reading for each token, falling back to the secondary dictionary and then
// the custom single-kanji dictionary when the main dictionary has no
// reading.
func (k *KanjiConv) readSegment(text string) []string {
	tokens := k.tokenizer.Analyze(text, k.tokenizeMode)

	readings := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		if tok.Class == tokenizer.DUMMY {
			continue
		}
		surface := tok.Surface
		if surface == "" {
			continue
		}

		reading, hasReading := tokenReading(tok, k.mainDictIsUni)
		if hasReading && reading == surface {
			hasReading = false
		}

		if !hasReading && k.fallbackTokenizer != nil {
			if fallbackReading, found := k.readingFromFallback(surface); found {
				reading = fallbackReading
				hasReading = true
			}
		}

		if !hasReading && k.useCustomReadings {
			if candidates, found := k.customReadings.Single[surface]; found && len(candidates) > 0 {
				reading = candidates[0]
				hasReading = true
			}
		}

		if hasReading {
			readings = append(readings, reading)
		} else {
			readings = append(readings, surface)
		}
	}
	return readings
}

// readingFromFallback re-tokenizes surface with the fallback dictionary and
// concatenates the readings of the resulting tokens.
func (k *KanjiConv) readingFromFallback(surface string) (string, bool) {
	tokens := k.fallbackTokenizer.Analyze(surface, k.tokenizeMode)
	var sb strings.Builder
	found := false
	for _, tok := range tokens {
		if tok.Class == tokenizer.DUMMY || tok.Surface == "" {
			continue
		}
		if reading, ok := tokenReading(tok, k.fallbackDictIsUni); ok {
			sb.WriteString(reading)
			found = true
		} else {
			sb.WriteString(tok.Surface)
		}
	}
	if !found {
		return "", false
	}
	return sb.String(), true
}

type textSegment struct {
	text    string
	literal bool
}

// splitLiteralSegments splits text on [...] brackets, marking the bracket
// contents (with brackets stripped) as literal readings that bypass
// tokenization entirely.
func splitLiteralSegments(text string) []textSegment {
	var segments []textSegment
	rest := text
	for {
		start := strings.IndexRune(rest, '[')
		if start == -1 {
			if rest != "" {
				segments = append(segments, textSegment{text: rest})
			}
			break
		}
		end := strings.IndexRune(rest[start:], ']')
		if end == -1 {
			if rest != "" {
				segments = append(segments, textSegment{text: rest})
			}
			break
		}
		end += start

		if start > 0 {
			segments = append(segments, textSegment{text: rest[:start]})
		}
		segments = append(segments, textSegment{text: rest[start+1 : end], literal: true})
		rest = rest[end+1:]
	}
	return segments
}
