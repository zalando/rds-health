//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package compact

import (
	"bytes"
	"fmt"

	"github.com/zalando/rds-health/internal/show"
	"github.com/zalando/rds-health/internal/types"
)

//
// Compact rendering, 2 lines per entity.
//

var (
	showMinMax = show.FromShow[types.MinMax](
		func(mm types.MinMax) ([]byte, error) {
			text := fmt.Sprintf("min: %4.2f\tavg: %4.2f\tmax: %4.2f", mm.Min, mm.Avg, mm.Max)
			return []byte(text), nil
		},
	)

	showHealthRule = show.FromShow[types.Status](
		func(status types.Status) ([]byte, error) {
			if status.Code > types.STATUS_CODE_SUCCESS {
				rate := 100.0 - *status.SuccessRate
				soft, _ := showMinMax.Show(*status.SoftMM)

				b := &bytes.Buffer{}
				b.WriteString(fmt.Sprintf("%s %6.2f%% ¦ %s: %s\n", status.Code, rate, status.Rule.ID, status.Rule.About))
				b.WriteString(fmt.Sprintf("\t%6s ¦ %s\n\n", status.Rule.Unit, string(soft)))
				return b.Bytes(), nil
			}

			return nil, nil
		},
	)

	showHealthNode = show.FromShow[types.StatusNode](
		func(node types.StatusNode) ([]byte, error) {
			status := show.StatusText(node.Status)
			text := fmt.Sprintf("%s %s (%s, %s)\n\n", status, node.Node.Name, node.Node.Engine, node.Node.Type)
			return []byte(text), nil
		},
	)

	//	FAIL example-database-a
	//		FAILED 99.9% ¦ C01: cpu utilization
	//			         % ¦ min: 17.5	avg: 25.0	max: 80.0
	//
	ShowHealthNode = show.Printer2[types.StatusNode, types.StatusNode, []types.Status]{
		A: showHealthNode,
		B: show.Seq[types.Status]{T: showHealthRule},
		UnApply2: func(sn types.StatusNode) (types.StatusNode, []types.Status) {
			return sn, sn.Checks
		},
	}
)
