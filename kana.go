package kanjiconv

// katakanaToHiragana converts katakana runes in text to hiragana.
func katakanaToHiragana(text string) string {
	runes := []rune(text)
	out := make([]rune, len(runes))
	for i, r := range runes {
		if r >= 'ァ' && r <= 'ン' {
			out[i] = r - 0x60
		} else {
			out[i] = r
		}
	}
	return string(out)
}

// katakanaToRoman converts katakana runes in text to their romaji equivalent,
// handling contracted sounds (拗音) formed from a full-width katakana followed
// by a small katakana (e.g. キャ -> kya).
func katakanaToRoman(text string, kana kanaData) string {
	fullSet := make(map[rune]struct{}, len(kana.FullKatakana))
	for _, s := range kana.FullKatakana {
		for _, r := range s {
			fullSet[r] = struct{}{}
		}
	}
	smallSet := make(map[rune]struct{}, len(kana.SmallKatakana))
	for _, s := range kana.SmallKatakana {
		for _, r := range s {
			smallSet[r] = struct{}{}
		}
	}

	runes := []rune(text)
	var out []rune
	i := 0
	for i < len(runes) {
		if i+1 < len(runes) {
			_, isFull := fullSet[runes[i]]
			_, isSmall := smallSet[runes[i+1]]
			if isFull && isSmall {
				combined := string(runes[i]) + string(runes[i+1])
				if roman, ok := kana.Katakana2Roman[combined]; ok {
					out = append(out, []rune(roman)...)
					i += 2
					continue
				}
			}
		}

		r := runes[i]
		if (r >= 'ァ' && r <= 'ン') || (r >= '、' && r <= 'ー') {
			if roman, ok := kana.Katakana2Roman[string(r)]; ok {
				out = append(out, []rune(roman)...)
			}
		} else {
			out = append(out, r)
		}
		i++
	}
	return string(out)
}
