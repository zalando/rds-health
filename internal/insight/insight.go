//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package insight

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pi"
	"github.com/aws/aws-sdk-go-v2/service/pi/types"
)

//go:generate mockgen -destination=../mocks/insight.go -mock_names Provider=Insight -package=mocks . Provider
type Provider interface {
	GetResourceMetrics(
		context.Context,
		*pi.GetResourceMetricsInput,
		...func(*pi.Options),
	) (*pi.GetResourceMetricsOutput, error)
}

type Insight struct {
	api     Provider
	service *string
}

func New(provider Provider) *Insight {
	return &Insight{
		api:     provider,
		service: aws.String("RDS"),
	}
}

func (in *Insight) periodInSeconds(dur time.Duration) int32 {
	// Note: Valid values are: 1, 60, 300, 3600, 86400
	switch {
	case dur <= 10*time.Minute:
		return 1
	case dur <= 5*time.Hour:
		return 60
	case dur <= 24*time.Hour:
		return 300
	default:
		return 3600
	}
}

func (in *Insight) Fetch(ctx context.Context, dbiResourceId string, dur time.Duration, metrics ...string) (map[string]Samples, error) {
	var chunks [][]string
	chunkSize := 15
	for i := 0; i < len(metrics); i += chunkSize {
		end := i + chunkSize

		if end > len(metrics) {
			end = len(metrics)
		}

		chunks = append(chunks, metrics[i:end])
	}

	childContext, cancel := context.WithCancel(ctx)
	defer cancel()

	samples := map[string]Samples{}
	var wg sync.WaitGroup
	var err error
	var mu sync.Mutex

	for _, chunk := range chunks {
		chunk := chunk
		wg.Add(1)

		go func() {
			defer wg.Done()

			set, e := in.fetch(childContext, dbiResourceId, dur, chunk...)

			mu.Lock()
			defer mu.Unlock()

			if e != nil {
				if err == nil {
					cancel()
					err = e
				}
				return
			}

			for k, v := range set {
				samples[k] = v
			}
		}()
	}
	wg.Wait()
	if err != nil {
		return nil, err
	}

	return samples, nil
}

func (in *Insight) fetch(ctx context.Context, dbiResourceId string, dur time.Duration, metrics ...string) (map[string]Samples, error) {
	query := make([]types.MetricQuery, 0, len(metrics))
	for _, metric := range metrics {
		query = append(query,
			types.MetricQuery{Metric: aws.String(metric)},
		)
	}

	period := in.periodInSeconds(dur)
	req := pi.GetResourceMetricsInput{
		ServiceType:     types.ServiceType(*in.service),
		Identifier:      aws.String(dbiResourceId),
		StartTime:       aws.Time(time.Now().Add(-dur)),
		EndTime:         aws.Time(time.Now()),
		PeriodInSeconds: aws.Int32(period),
		MetricQueries:   query,
	}

	ret, err := in.api.GetResourceMetrics(ctx, &req)
	if err != nil {
		return nil, err
	}

	series := make(map[string]Samples)
	for _, metric := range ret.MetricList {
		seq := make(Samples, len(metric.DataPoints))
		for i, v := range metric.DataPoints {
			seq[i] = sample(v)
		}
		series[aws.ToString(metric.Key.Metric)] = seq
	}

	return series, nil
}
