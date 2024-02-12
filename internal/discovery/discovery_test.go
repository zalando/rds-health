//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package discovery_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/zalando/rds-health/internal/discovery"
	"github.com/zalando/rds-health/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestLookupAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//
	dbs := &rds.DescribeDBInstancesOutput{
		DBInstances: []rdstypes.DBInstance{
			database("a"),
			database("e"),
			database("b"),
			database("c"),
			database("d"),
		},
	}
	databases := mocks.NewDatabase(ctrl)
	databases.EXPECT().DescribeDBInstances(gomock.Any(), gomock.Any()).Return(dbs, nil)

	//
	cls := &rds.DescribeDBClustersOutput{
		DBClusters: []rdstypes.DBCluster{
			cluster("B", "b", ""),
			cluster("A", "a", "d"),
		},
	}
	clusters := mocks.NewCluster(ctrl)
	clusters.EXPECT().DescribeDBClusters(gomock.Any(), gomock.Any()).Return(cls, nil)

	//
	its := &ec2.DescribeInstanceTypesOutput{
		InstanceTypes: []ec2types.InstanceTypeInfo{},
	}
	instances := mocks.NewInstance(ctrl)
	instances.EXPECT().DescribeInstanceTypes(gomock.Any(), gomock.Any()).Return(its, nil)

	sut := discovery.New(clusters, databases, instances)

	c, n, err := sut.LookupAll(context.Background())
	switch {
	case err != nil:
		t.Errorf("should not failed with error %s", err)
	case len(c) != 2:
		t.Errorf("should return clusters")

	case c[0].ID != "A":
		t.Errorf("should not return unexpected value of 1st cluster |%s|", c[0].ID)
	case len(c[0].Writer) != 1:
		t.Errorf("should have writer nodes at 1st cluster")
	case c[0].Writer[0].Name != "a":
		t.Errorf("unexpected writer node at 1st cluster |%s|", c[0].Writer[0].Name)
	case len(c[0].Reader) != 1:
		t.Errorf("should have reader nodes at 1st cluster")
	case c[0].Reader[0].Name != "d":
		t.Errorf("unexpected reader node at 1st cluster |%s|", c[0].Reader[0].Name)

	case c[1].ID != "B":
		t.Errorf("should not return unexpected value of 2nd cluster |%s|", c[1].ID)
	case len(c[1].Writer) != 1:
		t.Errorf("should have writer nodes at 2nd cluster")
	case c[1].Writer[0].Name != "b":
		t.Errorf("unexpected writer node at 2nd cluster |%s|", c[1].Writer[0].Name)
	case len(c[1].Reader) != 0:
		t.Errorf("should not have reader nodes at 2nd cluster")

	case len(n) != 2:
		t.Errorf("should return databases")
	case n[0].Name != "c":
		t.Errorf("should not return unexpected value of 1st database |%s|", n[0].Name)
	case n[1].Name != "e":
		t.Errorf("should not return unexpected value of 2nd database |%s|", n[1].Name)
	}
}

//
// Helper
//

// mock database
func database(name string) rdstypes.DBInstance {
	return rdstypes.DBInstance{
		DBInstanceIdentifier:      aws.String(name),
		DBInstanceClass:           aws.String("db.t2.small"),
		Engine:                    aws.String("postgres"),
		EngineVersion:             aws.String("13.14"),
		StorageType:               aws.String("gp2"),
		AllocatedStorage:          100,
		AvailabilityZone:          aws.String("eu-central-1a"),
		SecondaryAvailabilityZone: nil,
	}
}

// mock cluster
func cluster(name string, writer string, reader string) rdstypes.DBCluster {
	members := []rdstypes.DBClusterMember{}

	if writer != "" {
		members = append(members, rdstypes.DBClusterMember{
			DBInstanceIdentifier: aws.String(writer),
			IsClusterWriter:      true,
		})
	}

	if reader != "" {
		members = append(members, rdstypes.DBClusterMember{
			DBInstanceIdentifier: aws.String(reader),
			IsClusterWriter:      false,
		})
	}

	return rdstypes.DBCluster{
		DBClusterIdentifier: aws.String(name),
		Engine:              aws.String("postgres"),
		EngineVersion:       aws.String("13.14"),
		DBClusterMembers:    members,
	}
}
