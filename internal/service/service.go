//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package service

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/pi"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/zalando/rds-health/internal/database"
	"github.com/zalando/rds-health/internal/discovery"
	"github.com/zalando/rds-health/internal/insight"
	"github.com/zalando/rds-health/internal/instance"
	"github.com/zalando/rds-health/internal/rules"
	"github.com/zalando/rds-health/internal/types"
)

type ProgressBar interface {
	Describe(string)
}

type Service struct {
	progress  ProgressBar
	database  *database.Database
	instance  *instance.Instance
	insight   *insight.Insight
	discovery *discovery.Discovery
}

func New(conf aws.Config, progress ProgressBar) *Service {
	rds := rds.NewFromConfig(conf)
	ec2 := ec2.NewFromConfig(conf)

	return &Service{
		progress:  progress,
		database:  database.New(rds),
		instance:  instance.New(ec2),
		insight:   insight.New(pi.NewFromConfig(conf)),
		discovery: discovery.New(rds, rds, ec2),
	}
}

//
//

func (service *Service) CheckHealthRegion(ctx context.Context, interval time.Duration) (*types.StatusRegion, error) {
	service.progress.Describe("discovering")

	clusters, nodes, err := service.discovery.LookupAll(context.Background())
	if err != nil {
		return nil, err
	}

	region := types.StatusRegion{
		Status:   types.STATUS_CODE_UNKNOWN,
		Clusters: make([]types.StatusCluster, len(clusters)),
		Nodes:    make([]types.StatusNode, len(nodes)),
	}

	for c := 0; c < len(clusters); c++ {
		cluster := clusters[c]
		status := types.StatusCluster{
			Status:  types.STATUS_CODE_UNKNOWN,
			Cluster: &cluster,
			Writer:  make([]types.StatusNode, len(cluster.Writer)),
			Reader:  make([]types.StatusNode, len(cluster.Reader)),
		}

		for w := 0; w < len(cluster.Writer); w++ {
			v, err := service.checkHealthNode(ctx, cluster.Writer[w], interval)
			if err != nil {
				return nil, err
			}
			status.Writer[w] = *v
			if status.Status < v.Status {
				status.Status = v.Status
			}
		}
		for r := 0; r < len(cluster.Reader); r++ {
			v, err := service.checkHealthNode(ctx, cluster.Reader[r], interval)
			if err != nil {
				return nil, err
			}
			status.Reader[r] = *v
			if status.Status < v.Status {
				status.Status = v.Status
			}
		}

		region.Clusters[c] = status
		if region.Status < status.Status {
			region.Status = status.Status
		}
	}

	for n := 0; n < len(nodes); n++ {
		v, err := service.checkHealthNode(ctx, nodes[n], interval)
		if err != nil {
			return nil, err
		}
		region.Nodes[n] = *v
		if region.Status < v.Status {
			region.Status = v.Status
		}
	}

	return &region, nil
}

func (service *Service) CheckHealthNode(ctx context.Context, name string, interval time.Duration) (*types.StatusNode, error) {
	service.progress.Describe("discovering " + name)

	node, err := service.database.Lookup(ctx, name)
	if err != nil {
		return nil, err
	}

	node.Compute, _ = service.instance.Lookup(context.Background(), node.Type)

	return service.checkHealthNode(ctx, *node, interval)
}

func (service *Service) checkHealthNode(ctx context.Context, node types.Node, interval time.Duration) (*types.StatusNode, error) {
	service.progress.Describe("checking " + node.Name)

	check := rules.New(service.insight).
		Should(rules.OsCpuUtil.Below(40.0, 60.0)).
		Should(rules.OsCpuWait.Below(8.0, 10.0)).
		Should(rules.OsSwapIn.Below(1.0, 1.0)).
		Should(rules.OsSwapOut.Below(1.0, 1.0)).
		Should(rules.DbStorageReadIO.Below(100.0, 300.0)).
		Should(rules.DbStorageWriteIO.Below(100.0, 300.0)).
		Should(rules.DbStorageAwait.Below(10.0, 20.0)).
		Should(rules.DbDataBlockCacheHitRatio.Above(80, 90)).
		Should(rules.DbDataBlockReadTime.Below(10.0, 20.0)).
		Should(rules.DbDeadlocks.Below(0.001, 0.01)).
		Should(rules.DbXactCommit.Above(3.0, 5.0)).
		Should(rules.SqlEfficiency.Above(10.0, 20.0))

	status, err := check.Run(ctx, node.ID, interval)
	if err != nil {
		return nil, err
	}

	code := types.STATUS_CODE_UNKNOWN
	for _, v := range status {
		if v.Code > code {
			code = v.Code
		}
	}

	return &types.StatusNode{
		Status: code,
		Node:   &node,
		Checks: status,
	}, nil
}

//
//

func (service *Service) ShowRegion(ctx context.Context) (*types.Region, error) {
	service.progress.Describe("discovering")

	clusters, nodes, err := service.discovery.LookupAll(context.Background())
	if err != nil {
		return nil, err
	}

	return &types.Region{
		Clusters: clusters,
		Nodes:    nodes,
	}, nil
}

//
//

func (service *Service) ShowNode(ctx context.Context, name string, interval time.Duration) (*types.StatusNode, error) {
	service.progress.Describe("checking " + name)

	db, err := service.database.Lookup(context.Background(), name)
	if err != nil {
		return nil, err
	}

	db.Compute, _ = service.instance.Lookup(context.Background(), db.Type)

	node := types.StatusNode{Node: db}
	node.Checks, err = rules.New(service.insight).
		Should(rules.DbXactCommit.ShowMinMax()).
		Should(rules.SqlTuplesFetched.ShowMinMax()).
		Should(rules.SqlTuplesReturned.ShowMinMax()).
		Should(rules.SqlTuplesInserted.ShowMinMax()).
		Should(rules.SqlTuplesUpdated.ShowMinMax()).
		Should(rules.SqlTuplesDeleted.ShowMinMax()).
		Should(rules.OsCpuUtil.ShowMinMax()).
		Should(rules.OsCpuWait.ShowMinMax()).
		Should(rules.DbStorageReadIO.ShowMinMax()).
		Should(rules.DbStorageWriteIO.ShowMinMax()).
		Should(rules.DbDataBlockReadIO.ShowMinMax()).
		Should(rules.DbDataBlockCacheHit.ShowMinMax()).
		Should(rules.DbBuffersCheckpoints.ShowMinMax()).
		Should(rules.DbBuffersCheckpointsTime.ShowMinMax()).
		Should(rules.OsMemoryFree.ShowMinMax()).
		Should(rules.OsMemoryCached.ShowMinMax()).
		Should(rules.OsFileSysUsed.ShowMinMax()).
		Run(context.Background(), db.ID, interval)

	if err != nil {
		return nil, err
	}

	return &node, nil
}
