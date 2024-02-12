//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/zalando/rds-health/internal/show"
	"github.com/zalando/rds-health/internal/show/minimal"
	"github.com/zalando/rds-health/internal/show/verbose"
	"github.com/zalando/rds-health/internal/types"
)

var (
	showDuration time.Duration
)

func init() {
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show resource utilization",
	Long:  "show system resource utilization of RDS instance using AWS Performance Inside",
	Example: `
rds-health show -n name-of-rds-instance -t 7d
rds-health show -n name-of-rds-instance -t 7d -a max
	`,
	SilenceUsage: true,
	PreRunE:      usageOpts,
	RunE:         WithService(showNode),
}

func usageOpts(cmd *cobra.Command, args []string) (err error) {
	showDuration, err = parseInterval()
	if err != nil {
		return err
	}

	if rootDatabase == "" {
		return fmt.Errorf("undefined database name")
	}

	return nil
}

func showNode(cmd *cobra.Command, args []string, api Service) error {
	var out show.Printer[types.StatusNode] = minimal.ShowValueNode
	switch {
	case outVerbose:
		out = verbose.ShowValueNode
	case outSilent:
		out = show.None[types.StatusNode]()
	case outJsonify:
		out = show.JSON[types.StatusNode]()
	}

	usage, err := api.ShowNode(cmd.Context(), rootDatabase, showDuration)
	if err != nil {
		return err
	}

	return stdout(out.Show(*usage))
}
