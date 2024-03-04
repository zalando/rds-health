//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package minimal

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/lynn9388/supsub"
	"github.com/zalando/rds-health/internal/show"
	"github.com/zalando/rds-health/internal/types"
)

//
// Show config information about Nodes, Clusters, Regions
//

var (
	// Show node config as one liner
	// 1a postgres 14.7 db.m5.large 2x 8 GiB 100 GiB gp2 ro example-database-a
	showConfigNode = show.FromShow[types.Node](
		func(n types.Node) ([]byte, error) {
			zone := ""
			if len(n.Zones) > 0 {
				zone = n.Zones[0]
			}
			az := zone[len(zone)-2:]

			cpu := "-"
			mem := "-"
			if n.Compute != nil {
				cpu = fmt.Sprintf("%dx", n.Compute.CPU.Cores)
				mem = n.Compute.Memory.Size.String()
			}

			role := ""
			if n.ReadOnly {
				role = "ro"
			}

			text := fmt.Sprintf("%2s %-17s %-6s %-15s %3s %7s %8s %-6s %2s %s\n", az, n.Engine.ID, n.Engine.Version, n.Type, cpu, mem, n.Storage.Size, n.Storage.Type, role, n.Name)
			return []byte(text), nil
		},
	)

	// Show cluster config as one liner
	// postgres 14.7                                       example-cluster
	showConfigCluster = show.FromShow[types.Cluster](
		func(c types.Cluster) ([]byte, error) {
			text := fmt.Sprintf("%2s %-17s %-6s %46s "+show.SCHEMA.Cluster+"\n", "", c.Engine.ID, c.Engine.Version, "", c.ID)
			return []byte(text), nil
		},
	)

	// Show cluster and its nodes config, one line per node
	ShowConfigCluster = show.Cluster(
		showConfigCluster,
		showConfigNode,
		func(c types.Cluster) ([]types.Node, []types.Node) { return c.Writer, c.Reader },
	)

	// Show all instances in region and its configuration
	ShowConfigRegion = show.Prefix[types.Region](
		fmt.Sprintf("%2s %-17s %-6s %-15s %3s %7s %8s %-6s %-2s %s\n", "AZ", "ENGINE", "VSN", "INSTANCE", "CPU", "MEM", "STORAGE", "TYPE", "RO", "NAME"),
	).FMap(
		show.Region[types.Region](
			ShowConfigCluster,
			showConfigNode,
			func(sr types.Region) ([]types.Cluster, []types.Node) { return sr.Clusters, sr.Nodes },
		),
	)
)

//
// Show health status about Nodes, Clusters, Regions
//

