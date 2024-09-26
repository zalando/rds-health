//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package cluster

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/zalando/rds-health/internal/types"
)

//go:generate mockgen -destination=../mocks/cluster.go -package=mocks -mock_names Provider=Cluster . Provider
type Provider interface {
	DescribeDBClusters(
		context.Context,
		*rds.DescribeDBClustersInput,
		...func(*rds.Options),
	) (*rds.DescribeDBClustersOutput, error)
}

type Cluster struct {
	provider Provider
}

func New(provider Provider) *Cluster {
	return &Cluster{provider: provider}
}

func (api Cluster) Lookup(ctx context.Context) ([]types.Cluster, error) {
	clusters := make([]types.Cluster, 0)

	var cursor *string
	for do := true; do; do = cursor != nil {
		bag, err := api.provider.DescribeDBClusters(ctx,
			&rds.DescribeDBClustersInput{
				Marker: cursor,
			},
		)
		if err != nil {
			return nil, err
		}

		for _, c := range bag.DBClusters {
			clusters = append(clusters, api.toCluster(c))
		}
		cursor = bag.Marker
	}

	return clusters, nil
}

func (api Cluster) toCluster(c rdstypes.DBCluster) types.Cluster {
	cluster := types.Cluster{
		ID:     aws.ToString(c.DBClusterIdentifier),
		Reader: make([]types.Node, 0),
		Writer: make([]types.Node, 0),
	}

	if aws.ToString(c.Engine) != "" {
		cluster.Engine = &types.Engine{
			ID:      aws.ToString(c.Engine),
			Version: aws.ToString(c.EngineVersion),
		}
	}

	for _, member := range c.DBClusterMembers {
		node := types.Node{
			Name: aws.ToString(member.DBInstanceIdentifier),
		}
		if aws.ToBool(member.IsClusterWriter) {
			cluster.Writer = append(cluster.Writer, node)
		} else {
			cluster.Reader = append(cluster.Reader, node)
		}
	}

	return cluster
}
