package kanjiconv

import (
	"embed"
	"encoding/json"
	"sync"
)

//go:embed data/kana.json data/kanji_readings.json
var embeddedData embed.FS

type kanaData struct {
	Katakana2Roman map[string]string `json:"katakana2roman"`
	FullKatakana   []string          `json:"full_katakana"`
	SmallKatakana  []string          `json:"small_katakana"`
}

var (
	kanaOnce  sync.Once
	kanaCache kanaData
)

func loadKana() kanaData {
	kanaOnce.Do(func() {
		b, err := embeddedData.ReadFile("data/kana.json")
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(b, &kanaCache); err != nil {
			panic(err)
		}
	})
	return kanaCache
}

type customReadingsFile struct {
	Single   map[string][]string `json:"single"`
	Compound map[string]string   `json:"compound"`
}

var (
	defaultCustomReadingsOnce sync.Once
	defaultCustomReadingsData CustomReadings
)

// defaultCustomReadings loads the bundled custom kanji reading dictionary.
func defaultCustomReadings() CustomReadings {
	defaultCustomReadingsOnce.Do(func() {
		b, err := embeddedData.ReadFile("data/kanji_readings.json")
		if err != nil {
			panic(err)
		}
		var f customReadingsFile
		if err := json.Unmarshal(b, &f); err != nil {
			panic(err)
		}
		defaultCustomReadingsData = CustomReadings{
			Single:   f.Single,
			Compound: f.Compound,
		}
	})
	// Return a copy so per-instance mutation doesn't leak via the cache.
	single := make(map[string][]string, len(defaultCustomReadingsData.Single))
	for k, v := range defaultCustomReadingsData.Single {
		readings := make([]string, len(v))
		copy(readings, v)
		single[k] = readings
	}
	compound := make(map[string]string, len(defaultCustomReadingsData.Compound))
	for k, v := range defaultCustomReadingsData.Compound {
		compound[k] = v
	}
	return CustomReadings{Single: single, Compound: compound}
}
