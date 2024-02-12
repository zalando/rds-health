//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package types

import (
	"fmt"
	"strings"
)

//
// Common domain types used by the application
//

// Binary Storage Unit
type BiB uint

const (
	KiB = BiB(1024)
	MiB = BiB(1024 * 1024)
	GiB = BiB(1024 * 1024 * 1024)
	TiB = BiB(1024 * 1024 * 1024 * 1024)
)

func (v BiB) String() string {
	switch {
	case v >= TiB:
		return fmt.Sprintf("%d TiB", v/TiB)
	case v >= GiB:
		return fmt.Sprintf("%d GiB", v/GiB)
	case v >= MiB:
		return fmt.Sprintf("%d MiB", v/MiB)
	case v >= KiB:
		return fmt.Sprintf("%d KiB", v/KiB)
	default:
		return fmt.Sprintf("%d bytes", v)
	}
}

// Frequency data type
type GHz float64

func (v GHz) String() string {
	return fmt.Sprintf("%.2f GHz", v)
}

// Storage specification
type Storage struct {
	Type string `json:"type"`
	Size BiB    `json:"size"`
}

func (v Storage) String() string {
	if v.Type == "memory" {
		return fmt.Sprintf("mem %s", v.Size)
	}

	return fmt.Sprintf("storage %s %s", v.Type, v.Size)
}

// CPU specification
type CPU struct {
	Cores int `json:"cores"`
	Clock GHz `json:"clock"`
}

func (v CPU) String() string {
	return fmt.Sprintf("%d vcpu %s", v.Cores, v.Clock)
}

// Compute resource
type Compute struct {
	CPU    *CPU     `json:"cpu,omitempty"`
	Memory *Storage `json:"memory,omitempty"`
}

func (v Compute) String() string {
	spec := []string{}
	if v.CPU != nil {
		spec = append(spec, v.CPU.String())
	}

	if v.Memory != nil {
		spec = append(spec, v.Memory.String())
	}

	return strings.Join(spec, ", ")
}

// Availability Zones Node is deployed to
type AvailabilityZones []string

func (v AvailabilityZones) String() string {
	return strings.Join(v, ", ")
}

// Database engine specification
type Engine struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

func (v Engine) String() string {
	return fmt.Sprintf("%s v%s", v.ID, v.Version)
}

// Cluster Node
type Node struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Zones    AvailabilityZones `json:"zones"`
	Engine   *Engine           `json:"engine,omitempty"`
	Storage  *Storage          `json:"storage,omitempty"`
	Compute  *Compute          `json:"compute,omitempty"`
	ReadOnly bool              `json:"readonly"`
}

func (v Node) String() string {
	engine := ""
	if v.Engine != nil {
		engine = " " + v.Engine.String()
	}

	spec := []string{}
	if v.Compute != nil {
		spec = append(spec, v.Compute.String())
	}

	if v.Storage != nil {
		spec = append(spec, v.Storage.String())
	}

	conf := ""
	if len(spec) > 0 {
		conf = " (" + strings.Join(spec, ", ") + ")"
	}

	return v.Type + engine + conf
}

// DB cluster topology
type Cluster struct {
	ID     string  `json:"id"`
	Engine *Engine `json:"engine,omitempty"`
	Reader []Node  `json:"reader,omitempty"`
	Writer []Node  `json:"writer,omitempty"`
}

// Region topology
type Region struct {
	Clusters []Cluster
	Nodes    []Node
}
