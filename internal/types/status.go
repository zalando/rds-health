//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

//
//

type Rule struct {
	ID    string `json:"id,omitempty"`
	Unit  string `json:"unit,omitempty"`
	About string `json:"about,omitempty"`
}

func (v Rule) String() string {
	if v.ID == "" {
		return fmt.Sprintf("%-37s", v.About)
	}

	return fmt.Sprintf("%s: %-37s", v.ID, v.About)
}

//
//

type StatusCode int

func (v StatusCode) String() string {
	switch v {
	case STATUS_CODE_UNKNOWN:
		return "UNKNOWN"
	case STATUS_CODE_SUCCESS:
		return "PASSED"
	case STATUS_CODE_WARNING:
		return "WARNED"
	case STATUS_CODE_FAILURE:
		return "FAILED"
	default:
		panic(fmt.Errorf("status code %d unknown", v))
	}
}

const (
	STATUS_CODE_UNKNOWN StatusCode = iota
	STATUS_CODE_SUCCESS
	STATUS_CODE_WARNING
	STATUS_CODE_FAILURE
)

// helper formatter for colored output
func (code StatusCode) sprintf(m string) string {
	switch code {
	case STATUS_CODE_UNKNOWN:
		return "\033[32m" + m + "\033[0m"
	case STATUS_CODE_SUCCESS:
		return "\033[32m PASSED: " + m + "\033[0m"
	case STATUS_CODE_WARNING:
		return "\033[33m WARNED: " + m + "\033[0m"
	case STATUS_CODE_FAILURE:
		return "\033[31m FAILED: " + m + "\033[0m"
	default:
		return m
	}
}

func (code StatusCode) MarshalJSON() ([]byte, error) {
	switch code {
	case STATUS_CODE_UNKNOWN:
		return json.Marshal("unknown")
	case STATUS_CODE_SUCCESS:
		return json.Marshal("passed")
	case STATUS_CODE_WARNING:
		return json.Marshal("warned")
	case STATUS_CODE_FAILURE:
		return json.Marshal("failed")
	default:
		return nil, fmt.Errorf("status code %d unknown to JSON codec", code)
	}
}

//
//

// Status of rule evaluation
type Status struct {
	Code        StatusCode    `json:"status"`
	Rule        Rule          `json:"rule"`
	Interval    time.Duration `json:"-"`
	SuccessRate *float64      `json:"success_rate,omitempty"`
	HardMM      *MinMax       `json:"hard_minmax,omitempty"`
	SoftMM      *MinMax       `json:"soft_minmax,omitempty"`
	Aggregator  *string       `json:"aggregator,omitempty"`
	Percentile  *Percentile   `json:"distribution,omitempty"`
}

func (v Status) String() string {
	// Note: special formatting for percentiles
	if v.Percentile != nil && v.Aggregator != nil {
		return fmt.Sprintf("%-37s | %s %s %4s %s ", v.Rule.About, v.Interval, *v.Aggregator, v.Rule.Unit, v.Percentile)
	}

	seq := make([]string, 0)

	if v.SuccessRate != nil {
		seq = append(seq, fmt.Sprintf("%7.3f", *v.SuccessRate))
	}

	seq = append(seq, v.Rule.String())

	if v.HardMM != nil {
		seq = append(seq, fmt.Sprintf("%4s minmax %-32s soft %s on %s", v.Rule.Unit, *v.HardMM, *v.SoftMM, v.Interval))
	} else {
		seq = append(seq, fmt.Sprintf("%4s soft %s on %s", v.Rule.Unit, *v.SoftMM, v.Interval))
	}

	return v.Code.sprintf(strings.Join(seq, " | "))
}

func (v Status) MarshalJSON() ([]byte, error) {
	type Struct Status

	return json.Marshal(struct {
		*Struct
		IntervalInSec int `json:"interval"`
	}{
		Struct:        (*Struct)(&v),
		IntervalInSec: int(v.Interval.Seconds()),
	})
}

//
//

type StatusNode struct {
	Status StatusCode `json:"code,omitempty"`
	Node   *Node      `json:"node,omitempty"`
	Checks []Status   `json:"status,omitempty"`
}

func (v StatusNode) String() string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("\033[37m%s ⇒ %s\033[0m\n", v.Node.Name, v.Node))

	for _, s := range v.Checks {
		sb.WriteString(fmt.Sprintf("%s\n", s))
	}

	return sb.String()
}

type StatusCluster struct {
	Status  StatusCode
	Cluster *Cluster
	Writer  []StatusNode
	Reader  []StatusNode
}

type StatusRegion struct {
	Status   StatusCode
	Clusters []StatusCluster
	Nodes    []StatusNode
}

func (v StatusRegion) String() string {
	formatter := func(prefix string, status StatusNode) string {
		errors := make([]string, 0)
		for _, s := range status.Checks {
			if s.Code > STATUS_CODE_SUCCESS {
				errors = append(errors, s.Rule.ID)
			}
		}

		if len(errors) != 0 {
			return fmt.Sprintf("\033[31m%s %-36s | %-25s | %s | %s\033[0m\n", prefix, status.Node.Name, status.Node.Engine, status.Node.Zones, strings.Join(errors, " "))
		}

		return fmt.Sprintf("%s %-36s | %-25s | %s\n", prefix, status.Node.Name, status.Node.Engine, status.Node.Zones)
	}

	sb := strings.Builder{}
	for _, c := range v.Clusters {
		sb.WriteString(fmt.Sprintf("%s\n", c.Cluster.ID))
		for _, w := range c.Writer {
			sb.WriteString(formatter("[w]", w))
		}
		for _, r := range c.Reader {
			sb.WriteString(formatter("[r]", r))
		}
		sb.WriteString("\n")
	}

	for _, n := range v.Nodes {
		sb.WriteString(formatter("", n))
	}

	return sb.String()
}
