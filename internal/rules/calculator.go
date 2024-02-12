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

// Calculator is container for generic algorithms and utility functions
// required to introduce a new "calculate-able" rules.
type calculator struct {
	id   string // rule id
	lhm  Metric // left hand metric
	rhm  Metric // right hand metric
	fop  func(float64, float64) float64
	unit string // metric measurement unit (e.g. iops)
	info string // short human readable description about metric
	desc string // long human readable description
}

func (cal calculator) samplingInterval(samples insight.Samples) time.Duration {
	a := samples[0]
	b := samples[1]
	return b.T().Sub(a.T())
}

func (cal calculator) apply(lhm, rhm []float64) []float64 {
	seq := make([]float64, len(lhm))
	for i := 0; i < len(lhm); i++ {
		seq[i] = cal.fop(lhm[i], rhm[i])
	}
	return seq
}

// utility function to show metric values
func (cal calculator) ShowMinMax() ([]Metric, Eval) {
	return append(cal.lhm.ToMinMax(), cal.rhm.ToMinMax()...), func(samples ...insight.Samples) types.Status {
		t := cal.samplingInterval(samples[0])

		lmin, lavg, lmax := samples[0].ToSeq(), samples[1].ToSeq(), samples[2].ToSeq()
		rmin, ravg, rmax := samples[3].ToSeq(), samples[4].ToSeq(), samples[5].ToSeq()

		min := cal.apply(lmin, rmin)
		avg := cal.apply(lavg, ravg)
		max := cal.apply(lmax, rmax)

		minmax := types.NewMinMax(min, avg, max)
		softminmax := types.NewMinMaxSoft(min, avg, max)

		return types.Status{
			Code:     types.STATUS_CODE_UNKNOWN,
			Rule:     types.Rule{Unit: cal.unit, About: cal.info},
			Interval: t,
			HardMM:   &minmax,
			SoftMM:   &softminmax,
		}
	}
}

// utility function to estimate that statistic is below the threshold
func (cal calculator) Below(tAvg, tMax float64) ([]Metric, Eval) {
	return append(cal.lhm.ToMinMax(), cal.rhm.ToMinMax()...), func(samples ...insight.Samples) types.Status {
		t := cal.samplingInterval(samples[0])
		lmin, lavg, lmax := samples[0].ToSeq(), samples[1].ToSeq(), samples[2].ToSeq()
		rmin, ravg, rmax := samples[3].ToSeq(), samples[4].ToSeq(), samples[5].ToSeq()

		min := cal.apply(lmin, rmin)
		avg := cal.apply(lavg, ravg)
		max := cal.apply(lmax, rmax)

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
			Rule:        types.Rule{ID: cal.id, Unit: cal.unit, About: cal.info},
			Interval:    t,
			SuccessRate: &val,
			SoftMM:      &minmax,
		}
	}
}

// utility function to estimate that statistic is above the threshold
func (cal calculator) Above(tMin, tAvg float64) ([]Metric, Eval) {
	return append(cal.lhm.ToMinMax(), cal.rhm.ToMinMax()...), func(samples ...insight.Samples) types.Status {
		t := cal.samplingInterval(samples[0])
		lmin, lavg, lmax := samples[0].ToSeq(), samples[1].ToSeq(), samples[2].ToSeq()
		rmin, ravg, rmax := samples[3].ToSeq(), samples[4].ToSeq(), samples[5].ToSeq()

		min := cal.apply(lmin, rmin)
		avg := cal.apply(lavg, ravg)
		max := cal.apply(lmax, rmax)

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
			Rule:        types.Rule{ID: cal.id, Unit: cal.unit, About: cal.info},
			Interval:    t,
			SuccessRate: &val,
			SoftMM:      &minmax,
		}
	}
}
