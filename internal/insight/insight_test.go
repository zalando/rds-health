//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package insight_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pi"
	"github.com/aws/aws-sdk-go-v2/service/pi/types"
	"github.com/zalando/rds-health/internal/insight"
	"github.com/zalando/rds-health/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestFetch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fixKey := "db.cpu.avg"
	fixRet := &pi.GetResourceMetricsOutput{
		MetricList: []types.MetricKeyDataPoints{
			{
				Key: &types.ResponseResourceMetricKey{Metric: aws.String(fixKey)},
				DataPoints: []types.DataPoint{
					{Timestamp: aws.Time(time.Now()), Value: aws.Float64(100.0)},
				},
			},
		},
	}

	mock := mocks.NewInsight(ctrl)
	mock.EXPECT().GetResourceMetrics(gomock.Any(), gomock.Any()).Return(fixRet, nil)

	sut := insight.New(mock)

	samples, err := sut.Fetch(context.TODO(), "db-XXXXXXXXXXXXXXXXXXXXXXXXXX", 60*time.Minute, fixKey)
	switch {
	case err != nil:
		t.Errorf("should not fail with error %s", err)
	case len(samples) != 1:
		t.Errorf("should not return multiple metrics %v", samples)
	case samples[fixKey] == nil:
		t.Errorf("should %v contain samples for %v", samples, fixKey)
	case len(samples[fixKey]) != 1:
		t.Errorf("should return %v samples of length 1", samples[fixKey])
	case samples[fixKey][0].X() != *fixRet.MetricList[0].DataPoints[0].Value:
		t.Errorf("should return %v samples with first value %v", samples[fixKey], *fixRet.MetricList[0].DataPoints[0].Value)
	}
}
