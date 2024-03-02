//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package verbose

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/zalando/rds-health/internal/show"
	"github.com/zalando/rds-health/internal/show/minimal"
	"github.com/zalando/rds-health/internal/types"
)

//
// Show config information about Nodes, Clusters, Regions
//

var (
	// Show detailed information about node:
	//
	//	example-database-a
	//		  Engine ¦ postgres v11.19
	//		Instance ¦ db.m5.large
	//		     CPU ¦ 2x 3.10 GHz
	//		  Memory ¦ 8 GiB
	//		 Storage ¦ 100 GiB, gp2
	//       Zones ¦ eu-central-1b
	showConfigNode = show.FromShow[types.Node](
		func(node types.Node) ([]byte, error) {
			cpu := "-"
			mem := "-"
			if node.Compute != nil {
				cpu = node.Compute.CPU.String()
				mem = node.Compute.Memory.String()
			}

			ro := ""
			if node.ReadOnly {
				ro = " (read-only)"
			}

			b := &bytes.Buffer{}
			b.WriteString(fmt.Sprintf("\n%s%s\n", node.Name, ro))
			b.WriteString(fmt.Sprintf("\t%9s ¦ %s\n", "Engine", node.Engine))
			b.WriteString(fmt.Sprintf("\t%9s ¦ %s\n", "Instance", node.Type))
			b.WriteString(fmt.Sprintf("\t%9s ¦ %s\n", "CPU", cpu))
			b.WriteString(fmt.Sprintf("\t%9s ¦ %s\n", "Memory", mem))
			b.WriteString(fmt.Sprintf("\t%9s ¦ %s\n", "Storage", node.Storage))
			b.WriteString(fmt.Sprintf("\t%9s ¦ %s\n", "Zones", strings.Join(node.Zones, ", ")))
			return b.Bytes(), nil
		},
	)

	// Show cluster information as one liner
	//	example-cluster
	//		  Engine ¦ postgres v11.19
	//		 Writers ¦ example-node-a, example-node-b
	//		 Readers ¦ example-node-c, example-node-d
	showConfigCluster = show.FromShow[types.Cluster](
		func(c types.Cluster) ([]byte, error) {
			w := make([]string, len(c.Writer))
			for i, x := range c.Writer {
				w[i] = x.Name
			}

			r := make([]string, len(c.Reader))
			for i, x := range c.Reader {
				r[i] = x.Name
			}

			b := &bytes.Buffer{}
			b.WriteString(fmt.Sprintf("\n"+show.SCHEMA.Cluster+"\n", c.ID))
			b.WriteString(fmt.Sprintf("\t%9s ¦ %s\n", "Engine", c.Engine))
			b.WriteString(fmt.Sprintf("\t%9s ¦ %s\n", "Writers", strings.Join(w, ", ")))
			b.WriteString(fmt.Sprintf("\t%9s ¦ %s\n", "Readers", strings.Join(r, ", ")))
			return b.Bytes(), nil
		},
	)

	// Show cluster and its nodes config
	ShowConfigCluster = show.Cluster(
		showConfigCluster,
		showConfigNode,
		func(c types.Cluster) ([]types.Node, []types.Node) { return c.Writer, c.Reader },
	)

	ShowConfigRegion = show.Region[types.Region](
		ShowConfigCluster,
		showConfigNode,
		func(sr types.Region) ([]types.Cluster, []types.Node) { return sr.Clusters, sr.Nodes },
	)
)

//
// Show health status about Nodes, Clusters, Regions
//

