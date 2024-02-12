//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/zalando/rds-health/internal/show"
	"github.com/zalando/rds-health/internal/show/minimal"
	"github.com/zalando/rds-health/internal/show/verbose"
	"github.com/zalando/rds-health/internal/types"
)

var (
	// checkIgnore   string
	checkDuration time.Duration
	checkStatus   types.StatusCode
)

func init() {
	rootCmd.AddCommand(checkCmd)
	// checkCmd.Flags().StringVar(&checkIgnore, "ignore", "", "comma separated list of rules to ignore")
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "check health status of database instance using AWS Performance Insights service",
	Example: `
rds-health check -n myrds -t 7d
	`,
	SilenceUsage: true,
	PreRunE:      checkOpts,
	RunE:         WithService(check),
	PostRunE:     checkPost,
}

func checkOpts(cmd *cobra.Command, args []string) (err error) {
	checkDuration, err = parseInterval()
	if err != nil {
		return err
	}

	return nil
}

func checkPost(cmd *cobra.Command, args []string) error {
	if (rootDatabase == "") && !outVerbose && !outJsonify {
		stderr("\n(use \"rds-health check -v\" to see details)\n")
	}

	if (rootDatabase == "") && outVerbose && !outJsonify {
		stderr("\n(use \"rds-health check -n NAME\" for the status of the instance)\n")
	}

	if rootDatabase != "" && !outVerbose && !outJsonify {
		stderr("\n(use \"rds-health check -v -n " + rootDatabase + "\" to see full report)\n")
	}

	if checkStatus > types.STATUS_CODE_SUCCESS {
		os.Exit(128)
	}

	return nil
}

func check(cmd *cobra.Command, args []string, api Service) error {
	if rootDatabase == "" {
		var out show.Printer[types.StatusRegion] = minimal.ShowHealthRegion
		switch {
		case outVerbose:
			out = minimal.ShowHealthRegionWithRules
		case outSilent:
			out = show.None[types.StatusRegion]()
		case outJsonify:
			out = show.JSON[types.StatusRegion]()
		}

		return checkRegion(cmd, args, api, out)
	}

	var out show.Printer[types.StatusNode] = minimal.ShowHealthNode
	switch {
	case outVerbose:
		out = verbose.ShowHealthNode
	case outSilent:
		out = show.None[types.StatusNode]()
	case outJsonify:
		out = show.JSON[types.StatusNode]()
	}

	return checkNode(cmd, args, api, out)
}

func checkRegion(cmd *cobra.Command, args []string, api Service, show show.Printer[types.StatusRegion]) error {
	status, err := api.CheckHealthRegion(cmd.Context(), checkDuration)
	if err != nil {
		return err
	}

	checkStatus = status.Status
	return stdout(show.Show(*status))
}

func checkNode(cmd *cobra.Command, args []string, api Service, show show.Printer[types.StatusNode]) error {
	status, err := api.CheckHealthNode(cmd.Context(), rootDatabase, checkDuration)
	if err != nil {
		return err
	}

	checkStatus = status.Status
	return stdout(show.Show(*status))
}
