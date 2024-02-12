//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package insight

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pi/types"
)

// Sample abstract the dependencies to AWS types.DataPoint
type Sample interface {
	T() time.Time
	X() float64
}

// Samples is a time series sequence
type Samples []Sample

func (samples Samples) ToSeq() []float64 {
	seq := make([]float64, len(samples))
	for i, val := range samples {
		seq[i] = val.X()
	}
	return seq
}

type sample types.DataPoint

func (v sample) T() time.Time { return aws.ToTime(v.Timestamp) }
func (v sample) X() float64   { return aws.ToFloat64(v.Value) }
