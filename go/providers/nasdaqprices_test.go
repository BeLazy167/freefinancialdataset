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
		{"bad volume", strings.Replace(string(raw), `34,054,200`, `N/A`, 1)},
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