var (
	// show a symbol (NONE, PASS, WARN, FAIL) before label
	ShowHealthSymbol = show.FromShow[types.StatusCode](
		func(code types.StatusCode) ([]byte, error) {
			text := show.StatusIcon(code)
			return []byte(text), nil
		},
	)

	// Show the status of single check
	// FAILED   5.55%           0.56          11.53          44.80	 D3: storage i/o latency
	showHealthRule = show.FromShow[types.Status](
		func(status types.Status) ([]byte, error) {
			if status.Code > types.STATUS_CODE_SUCCESS {
				rate := 100.0 - *status.SuccessRate

				b := &bytes.Buffer{}
				ffs := show.SCHEMA.FmtForStatus(status.Code)
				b.WriteString(fmt.Sprintf(ffs+" "+ffs+" %4s %14.2f %14.2f %14.2f\t %s: %s\n", status.Code, fmt.Sprintf("%6.2f%%", rate), status.Rule.Unit, status.SoftMM.Min, status.SoftMM.Avg, status.SoftMM.Max, status.Rule.ID, status.Rule.About))
				return b.Bytes(), nil
			}

			return nil, nil
		},
	)

	// Show node health status as one line
	// PASS example-database-a
	showHealthNode = show.FromShow[types.StatusNode](
		func(node types.StatusNode) ([]byte, error) {
			status := show.StatusText(node.Status)
			text := fmt.Sprintf("%s %s\n", status, node.Node.Name)
			return []byte(text), nil
		},
	)

	// Show node health status as one line
	// ✅ PASS example-database-a
	showHealthNodeWithSymbol = show.Prefix[types.StatusNode]("\n").FMap(
		show.Printer2[types.StatusNode, types.StatusCode, types.StatusNode]{
			A: ShowHealthSymbol,
			B: showHealthNode,
			UnApply2: func(sn types.StatusNode) (types.StatusCode, types.StatusNode) {
				return sn.Status, sn
			},
		},
	)

	// Show node health status as one line including indicator for each rule
	// FAIL ¦ -- -- -- -- -- -- D3 -- -- -- P4 P5 ¦ example-database-a
	showHealthNodeWithRules = show.FromShow[types.StatusNode](
		func(n types.StatusNode) ([]byte, error) {
			seq := make([]string, len(n.Checks))
			for i, status := range n.Checks {
				switch status.Code {
				case types.STATUS_CODE_FAILURE:
					seq[i] = fmt.Sprintf(show.SCHEMA.StatusCodeText.FAIL, status.Rule.ID)
				case types.STATUS_CODE_WARNING:
					seq[i] = fmt.Sprintf(show.SCHEMA.StatusCodeText.WARN, supsub.ToSup(status.Rule.ID))
				default:
					seq[i] = "--"
				}
			}

			status := show.StatusText(n.Status)

			text := fmt.Sprintf("%s %s %s\n", status, strings.Join(seq, " "), n.Node.Name)
			return []byte(text), nil
		},
	)

	// Show health status of node and all failed rules
	// STATUS       % UNIT           MIN            AVG            MAX	 ID CHECK
	// FAILED   5.55% iops          0.56          11.53          44.80	 D3: storage i/o latency
	// WARNED  96.19%  tps          4.17           4.49           5.37	 P4: db transactions (xact_commit)
	// FAILED 100.00%    %          1.99           0.18           0.06	 P5: sql efficiency
	//
	// ❌ FAIL example-database
	//
	ShowHealthNode = show.Printer2[types.StatusNode, []types.Status, types.StatusNode]{
		A: show.Prefix[[]types.Status](
			fmt.Sprintf("%6s %7s %4s %14s %14s %14s\t%3s %s\n", "STATUS", "%", "UNIT", "MIN", "AVG", "MAX", "ID", "CHECK"),
		).FMap(show.Seq[types.Status]{T: showHealthRule}),
		B: showHealthNodeWithSymbol,
		UnApply2: func(sn types.StatusNode) ([]types.Status, types.StatusNode) {
			return sn.Checks, sn
		},
	}

	// Show cluster health status as one line
	// PASS example-cluster
	showHealthCluster = show.FromShow[types.StatusCluster](
		func(c types.StatusCluster) ([]byte, error) {
			status := show.StatusText(c.Status)
			text := fmt.Sprintf("%s "+show.SCHEMA.Cluster+"\n", status, c.Cluster.ID)
			return []byte(text), nil
		},
	)

	// Show cluster health as one line including the formatting for rules
	// PASS example-cluster
	showHealthClusterWithRules = show.FromShow[types.StatusCluster](
		func(c types.StatusCluster) ([]byte, error) {
			status := show.StatusText(c.Status)

			text := fmt.Sprintf("%s %35s "+show.SCHEMA.Cluster+"\n", status, "", c.Cluster.ID)
			return []byte(text), nil
		},
	)

	// Show region health status as one line
	// PASS 14 health checks
	showHealthRegion = show.FromShow[types.StatusRegion](
		func(r types.StatusRegion) ([]byte, error) {
			nall := len(r.Clusters) + len(r.Nodes)
			pass := 0
			for _, c := range r.Clusters {
				if c.Status <= types.STATUS_CODE_SUCCESS {
					pass++
				}
			}

			for _, n := range r.Nodes {
				if n.Status <= types.STATUS_CODE_SUCCESS {
					pass++
				}
			}

			if nall == pass {
				text := fmt.Sprintf("\n%s%s %d health checks\n", show.StatusIcon(r.Status), show.StatusText(r.Status), nall)
				return []byte(text), nil
			}

			text := fmt.Sprintf("\n%s%s %d health checks (%d passed)\n", show.StatusIcon(r.Status), show.StatusText(r.Status), nall-pass, pass)
			return []byte(text), nil
		},
	)

	// Show health of cluster and its nodes, one line per node
	ShowHealthCluster = show.Cluster(
		showHealthCluster,
		show.Prefix[types.StatusNode]("     ").FMap(showHealthNode),
		func(sc types.StatusCluster) ([]types.StatusNode, []types.StatusNode) { return sc.Writer, sc.Reader },
	)

	// Show health of clusters and nodes in the region
	showHealthRegionMembers = show.Region[types.StatusRegion](
		ShowHealthCluster,
		showHealthNode,
		func(sr types.StatusRegion) ([]types.StatusCluster, []types.StatusNode) { return sr.Clusters, sr.Nodes },
	)

	// Show health of clusters and nodes in the region, including status for each rule
	showHealthRegionMembersWithRules = show.Prefix[types.StatusRegion](
		"     C1 C2 M1 M2 D1 D2 D3 P1 P2 P3 P4 P5\n",
	).FMap(
		show.Region[types.StatusRegion](
			show.Cluster(
				showHealthClusterWithRules,
				showHealthNodeWithRules,
				func(sc types.StatusCluster) ([]types.StatusNode, []types.StatusNode) { return sc.Writer, sc.Reader },
			),
			showHealthNodeWithRules,
			func(sr types.StatusRegion) ([]types.StatusCluster, []types.StatusNode) { return sr.Clusters, sr.Nodes },
		),
	)

	// Show health of region and its objects
	ShowHealthRegion = show.Printer2[types.StatusRegion, types.StatusRegion, types.StatusRegion]{
		A: showHealthRegionMembers,
		B: showHealthRegion,
		UnApply2: func(sr types.StatusRegion) (types.StatusRegion, types.StatusRegion) {
			return sr, sr
		},
	}

	// Show enhanced health of region and its objects
	ShowHealthRegionWithRules = show.Printer2[types.StatusRegion, types.StatusRegion, types.StatusRegion]{
		A: showHealthRegionMembersWithRules,
		B: showHealthRegion,
		UnApply2: func(sr types.StatusRegion) (types.StatusRegion, types.StatusRegion) {
			return sr, sr
		},
	}
)

