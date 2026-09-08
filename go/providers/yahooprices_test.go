package providers

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

func TestYahooPricesRejectIncompleteReplacement(t *testing.T) {
	raw, err := os.ReadFile("testdata/yahoo_spy_history.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"missing volume", func(p map[string]any) {
			delete(p["data"].(map[string]any)["history"].([]any)[0].(map[string]any), "volume")
		}},
		{"missing day", func(p map[string]any) { d := p["data"].(map[string]any); d["history"] = d["history"].([]any)[1:] }},
		{"duplicate day", func(p map[string]any) {
			d := p["data"].(map[string]any)
			h := d["history"].([]any)
			d["history"] = append(h, h[0])
		}},
		{"wrong symbol", func(p map[string]any) { p["data"].(map[string]any)["symbol"] = "QQQ" }},
		{"negative volume", func(p map[string]any) {
			p["data"].(map[string]any)["history"].([]any)[0].(map[string]any)["volume"] = -1
		}},
		{"reversed high low", func(p map[string]any) { p["data"].(map[string]any)["history"].([]any)[0].(map[string]any)["high"] = 1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var payload map[string]any
			if err := json.Unmarshal(raw, &payload); err != nil {
				t.Fatal(err)
			}
			tc.mutate(payload)
			encoded, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			_, err = NormalizeYahooPrices(encoded, "SPY", "2025-09-08", "2026-09-08", "day", []string{"2025-09-08", "2026-04-20"})
			var schema *SchemaDriftError
			if !errors.As(err, &schema) {
				t.Fatalf("expected schema error: %v", err)
			}
		})
	}
}

func TestYahooPricesFilteringAndAggregation(t *testing.T) {
	raw, err := os.ReadFile("testdata/yahoo_spy_history.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		interval string
		count    int
	}{{"day", 4}, {"week", 1}} {
		t.Run(tc.interval, func(t *testing.T) {
			rows, err := NormalizeYahooPrices(raw, "SPY", "2026-09-01", "2026-09-08", tc.interval, []string{"2026-09-01", "2026-09-02", "2026-09-03", "2026-09-04"})
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != tc.count {
				t.Fatalf("got %d rows", len(rows))
			}
			if rows[0].Volume == nil || *rows[0].Volume <= 0 {
				t.Fatal("missing volume")
			}
		})
	}
}
