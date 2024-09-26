//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package cluster_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/zalando/rds-health/internal/cluster"
	"github.com/zalando/rds-health/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestLookup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fix := &rds.DescribeDBClustersOutput{
		DBClusters: []rdstypes.DBCluster{
			{
				DBClusterIdentifier: aws.String("test-db"),
				Engine:              aws.String("postgres"),
				EngineVersion:       aws.String("13.14"),
				DBClusterMembers: []rdstypes.DBClusterMember{
					{
						DBInstanceIdentifier: aws.String("test-1"),
						IsClusterWriter:      aws.Bool(true),
					},
					{
						DBInstanceIdentifier: aws.String("test-2"),
						IsClusterWriter:      aws.Bool(false),
					},
				},
			},
		},
	}

	mock := mocks.NewCluster(ctrl)
	mock.EXPECT().DescribeDBClusters(gomock.Any(), gomock.Any()).Return(fix, nil)

	sut := cluster.New(mock)

	seq, err := sut.Lookup(context.TODO())
	switch {
	case err != nil:
		t.Errorf("should not failed with error %s", err)
	case len(seq) == 0:
		t.Errorf("should return db instances")
	case seq[0].ID != "test-db":
		t.Errorf("should not return unexpected value |%s|", seq[0])
	case len(seq[0].Writer) == 0:
		t.Errorf("should have writer nodes")
	case len(seq[0].Reader) == 0:
		t.Errorf("should have reader nodes")
	}
}
