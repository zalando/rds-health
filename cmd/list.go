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

	"github.com/spf13/cobra"
	"github.com/zalando/rds-health/internal/show"
	"github.com/zalando/rds-health/internal/show/minimal"
	"github.com/zalando/rds-health/internal/show/verbose"
	"github.com/zalando/rds-health/internal/types"
)

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.InheritedFlags().SetAnnotation("database", cobra.BashCompOneRequiredFlag, []string{"false"})
	listCmd.InheritedFlags().SetAnnotation("interval", cobra.BashCompOneRequiredFlag, []string{"false"})
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all database instances and clusters in AWS account",
	Example: `
rds-health list
	`,
	SilenceUsage: true,
	RunE:         WithService(list),
	PostRun:      listPost,
}

func list(cmd *cobra.Command, args []string, api Service) error {
	out := minimal.ShowConfigRegion
	switch {
	case outVerbose:
		out = verbose.ShowConfigRegion
	case outSilent:
		out = show.None[types.Region]()
	case outJsonify:
		out = show.JSON[types.Region]()
	}

	region, err := api.ShowRegion(cmd.Context())
	if err != nil {
		return err
	}

	if len(region.Clusters)+len(region.Nodes) == 0 {
		return fmt.Errorf("no instances are found")
	}

	return stdout(out.Show(*region))
}

func listPost(cmd *cobra.Command, args []string) {
	if !outJsonify {
		stderr("\n(use \"rds-health check\" to check health status of instances)\n")
	}
}
