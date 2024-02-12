//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package types

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/montanaflynn/stats"
)

type Percentile struct {
	P50  float64 `json:"p50"`
	P95  float64 `json:"p95"`
	P99  float64 `json:"p99"`
	P999 float64 `json:"p999"`
}

func (x Percentile) String() string {
	return fmt.Sprintf("p50: %-8.2f p95: %-8.2f p99: %-8.2f p999: %-8.2f",
		x.P50, x.P95, x.P99, x.P999)
}

func (x Percentile) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"p50":  encodeVal(x.P50),
		"p95":  encodeVal(x.P95),
		"p99":  encodeVal(x.P99),
		"p999": encodeVal(x.P999),
	})
}

func NewPercentile(seq []float64) Percentile {
	return Percentile{
		P50:  maybeNaN(stats.Percentile(seq, 50.0)),
		P95:  maybeNaN(stats.Percentile(seq, 95.0)),
		P99:  maybeNaN(stats.Percentile(seq, 99.0)),
		P999: maybeNaN(stats.Percentile(seq, 99.9)),
	}
}

type MinMax struct {
	Min, Avg, Max float64
}

func (x MinMax) String() string {
	return fmt.Sprintf("[%8.2f, %8.2f, %8.2f]",
		x.Min, x.Avg, x.Max)
}

func (x MinMax) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"min": encodeVal(x.Min),
		"avg": encodeVal(x.Avg),
		"max": encodeVal(x.Max),
	})
}

func NewMinMax(min, avg, max []float64) MinMax {
	return MinMax{
		Min: maybeNaN(stats.Min(min)),
		Avg: maybeNaN(stats.Mean(avg)),
		Max: maybeNaN(stats.Max(max)),
	}
}

func NewMinMaxSoft(min, avg, max []float64) MinMax {
	return MinMax{
		Min: maybeNaN(stats.Percentile(min, 95.0)),
		Avg: maybeNaN(stats.Percentile(avg, 95.0)),
		Max: maybeNaN(stats.Percentile(max, 95.0)),
	}
}

// Approximate percentile value for threshold X
func PercentileOf(seq []float64, x float64) float64 {
	hi := 100.0
	lo := 0.0
	md := (lo + hi) / 2

	for lo <= hi {
		md = (lo + hi) / 2
		p, _ := stats.Percentile(seq, md)
		switch {
		case p < x:
			lo = md + 0.01
		case p > x:
			hi = md - 0.01
		default:
			return md
		}
	}

	return md
}

func maybeNaN(x float64, err error) float64 { return x }

func encodeVal(x float64) any {
	if math.IsNaN(x) {
		return "NaN"
	}

	return x
}
