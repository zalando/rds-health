//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package rules

import (
	"context"
	"fmt"
	"time"

	"github.com/zalando/rds-health/internal/insight"
	"github.com/zalando/rds-health/internal/types"
)

type Metric string

func (m Metric) ToMin() []Metric { return []Metric{m + ".min"} }
func (m Metric) ToAvg() []Metric { return []Metric{m + ".avg"} }
func (m Metric) ToMax() []Metric { return []Metric{m + ".max"} }
func (m Metric) ToSum() []Metric { return []Metric{m + ".sum"} }

func (m Metric) ToAgg(agg Aggregator) []Metric {
	return []Metric{m + "." + Metric(agg)}
}

func (m Metric) ToMinMax() []Metric {
	return append(append(m.ToMin(), m.ToAvg()...), m.ToMax()...)
}

// Aggregator function used by telemetry system
//
// The following statistic aggregators are supported for the metrics:
//
// `.avg` - The average value for the metric over a period of time.
//
// `.min` - The minimum value for the metric over a period of time.
//
// `.max` - The maximum value for the metric over a period of time.
//
// `.sum` - The sum of the metric values over a period of time.
//
// `.sample_count` - The number of times the metric was collected over a period of time. Append  to the metric name.
//
// See https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_PerfInsights.API.html
type Aggregator string

const (
	STATS_SUM = Aggregator("sum")
	STATS_AVG = Aggregator("avg")
	STATS_MIN = Aggregator("min")
	STATS_MAX = Aggregator("max")
)

type Eval func(...insight.Samples) types.Status
type Rule func() ([]Metric, Eval)

type Source interface {
	Fetch(context.Context, string, time.Duration, ...string) (map[string]insight.Samples, error)
}

type Check struct {
	source  Source
	metrics []Metric
	related map[Metric][]Metric
	should  map[Metric]Eval
	index   []Metric
}

func New(source Source) *Check {
	return &Check{
		source:  source,
		metrics: []Metric{},
		related: map[Metric][]Metric{},
		should:  map[Metric]Eval{},
		index:   []Metric{},
	}
}

func (check *Check) Should(metrics []Metric, eval Eval) *Check {
	root := metrics[0]
	check.related[root] = metrics
	check.metrics = append(check.metrics, metrics...)
	check.should[root] = eval
	check.index = append(check.index, root)
	return check
}

func (check *Check) Run(ctx context.Context, dbiResourceId string, dur time.Duration) ([]types.Status, error) {
	seqToFetch := make([]string, len(check.metrics))
	for i, metric := range check.metrics {
		seqToFetch[i] = string(metric)
	}

	samples, err := check.source.Fetch(ctx, dbiResourceId, dur, seqToFetch...)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch samples", err)
	}

	status := make([]types.Status, 0)
	for _, root := range check.index {
		eval := check.should[root]
		seqOfSamples := make([]insight.Samples, 0)
		for _, related := range check.related[root] {
			seqOfSamples = append(seqOfSamples, samples[string(related)])
		}

		status = append(status, eval(seqOfSamples...))
	}

	return status, nil
}
