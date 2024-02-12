//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package instance

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/zalando/rds-health/internal/types"
)

//go:generate mockgen -destination=../mocks/instance.go -package=mocks -mock_names Provider=Instance . Provider
type Provider interface {
	DescribeInstanceTypes(
		context.Context,
		*ec2.DescribeInstanceTypesInput,
		...func(*ec2.Options),
	) (*ec2.DescribeInstanceTypesOutput, error)
}

type Instance struct {
	provider Provider
}

func New(provider Provider) *Instance {
	return &Instance{provider: provider}
}

// Lookup metadata about the database
func (in *Instance) Lookup(ctx context.Context, dbInstanceType string) (*types.Compute, error) {
	dbInstanceType = strings.TrimPrefix(dbInstanceType, "db.")

	spec, err := in.provider.DescribeInstanceTypes(ctx,
		&ec2.DescribeInstanceTypesInput{
			InstanceTypes: []ec2types.InstanceType{
				ec2types.InstanceType(dbInstanceType),
			},
		},
	)
	if err != nil {
		return nil, err
	}

	if len(spec.InstanceTypes) != 1 {
		return nil, nil
	}

	instance := spec.InstanceTypes[0]

	compute := types.Compute{}

	if mem := instance.MemoryInfo; mem != nil {
		compute.Memory = &types.Storage{
			Type: "memory",
			Size: types.BiB(aws.ToInt64(mem.SizeInMiB)) * types.MiB,
		}
	}

	if cpu := instance.VCpuInfo; cpu != nil {
		compute.CPU = &types.CPU{
			Cores: int(aws.ToInt32(cpu.DefaultVCpus)),
		}
	}

	if proc := instance.ProcessorInfo; proc != nil && compute.CPU != nil {
		compute.CPU.Clock = types.GHz(aws.ToFloat64(proc.SustainedClockSpeedInGhz))
	}

	return &compute, nil
}