//
// Show Values of Rules
//

var (
	// Show measured values for each rule
	showValueRule = show.FromShow[types.Status](
		func(status types.Status) ([]byte, error) {
			b := &bytes.Buffer{}
			if status.SoftMM != nil {
				b.WriteString(fmt.Sprintf("%4s %14.2f %14.2f %14.2f %s\n", status.Rule.Unit, status.SoftMM.Min, status.SoftMM.Avg, status.SoftMM.Max, status.Rule.About))
			}

			return b.Bytes(), nil
		},
	)

	// Show short information about node
	showInfoNode = show.FromShow[types.StatusNode](
		func(node types.StatusNode) ([]byte, error) {
			text := fmt.Sprintf("\n%s (%s, %s)\n", node.Node.Name, node.Node.Type, node.Node.Engine)
			return []byte(text), nil
		},
	)

	// Show stats about node
	ShowValueNode = show.Prefix[types.StatusNode](
		fmt.Sprintf("%s %14s %14s %14s\n", "UNIT", "MIN", "AVG", "MAX"),
	).FMap(
		show.Printer2[types.StatusNode, []types.Status, types.StatusNode]{
			A:        show.Seq[types.Status]{T: showValueRule},
			B:        showInfoNode,
			UnApply2: func(sn types.StatusNode) ([]types.Status, types.StatusNode) { return sn.Checks, sn },
		},
	)
)
