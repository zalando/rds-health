//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package instance_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/zalando/rds-health/internal/instance"
	"github.com/zalando/rds-health/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestLookupSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fix := &ec2.DescribeInstanceTypesOutput{
		InstanceTypes: []ec2types.InstanceTypeInfo{
			{
				MemoryInfo:    &ec2types.MemoryInfo{SizeInMiB: aws.Int64(4 * 1024)},
				VCpuInfo:      &ec2types.VCpuInfo{DefaultVCpus: aws.Int32(2)},
				ProcessorInfo: &ec2types.ProcessorInfo{SustainedClockSpeedInGhz: aws.Float64(2.20)},
			},
		},
	}

	mock := mocks.NewInstance(ctrl)
	mock.EXPECT().DescribeInstanceTypes(gomock.Any(), gomock.Any()).Return(fix, nil)

	sut := instance.New(mock)

	compute, err := sut.Lookup(context.TODO(), "db.t2.small")
	switch {
	case err != nil:
		t.Errorf("should not failed with error %s", err)
	case compute == nil:
		t.Errorf("should not return nil")
	case compute.String() != "2 vcpu 2.20 GHz, mem 4 GiB":
		t.Errorf("should not return unexpected value |%s|", compute)
	}
}

func TestLookupNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fix := &ec2.DescribeInstanceTypesOutput{
		InstanceTypes: []ec2types.InstanceTypeInfo{},
	}

	mock := mocks.NewInstance(ctrl)
	mock.EXPECT().DescribeInstanceTypes(gomock.Any(), gomock.Any()).Return(fix, nil)

	sut := instance.New(mock)

	compute, err := sut.Lookup(context.TODO(), "db.t2.small")
	switch {
	case err != nil:
		t.Errorf("should not failed with error %s", err)
	case compute != nil:
		t.Errorf("should return nil")
	}
}
