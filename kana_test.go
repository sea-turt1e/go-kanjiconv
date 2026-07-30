package kanjiconv

import "testing"

func TestKatakanaToHiragana(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain katakana", "ユウユウハクショ", "ゆうゆうはくしょ"},
		{"mixed with punctuation", "ハ、サイコウ。", "は、さいこう。"},
		{"non-katakana untouched", "幽☆遊☆白書", "幽☆遊☆白書"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := katakanaToHiragana(tt.in); got != tt.want {
				t.Errorf("katakanaToHiragana(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestKatakanaToRoman(t *testing.T) {
	kana := loadKana()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"basic", "ユウユウハクショ", "yuuyuuhakusho"},
		{"with punctuation", "ハ、サイコウ。", "ha, saikou. "},
		{"contracted sound", "キャベツ", "kyabetsu"},
		{"long vowel mark", "デス", "desu"},
		{"non-katakana untouched", "☆", "☆"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := katakanaToRoman(tt.in, kana); got != tt.want {
				t.Errorf("katakanaToRoman(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
