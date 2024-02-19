//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package main

import (
	"fmt"

	"github.com/zalando/rds-health/cmd"
)

var (
	// See https://goreleaser.com/cookbooks/using-main.version/
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	cmd.Execute(fmt.Sprintf("rds-health/%s (%s), %s", version, commit[:7], date))
}
