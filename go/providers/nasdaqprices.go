package providers

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/belazy/monid-finance/fd"
)

// IncompletePricesError identifies explicit missing values in otherwise
// structured Nasdaq bars. Dates lets a fallback prove it covers the same days.
type IncompletePricesError struct {
	Dates []string
}

func (e *IncompletePricesError) Error() string {
	return "Nasdaq historical quotes contain missing OHLCV values"
}

// Unwrap preserves the existing schema-error classification if no fallback runs.
func (e *IncompletePricesError) Unwrap() error { return schemaDriftf("%s", e.Error()) }

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
	dates := make([]time.Time, len(rows))
	requiredDates := make([]string, 0, len(rows))
	incomplete := false
	for i, row := range rows {
		date, err := time.Parse("01/02/2006", row["date"])
		if err != nil {
			return nil, schemaDriftf("Nasdaq historical quote %d has invalid date", i)
		}
		dates[i] = date
		day := date.Format("2006-01-02")
		if day < startDate || day > endDate {
			continue
		}
		requiredDates = append(requiredDates, day)
		for _, key := range []string{"open", "high", "low", "close", "volume"} {
			switch strings.ToUpper(strings.TrimSpace(row[key])) {
			case "", "N/A", "--":
				incomplete = true
			}
		}
	}
	if incomplete {
		return nil, &IncompletePricesError{Dates: requiredDates}
	}
	bars := make([][]float64, 0, len(rows))
	for i, row := range rows {
		day := dates[i].Format("2006-01-02")
		if day < startDate || day > endDate {
			continue
		}
		bar := []float64{float64(dates[i].Unix())}
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
