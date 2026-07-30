package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// runCapture invokes run() and captures stdout/stderr via a pipe.
func runCapture(t *testing.T, args []string) (stdout, stderr string, code int) {
	t.Helper()

	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	code = run(args, outW, errW)

	outW.Close()
	errW.Close()

	var outBuf, errBuf bytes.Buffer
	outBuf.ReadFrom(outR)
	errBuf.ReadFrom(errR)

	return outBuf.String(), errBuf.String(), code
}

func TestRunDefaultRoman(t *testing.T) {
	out, _, code := runCapture(t, []string{"今日は良い天気です"})
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if got, want := strings.TrimSpace(out), "kyou ha yoi tenki desu"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunHiraganaWithSeparatorAfterText(t *testing.T) {
	// Flags placed after the positional text must still be recognized.
	out, _, code := runCapture(t, []string{"今日は良い天気です", "--mode", "hiragana", "--separator", "/"})
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if got, want := strings.TrimSpace(out), "きょう/は/よい/てんき/です"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunShorthandFlags(t *testing.T) {
	out, _, code := runCapture(t, []string{"-m", "katakana", "-s", "/", "今日は良い天気です"})
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if got, want := strings.TrimSpace(out), "キョウ/ハ/ヨイ/テンキ/デス"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunNoText(t *testing.T) {
	_, errOut, code := runCapture(t, []string{"--mode", "hiragana"})
	if code != 2 {
		t.Fatalf("run() code = %d, want 2", code)
	}
	if !strings.Contains(errOut, "no text provided") {
		t.Errorf("stderr = %q, want it to contain %q", errOut, "no text provided")
	}
}

func TestRunInvalidMode(t *testing.T) {
	_, errOut, code := runCapture(t, []string{"text", "--mode", "nope"})
	if code != 2 {
		t.Fatalf("run() code = %d, want 2", code)
	}
	if !strings.Contains(errOut, "invalid mode") {
		t.Errorf("stderr = %q, want it to contain %q", errOut, "invalid mode")
	}
}

func TestRunVersion(t *testing.T) {
	out, _, code := runCapture(t, []string{"--version"})
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if strings.TrimSpace(out) == "" {
		t.Errorf("stdout is empty, want a version string")
	}
}

func TestRunDictFlag(t *testing.T) {
	out, _, code := runCapture(t, []string{"今日は良い天気です", "--mode", "hiragana", "--dict", "ipa"})
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if got, want := strings.TrimSpace(out), "きょう は よい てんき です"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunInvalidDict(t *testing.T) {
	_, errOut, code := runCapture(t, []string{"text", "--dict", "nope"})
	if code != 2 {
		t.Fatalf("run() code = %d, want 2", code)
	}
	if !strings.Contains(errOut, "invalid dict") {
		t.Errorf("stderr = %q, want it to contain %q", errOut, "invalid dict")
	}
}

func TestReorderArgs(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "flags after text",
			in:   []string{"text", "--mode", "hiragana", "--no-custom-readings"},
			want: []string{"--mode", "hiragana", "--no-custom-readings", "text"},
		},
		{
			name: "flags before text",
			in:   []string{"--mode", "hiragana", "text"},
			want: []string{"--mode", "hiragana", "text"},
		},
		{
			name: "equals form",
			in:   []string{"text", "--mode=hiragana"},
			want: []string{"--mode=hiragana", "text"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reorderArgs(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("reorderArgs(%v) = %v, want %v", tt.in, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("reorderArgs(%v) = %v, want %v", tt.in, got, tt.want)
					break
				}
			}
		})
	}
}
