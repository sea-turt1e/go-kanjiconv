package kanjiconv

import (
	"testing"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome-dict/uni"
)

func TestToHiragana(t *testing.T) {
	k, err := New(WithSeparator("/"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	got := k.ToHiragana("今日は良い天気です")
	want := "きょう/は/よい/てんき/です"
	if got != want {
		t.Errorf("ToHiragana() = %q, want %q", got, want)
	}
}

func TestToKatakana(t *testing.T) {
	k, err := New(WithSeparator("/"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	got := k.ToKatakana("今日は良い天気です")
	want := "キョウ/ハ/ヨイ/テンキ/デス"
	if got != want {
		t.Errorf("ToKatakana() = %q, want %q", got, want)
	}
}

func TestToRoman(t *testing.T) {
	k, err := New(WithSeparator("/"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	got := k.ToRoman("今日は良い天気です")
	want := "kyou/ha/yoi/tenki/desu"
	if got != want {
		t.Errorf("ToRoman() = %q, want %q", got, want)
	}
}

func TestSeparator(t *testing.T) {
	tests := []struct {
		separator string
		want      string
	}{
		{" ", "きょう は よい てんき です"},
		{"", "きょうはよいてんきです"},
		{"_", "きょう_は_よい_てんき_です"},
	}
	for _, tt := range tests {
		k, err := New(WithSeparator(tt.separator))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		got := k.ToHiragana("今日は良い天気です")
		if got != tt.want {
			t.Errorf("ToHiragana() with separator %q = %q, want %q", tt.separator, got, tt.want)
		}
	}
}

func TestCustomReadingsCompound(t *testing.T) {
	k, err := New(
		WithSeparator("/"),
		WithCustomReadings(CustomReadings{
			Compound: map[string]string{"激を飛ばす": "げきをとばす"},
		}),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	got := k.ToHiragana("激を飛ばす")
	want := "げき/を/とばす"
	if got != want {
		t.Errorf("ToHiragana() = %q, want %q", got, want)
	}
}

func TestCustomReadingsSingleFallback(t *testing.T) {
	// "㐬" (a rare CJK extension character not in UniDic's vocabulary) is
	// tokenized as an UNKNOWN token with no reading, so it exercises the
	// custom single reading fallback.
	k, err := New(
		WithSeparator("/"),
		WithCustomReadings(CustomReadings{
			Single: map[string][]string{"㐬": {"テスト"}},
		}),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	got := k.ToHiragana("☆㐬☆")
	want := "☆/てすと/☆"
	if got != want {
		t.Errorf("ToHiragana() = %q, want %q", got, want)
	}
}

func TestWithoutCustomReadings(t *testing.T) {
	k, err := New(
		WithSeparator("/"),
		WithCustomReadings(CustomReadings{
			Single: map[string][]string{"㐬": {"テスト"}},
		}),
		WithoutCustomReadings(),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	// WithoutCustomReadings is applied after WithCustomReadings here, so the
	// fallback stays disabled and the unresolved token falls back to its
	// surface form.
	got := k.ToHiragana("☆㐬☆")
	want := "☆/㐬/☆"
	if got != want {
		t.Errorf("ToHiragana() = %q, want %q", got, want)
	}
}

func TestLiteralBracketReading(t *testing.T) {
	k, err := New(WithSeparator("/"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	got := k.ToHiragana("[テスト]です")
	want := "てすと/です"
	if got != want {
		t.Errorf("ToHiragana() = %q, want %q", got, want)
	}
}

func TestWithFallbackDict(t *testing.T) {
	// "遊" surrounded by "☆" is UNKNOWN with no reading in the (smaller,
	// legacy) IPA dictionary, so this exercises falling back to UniDic when
	// the main dictionary is overridden to something lighter than the
	// default.
	k, err := New(
		WithSeparator("/"),
		WithDict(ipa.Dict()),
		WithFallbackDict(uni.Dict()),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	got := k.ToHiragana("☆遊☆")
	want := "☆/ゆう/☆"
	if got != want {
		t.Errorf("ToHiragana() = %q, want %q", got, want)
	}
}
