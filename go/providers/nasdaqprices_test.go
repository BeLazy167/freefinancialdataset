package providers

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestNormalizeNasdaqPrices(t *testing.T) {
	raw, err := os.ReadFile("testdata/nasdaq_spy_prices.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, interval, start string
		count                 int
	}{
		{"daily", "day", "2026-09-01", 4},
		{"date filter", "day", "2026-09-03", 2},
		{"weekly", "week", "2026-09-01", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prices, err := NormalizeNasdaqPrices(raw, tc.start, "2026-09-08", tc.interval)
			if err != nil {
				t.Fatal(err)
			}
			if len(prices) != tc.count {
				t.Fatalf("want %d bars, got %d", tc.count, len(prices))
			}
			if prices[len(prices)-1].Close == nil || *prices[len(prices)-1].Close != 770.19 {
				t.Fatalf("wrong last close: %#v", prices)
			}
		})
	}
	for _, tc := range []struct {
		name string
		raw  string
	}{
		{"truncated", strings.Replace(string(raw), `"totalRecords": 4`, `"totalRecords": 5`, 1)},
		{"bad volume", strings.Replace(string(raw), `34,054,200`, `not-a-number`, 1)},
		{"missing table", `{"data":null}`},
		{"bad date", strings.Replace(string(raw), `09/04/2026`, `invalid`, 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NormalizeNasdaqPrices(json.RawMessage(tc.raw), "2026-09-01", "2026-09-08", "day")
			if _, ok := err.(*SchemaDriftError); !ok {
				t.Fatalf("want schema error, got %v", err)
			}
		})
	}
}

func TestNasdaqMissingVolumeCarriesRequiredDates(t *testing.T) {
	raw, err := os.ReadFile("testdata/nasdaq_spy_missing_volume.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = NormalizeNasdaqPrices(raw, "2026-01-01", "2026-09-08", "day")
	incomplete, ok := err.(*IncompletePricesError)
	if !ok || len(incomplete.Dates) != 170 {
		t.Fatalf("want 170 required dates: %v", err)
	}
	// Missing data outside the requested window does not trigger a fallback.
	rows, err := NormalizeNasdaqPrices(raw, "2026-09-01", "2026-09-08", "day")
	if err != nil || len(rows) != 4 {
		t.Fatalf("unrelated missing volume broke short request: %v", err)
	}
}
