//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package rules

import (
	"time"

	"github.com/zalando/rds-health/internal/insight"
	"github.com/zalando/rds-health/internal/types"
)

// Estimator is container for generic algorithms and utility functions
// required to introduce a new rule.
type estimator struct {
	id   string // rule id
	name Metric // metric name (e.g. db.Transactions.xact_commit)
	unit string // metric measurement unit (e.g. iops)
	info string // short human readable description about metric
	desc string // long human readable description
}

func (est estimator) samplingInterval(samples insight.Samples) time.Duration {
	a := samples[0]
	b := samples[1]
	return b.T().Sub(a.T())
}

// utility function to show metric values
func (est estimator) ShowMinMax() ([]Metric, Eval) {
	return est.name.ToMinMax(), func(samples ...insight.Samples) types.Status {
		t := est.samplingInterval(samples[0])
		min, avg, max := samples[0].ToSeq(), samples[1].ToSeq(), samples[2].ToSeq()
		minmax := types.NewMinMax(min, avg, max)
		softminmax := types.NewMinMaxSoft(min, avg, max)

		return types.Status{
			Code:     types.STATUS_CODE_UNKNOWN,
			Rule:     types.Rule{Unit: est.unit, About: est.info},
			Interval: t,
			HardMM:   &minmax,
			SoftMM:   &softminmax,
		}
	}
}

// utility function to show metric values
func (est estimator) Show(stats Aggregator) ([]Metric, Eval) {
	return est.name.ToAgg(stats), func(samples ...insight.Samples) types.Status {
		t := est.samplingInterval(samples[0])
		pps := types.NewPercentile(samples[0].ToSeq())

		return types.Status{
			Code:       types.STATUS_CODE_UNKNOWN,
			Rule:       types.Rule{Unit: est.unit, About: est.info},
			Interval:   t,
			Aggregator: (*string)(&stats),
			Percentile: &pps,
		}
	}
}

// utility function to estimate that statistic is below the threshold
func (est estimator) Below(tAvg, tMax float64) ([]Metric, Eval) {
	return est.name.ToMinMax(), func(samples ...insight.Samples) types.Status {
		t := est.samplingInterval(samples[0])
		min, avg, max := samples[0].ToSeq(), samples[1].ToSeq(), samples[2].ToSeq()
		minmax := types.NewMinMaxSoft(min, avg, max)
		val := types.PercentileOf(avg, tAvg)

		status := types.STATUS_CODE_SUCCESS
		if minmax.Avg > tAvg {
			status = types.STATUS_CODE_WARNING
		}
		if minmax.Avg > tAvg && minmax.Max > tMax {
			status = types.STATUS_CODE_FAILURE
		}

		return types.Status{
			Code:        status,
			Rule:        types.Rule{ID: est.id, Unit: est.unit, About: est.info},
			Interval:    t,
			SuccessRate: &val,
			SoftMM:      &minmax,
		}
	}
}

// utility function to estimate that statistic is above the threshold
func (est estimator) Above(tMin, tAvg float64) ([]Metric, Eval) {
	return est.name.ToMinMax(), func(samples ...insight.Samples) types.Status {
		t := est.samplingInterval(samples[0])
		min, avg, max := samples[0].ToSeq(), samples[1].ToSeq(), samples[2].ToSeq()
		minmax := types.NewMinMaxSoft(min, avg, max)
		val := 100.0 - types.PercentileOf(avg, tAvg)

		status := types.STATUS_CODE_SUCCESS
		if minmax.Avg < tAvg {
			status = types.STATUS_CODE_WARNING
		}
		if minmax.Avg < tAvg && minmax.Min < tMin {
			status = types.STATUS_CODE_FAILURE
		}

		return types.Status{
			Code:        status,
			Rule:        types.Rule{ID: est.id, Unit: est.unit, About: est.info},
			Interval:    t,
			SuccessRate: &val,
			SoftMM:      &minmax,
		}
	}
}