var (
	// show MinMax measurement as one liner
	showMinMax = show.FromShow[types.MinMax](
		func(mm types.MinMax) ([]byte, error) {
			text := fmt.Sprintf("min: %6.2f\tavg: %6.2f\tmax: %6.2f", mm.Min, mm.Avg, mm.Max)
			return []byte(text), nil
		},
	)

	// Show the status of single check
	// FAILED   5.55%           0.56          11.53          44.80	 D3: storage i/o latency
	showHealthRule = show.FromShow[types.Status](
		func(status types.Status) ([]byte, error) {
			b := &bytes.Buffer{}

			rate := *status.SuccessRate
			if status.Code > types.STATUS_CODE_SUCCESS {
				rate = 100.0 - *status.SuccessRate
			}

			ffs := show.SCHEMA.FmtForStatus(status.Code)
			b.WriteString(fmt.Sprintf(ffs+" "+ffs+" %4s %14.2f %14.2f %14.2f\t %s: %s\n", status.Code, fmt.Sprintf("%6.2f%%", rate), status.Rule.Unit, status.SoftMM.Min, status.SoftMM.Avg, status.SoftMM.Max, status.Rule.ID, status.Rule.About))
			return b.Bytes(), nil
		},
	)

	// Show node health status as one line
	showHealthNode = show.FromShow[types.StatusNode](
		func(node types.StatusNode) ([]byte, error) {
			status := show.StatusText(node.Status)

			cpu := "-"
			mem := "-"
			if node.Node.Compute != nil {
				cpu = node.Node.Compute.CPU.String()
				mem = node.Node.Compute.Memory.String()
			}

			ro := ""
			if node.Node.ReadOnly {
				ro = " (read-only)"
			}

			b := &bytes.Buffer{}
			b.WriteString(fmt.Sprintf("%s %s%s\n", status, node.Node.Name, ro))
			b.WriteString(fmt.Sprintf("%14s ¦ %s\n", "Engine", node.Node.Engine))
			b.WriteString(fmt.Sprintf("%14s ¦ %s\n", "Instance", node.Node.Type))
			b.WriteString(fmt.Sprintf("%14s ¦ %s\n", "CPU", cpu))
			b.WriteString(fmt.Sprintf("%14s ¦ %s\n", "Memory", mem))
			b.WriteString(fmt.Sprintf("%14s ¦ %s\n", "Storage", node.Node.Storage))
			b.WriteString(fmt.Sprintf("%14s ¦ %s\n\n", "Zones", strings.Join(node.Node.Zones, ", ")))
			return b.Bytes(), nil
		},
	)

	showHealthNodeWithSymbol = show.Prefix[types.StatusNode]("\n").FMap(
		show.Printer2[types.StatusNode, types.StatusCode, types.StatusNode]{
			A: minimal.ShowHealthSymbol,
			B: showHealthNode,
			UnApply2: func(sn types.StatusNode) (types.StatusCode, types.StatusNode) {
				return sn.Status, sn
			},
		},
	)

	//	FAIL example-database-a
	//		FAILED 99.9% ¦ C01: cpu utilization
	//			         % ¦ min: 17.5	avg: 25.0	max: 80.0
	//
	ShowHealthNode = show.Prefix[types.StatusNode](
		fmt.Sprintf("%6s %7s %4s %14s %14s %14s\t%3s %s\n", "STATUS", "%", "UNIT", "MIN", "AVG", "MAX", "ID", "CHECK"),
	).FMap(
		show.Printer2[types.StatusNode, []types.Status, types.StatusNode]{
			A: show.Seq[types.Status]{T: showHealthRule},
			B: showHealthNodeWithSymbol,
			UnApply2: func(sn types.StatusNode) ([]types.Status, types.StatusNode) {
				return sn.Checks, sn
			},
		},
	)
)

//
// Show Values of Rules
//

var (
	// Show measured values for each rule
	showValueRule = show.FromShow[types.Status](
		func(status types.Status) ([]byte, error) {
			b := &bytes.Buffer{}
			b.WriteString(fmt.Sprintf("\n%s (%s)\n", status.Rule.About, status.Rule.Unit))

			if status.SoftMM != nil {
				soft, _ := showMinMax.Show(*status.SoftMM)
				b.WriteString(fmt.Sprintf("%s ¦ %s\n", "soft", string(soft)))
			}

			if status.HardMM != nil {
				hard, _ := showMinMax.Show(*status.HardMM)
				b.WriteString(fmt.Sprintf("%s ¦ %s\n", "hard", string(hard)))
			}

			return b.Bytes(), nil
		},
	)

	// Show short information about node
	showInfoNode = show.FromShow[types.StatusNode](
		func(node types.StatusNode) ([]byte, error) {
			text := fmt.Sprintf("%s (%s, %s)\n", node.Node.Name, node.Node.Type, node.Node.Engine)
			return []byte(text), nil
		},
	)

	// Show stats about node
	ShowValueNode = show.Printer2[types.StatusNode, types.StatusNode, []types.Status]{
		A: showInfoNode,
		B: show.Seq[types.Status]{T: showValueRule},
		UnApply2: func(sn types.StatusNode) (types.StatusNode, []types.Status) {
			return sn, sn.Checks
		},
	}
)
