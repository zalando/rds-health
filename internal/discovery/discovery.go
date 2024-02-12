//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package discovery

import (
	"context"
	"sort"

	"github.com/zalando/rds-health/internal/cache"
	"github.com/zalando/rds-health/internal/cluster"
	"github.com/zalando/rds-health/internal/database"
	"github.com/zalando/rds-health/internal/instance"
	"github.com/zalando/rds-health/internal/types"
)

//
// Discovery of database instances in the cluster
//

type Discovery struct {
	cluster  *cluster.Cluster
	database *database.Database
	instance *cache.Cache[string, *types.Compute]
}

func New(
	cp cluster.Provider,
	dp database.Provider,
	ip instance.Provider,
) *Discovery {
	return &Discovery{
		cluster:  cluster.New(cp),
		database: database.New(dp),
		instance: cache.New(instance.New(ip)),
	}
}

func (service *Discovery) LookupAll(ctx context.Context) ([]types.Cluster, []types.Node, error) {
	allNodes, err := service.database.LookupAll(ctx)
	if err != nil {
		return nil, nil, err
	}

	mapNodes := make(map[string]types.Node)
	for i := 0; i < len(allNodes); i++ {
		node := allNodes[i]
		node.Compute, _ = service.instance.Lookup(context.Background(), node.Type)
		mapNodes[node.Name] = node
	}

	clusters, err := service.cluster.Lookup(context.Background())
	if err != nil {
		return nil, nil, err
	}

	for i := 0; i < len(clusters); i++ {
		writers := make([]types.Node, len(clusters[i].Writer))
		for k, w := range clusters[i].Writer {
			if node, has := mapNodes[w.Name]; has {
				writers[k] = node
				delete(mapNodes, w.Name)
			}
		}
		clusters[i].Writer = writers

		readers := make([]types.Node, len(clusters[i].Reader))
		for k, r := range clusters[i].Reader {
			if node, has := mapNodes[r.Name]; has {
				node.ReadOnly = true
				readers[k] = node
				delete(mapNodes, r.Name)
			}
		}
		clusters[i].Reader = readers
	}

	nodes := make([]types.Node, 0, len(mapNodes))
	for _, node := range mapNodes {
		nodes = append(nodes, node)
	}

	sort.SliceStable(clusters, func(i, j int) bool { return clusters[i].ID < clusters[j].ID })
	sort.SliceStable(nodes, func(i, j int) bool { return nodes[i].Name < nodes[j].Name })
	return clusters, nodes, nil
}
