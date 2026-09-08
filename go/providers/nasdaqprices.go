package providers

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/belazy/monid-finance/fd"
)

// NormalizeNasdaqPrices converts Nasdaq daily historical quotes to the same
// validated, date-filtered and aggregated bars as the primary price provider.
func NormalizeNasdaqPrices(raw json.RawMessage, startDate, endDate, interval string) ([]fd.Price, error) {
	var payload struct {
		TotalRecords *int `json:"totalRecords"`
		TradesTable  *struct {
			Rows []map[string]string `json:"rows"`
		} `json:"tradesTable"`
		Data json.RawMessage `json:"data"`
	}
	for depth := 0; depth < 4; depth++ {
		payload.TotalRecords = nil
		payload.TradesTable = nil
		payload.Data = nil
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, schemaDriftf("Nasdaq historical quotes: %v", err)
		}
		if payload.TradesTable != nil {
			break
		}
		if len(payload.Data) == 0 {
			break
		}
		raw = payload.Data
	}
	if payload.TradesTable == nil || payload.TotalRecords == nil {
		return nil, schemaDriftf("Nasdaq historical quotes omitted tradesTable or totalRecords")
	}
	rows := payload.TradesTable.Rows
	if *payload.TotalRecords != len(rows) {
		return nil, schemaDriftf("Nasdaq historical quotes returned %d of %d rows", len(rows), *payload.TotalRecords)
	}
	bars := make([][]float64, 0, len(rows))
	for i, row := range rows {
		date, err := time.Parse("01/02/2006", row["date"])
		if err != nil {
			return nil, schemaDriftf("Nasdaq historical quote %d has invalid date", i)
		}
		bar := []float64{float64(date.Unix())}
		for _, key := range []string{"open", "high", "low", "close", "volume"} {
			value, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimPrefix(strings.TrimSpace(row[key]), "$"), ",", ""), 64)
			if err != nil {
				return nil, schemaDriftf("Nasdaq historical quote %d has invalid %s", i, key)
			}
			bar = append(bar, value)
		}
		bars = append(bars, bar)
	}
	encoded, err := json.Marshal(bars)
	if err != nil {
		return nil, schemaDriftf("Nasdaq historical quotes contain invalid numbers: %v", err)
	}
	return NormalizePrices(encoded, startDate, endDate, interval)
}
