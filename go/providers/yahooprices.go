package providers

import (
	"encoding/json"
	"time"

	"github.com/belazy/monid-finance/fd"
)

// NormalizeYahooPrices validates Yahoo daily bars and replaces the complete
// requested series. Every required Nasdaq date must be present; a truncated or
// downsampled history cannot silently become a partial daily-price response.
func NormalizeYahooPrices(raw json.RawMessage, symbol, startDate, endDate, interval string, requiredDates []string) ([]fd.Price, error) {
	var payload struct {
		Status string `json:"status"`
		Data   struct {
			Symbol  string `json:"symbol"`
			History []struct {
				Timestamp *int64   `json:"timestamp"`
				Open      *float64 `json:"open"`
				High      *float64 `json:"high"`
				Low       *float64 `json:"low"`
				Close     *float64 `json:"close"`
				Volume    *float64 `json:"volume"`
			} `json:"history"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, schemaDriftf("Yahoo historical prices: %v", err)
	}
	if payload.Status != "success" || payload.Data.Symbol != symbol || len(payload.Data.History) == 0 {
		return nil, schemaDriftf("Yahoo historical prices omitted history for %s", symbol)
	}
	bars := make([][]float64, 0, len(payload.Data.History))
	days := make(map[string]bool)
	for i, row := range payload.Data.History {
		if row.Timestamp == nil {
			return nil, schemaDriftf("Yahoo historical price %d omitted timestamp", i)
		}
		day := time.Unix(*row.Timestamp, 0).UTC().Format("2006-01-02")
		if day < startDate || day > endDate {
			continue
		}
		if row.Open == nil || row.High == nil || row.Low == nil || row.Close == nil || row.Volume == nil {
			return nil, schemaDriftf("Yahoo historical price for %s omitted OHLCV values", day)
		}
		if days[day] {
			return nil, schemaDriftf("Yahoo historical prices contain duplicate day %s", day)
		}
		days[day] = true
		bars = append(bars, []float64{float64(*row.Timestamp), *row.Open, *row.High, *row.Low, *row.Close, *row.Volume})
	}
	for _, day := range requiredDates {
		if !days[day] {
			return nil, schemaDriftf("Yahoo daily history does not cover required date %s", day)
		}
	}
	encoded, err := json.Marshal(bars)
	if err != nil {
		return nil, schemaDriftf("Yahoo historical prices contain invalid numbers: %v", err)
	}
	return NormalizePrices(encoded, startDate, endDate, interval)
}
