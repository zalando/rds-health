//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/zalando/rds-health/internal/types"
)

//go:generate mockgen -destination=../mocks/database.go -package=mocks -mock_names Provider=Database . Provider
type Provider interface {
	DescribeDBInstances(
		context.Context,
		*rds.DescribeDBInstancesInput,
		...func(*rds.Options),
	) (*rds.DescribeDBInstancesOutput, error)
}

type Database struct {
	provider Provider
}

func New(provider Provider) *Database {
	return &Database{provider: provider}
}

func (db *Database) LookupAll(ctx context.Context) ([]types.Node, error) {
	clusters := make([]types.Node, 0)

	var cursor *string
	for do := true; do; do = cursor != nil {
		bag, err := db.provider.DescribeDBInstances(ctx,
			&rds.DescribeDBInstancesInput{
				Marker: cursor,
			},
		)
		if err != nil {
			return nil, err
		}

		for _, c := range bag.DBInstances {
			clusters = append(clusters, db.toNode(c))
		}
		cursor = bag.Marker
	}

	return clusters, nil
}

// Lookup database ID using human-friendly name
func (db *Database) Lookup(ctx context.Context, name string) (*types.Node, error) {
	val, err := db.provider.DescribeDBInstances(ctx,
		&rds.DescribeDBInstancesInput{DBInstanceIdentifier: &name},
	)
	if err != nil {
		return nil, err
	}

	if len(val.DBInstances) == 0 {
		return nil, fmt.Errorf("not found: rds %s", name)
	}

	node := db.toNode(val.DBInstances[0])
	return &node, nil
}

func (db *Database) toNode(instance rdstypes.DBInstance) types.Node {
	engine := types.Engine{
		ID:      aws.ToString(instance.Engine),
		Version: aws.ToString(instance.EngineVersion),
	}

	storage := types.Storage{
		Type: aws.ToString(instance.StorageType),
		Size: types.BiB(instance.AllocatedStorage) * types.GiB,
	}

	az := types.AvailabilityZones{}
	if instance.AvailabilityZone != nil {
		az = append(az, aws.ToString(instance.AvailabilityZone))
	}

	if instance.SecondaryAvailabilityZone != nil {
		az = append(az, aws.ToString(instance.SecondaryAvailabilityZone))
	}

	// instance.DbiResourceId

	node := types.Node{
		ID:      aws.ToString(instance.DbiResourceId),
		Name:    aws.ToString(instance.DBInstanceIdentifier),
		Type:    aws.ToString(instance.DBInstanceClass),
		Zones:   az,
		Engine:  &engine,
		Storage: &storage,
	}

	return node
}
