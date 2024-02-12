//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package database_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/zalando/rds-health/internal/database"
	"github.com/zalando/rds-health/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestLookupAllSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fix := &rds.DescribeDBInstancesOutput{
		DBInstances: []rdstypes.DBInstance{
			{
				DBInstanceIdentifier:      aws.String("test-db"),
				DBInstanceClass:           aws.String("db.t2.small"),
				Engine:                    aws.String("postgres"),
				EngineVersion:             aws.String("13.14"),
				StorageType:               aws.String(""),
				AllocatedStorage:          100,
				AvailabilityZone:          aws.String("eu-central-1a"),
				SecondaryAvailabilityZone: nil,
			},
		},
	}

	mock := mocks.NewDatabase(ctrl)
	mock.EXPECT().DescribeDBInstances(gomock.Any(), gomock.Any()).Return(fix, nil)

	sut := database.New(mock)

	seq, err := sut.LookupAll(context.TODO())
	switch {
	case err != nil:
		t.Errorf("should not failed with error %s", err)
	case len(seq) == 0:
		t.Errorf("should return db instances")
	case seq[0].String() != "db.t2.small postgres v13.14 (storage  100 GiB)":
		t.Errorf("should not return unexpected value |%s|", seq[0])
	}
}

func TestLookupSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fix := &rds.DescribeDBInstancesOutput{
		DBInstances: []rdstypes.DBInstance{
			{
				DBInstanceIdentifier:      aws.String("test-db"),
				DBInstanceClass:           aws.String("db.t2.small"),
				Engine:                    aws.String("postgres"),
				EngineVersion:             aws.String("13.14"),
				StorageType:               aws.String(""),
				AllocatedStorage:          100,
				AvailabilityZone:          aws.String("eu-central-1a"),
				SecondaryAvailabilityZone: nil,
			},
		},
	}

	mock := mocks.NewDatabase(ctrl)
	mock.EXPECT().DescribeDBInstances(gomock.Any(), gomock.Any()).Return(fix, nil)

	sut := database.New(mock)

	db, err := sut.Lookup(context.TODO(), "test-db")
	switch {
	case err != nil:
		t.Errorf("should not failed with error %s", err)
	case db == nil:
		t.Errorf("should return db instances")
	case db.String() != "db.t2.small postgres v13.14 (storage  100 GiB)":
		t.Errorf("should not return unexpected value |%s|", db)
	}
}
