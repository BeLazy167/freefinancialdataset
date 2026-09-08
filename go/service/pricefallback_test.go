package service

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestPrices_MissingNasdaqVolumeUsesYahooSeries(t *testing.T) {
	nasdaqRaw, err := os.ReadFile("../providers/testdata/nasdaq_spy_missing_volume.json")
	if err != nil {
		t.Fatal(err)
	}
	yahooRaw, err := os.ReadFile("../providers/testdata/yahoo_spy_history.json")
	if err != nil {
		t.Fatal(err)
	}
	svc, transport := newTestService(t, map[string]fakeOutcome{
		"defillama /equities/v1/ohlcv":        {providerHTTPStatus: 404},
		"nasdaq /get_stock_historical_quotes": {output: json.RawMessage(nasdaqRaw)},
		"yahoo-finance /get_historical_data":  {output: json.RawMessage(yahooRaw)},
	})
	result, err := svc.Call(context.Background(), "key", "get_stock_prices", map[string]any{"ticker": "SPY", "start_date": "2026-01-01", "end_date": "2026-09-08"})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(result.Value)
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Time   string
		Close  float64
		Volume int64
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 170 {
		t.Fatalf("want 170 days, got %d", len(rows))
	}
	found := false
	for _, row := range rows {
		if row.Time == "2026-04-20" {
			found = true
			if row.Volume != 43546800 || row.Close < 708.71 || row.Close > 708.73 {
				t.Fatalf("must use whole Yahoo bar: %+v", row)
			}
		}
	}
	if !found {
		t.Fatal("April 20 was dropped")
	}
	calls := transport.Calls()
	if len(calls) != 3 || calls[2].QueryParams["range"] != "5y" || calls[2].QueryParams["interval"] != "1d" {
		t.Fatalf("unexpected fallback: %+v", calls)
	}
}

func TestPricesYahooFailurePropagates(t *testing.T) {
	raw, err := os.ReadFile("../providers/testdata/nasdaq_spy_missing_volume.json")
	if err != nil {
		t.Fatal(err)
	}
	svc, transport := newTestService(t, map[string]fakeOutcome{
		"defillama /equities/v1/ohlcv":        {providerHTTPStatus: 404},
		"nasdaq /get_stock_historical_quotes": {output: json.RawMessage(raw)},
		"yahoo-finance /get_historical_data":  {providerHTTPStatus: 503},
	})
	_, err = svc.Call(context.Background(), "key", "get_stock_prices", map[string]any{"ticker": "SPY", "start_date": "2026-01-01", "end_date": "2026-09-08"})
	if err == nil || transport.CallCount() != 3 {
		t.Fatalf("fallback failure must propagate: %v", err)
	}
}
