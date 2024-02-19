//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/cobra"
	"github.com/zalando/rds-health/internal/service"
	"github.com/zalando/rds-health/internal/show"
)

// Execute is entry point for cobra cli application
func Execute(vsn string) {
	rootCmd.Version = vsn

	if err := rootCmd.Execute(); err != nil {
		e := err.Error()
		fmt.Println(strings.ToUpper(e[:1]) + e[1:])
		os.Exit(1)
	}
}

var (
	outColored   bool
	outVerbose   bool
	outSilent    bool
	outJsonify   bool
	rootDatabase string
	rootInterval string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&outColored, "color", "C", false, "output colored")
	rootCmd.PersistentFlags().BoolVarP(&outVerbose, "verbose", "v", false, "output detailed information")
	rootCmd.PersistentFlags().BoolVar(&outSilent, "silent", false, "output nothing")
	rootCmd.PersistentFlags().BoolVar(&outJsonify, "json", false, "output raw json")
	//
	rootCmd.PersistentFlags().StringVarP(&rootDatabase, "database", "n", "", "AWS RDS database name")
	rootCmd.PersistentFlags().StringVarP(&rootInterval, "interval", "t", "24h", "time interval either in minutes (m), hours (h), days (d) or week (w)")

}

var rootCmd = &cobra.Command{
	Use:   "rds-health",
	Short: "command line interface to check health of AWS RDS",
	Long: `
The health utility is a command-line utility to check "health" of AWS RDS
instances and clusters using 12 simple rules. The health utility conducts
analysis of using time-series metrics collected by AWS Performance Insights.

    It is essential requirement to enable AWS Performance Insight for
    AWS RDS instances before using rds-health. 

This utility is the faster way to check the health status of AWS RDS instance.
The health utility has defined 12 rules to be checked. For each rule,
the utility reports the status (passed, failed), percent of time the rules is
passed, and actual values. In order to reduce number of false positives,
the utility applies softening on raw data to remove outliers.

    rds-health check -t 7d -n my-example-database

    STATUS       %            MIN            AVG            MAX	 ID CHECK
    FAILED  32.14%           0.03          13.33         250.61	 D3: storage i/o latency
    WARNED 100.00%           4.10           4.34           4.69	 P4: db transactions (xact_commit)
    FAILED 100.00%           1.04           1.06           1.61	 P5: sql efficiency

    FAIL my-example-database

    (use "rds-health check -v -n my-example-database" to see full report)


Health rules

C1: CPU utilization (os.cpuUtilization.total) - Typical database workloads is
bound to memory or storage, high CPU is anomaly that requires further
investigation.

C2: CPU await for storage (os.cpuUtilization.wait) - High value is the indicated
of database instance to be bounded by the storage capacity. Highly likely the
storage needs to be scaled.

M1: swapped in from disk (os.swap.in) - Any intensive activities indicates that
system is swapping. It is an indication about having low memory.

M2: swapped out to disk (os.swap.out) - Any intensive activities indicates that
system is swapping. It is an indication about having low memory.

D1: storage read i/o (os.diskIO.rdsdev.readIOsPS) - A very low value shows that
the entire dataset is served from memory. In this case, align the storage
capacity with the overall database workload so that storage capacity is enough
to handle peak traffic. The number shall be aligned with the storage
architecture deployed for the database instance.

D2: storage write i/o (os.diskIO.rdsdev.writeIOsPS) - High number shows that
the workload is write-mostly and potentially bound to the disk storage.

D3: storage i/o latency (os.diskIO.rdsdev.await) - The metric reflect a time
used by the storage to fulfill the database queries. High latency on the storage
implies a high latency of SQL queries. Please be aware that latency above 10ms
requires improvement to the storage system. A typically disk latency should be
less than 4 - 5 ms. Please validate that application SLOs are not impacted if
application latency above 5 ms.

P1: database cache hit ratio - Any values below 80 percent show that database
have insufficient amount of shared buffers or physical RAM. Data required for
top-called queries don't fit into memory, and database has to read it from disk.

P2: database blocks read latency (db.IO.blk_read_time) - The metric reflect a
time used by the database to read blocks from the storage. High latency on the
storage implies a high latency of SQL queries.

P3: database deadlocks (db.Concurrency.deadlocks) - Number of deadlocks detected
in this database. Ideally, it shall be 0 shall be 0. The application schema and
I/O logic requires evaluation if number is high.

P4: database transactions (db.Transactions.xact_commit) - Number of transaction
executed by database. The low number indicates that database instance is standby.

P5: SQL efficiency - SQL efficiency shows the percentage of rows fetched by
the client vs rows returned from the storage. The metric does not necessarily
show any performance issue with databases but high ratio of returned vs
fetched rows should trigger the question about optimization of SQL queries,
schema or indexes.

Usage:

* checking the health status of individual instances or entire fleet
* plan database capacity and its scalability
* analysis of the database workloads
* debug anomalies

Examples:

  rds-health check -t 7d
  rds-health check -t 7d -n my-example-database
  rds-health show -t 7d -n my-example-database
  rds-health list

`,
	Run:              root,
	PersistentPreRun: setup,
}

func root(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func setup(cmd *cobra.Command, args []string) {
	if outColored {
		show.SCHEMA = show.SCHEMA_COLOR
	}
}

//
// utils for commands
//

// decodes human-readable time interval to time.Duration
func parseInterval() (time.Duration, error) {
	v, err := strconv.Atoi(rootInterval[0 : len(rootInterval)-1])
	if err != nil {
		return 0, err
	}

	switch rootInterval[len(rootInterval)-1] {
	case 'm':
		return time.Duration(v) * time.Minute, nil
	case 'h':
		return time.Duration(v) * time.Hour, nil
	case 'd':
		return time.Duration(v) * time.Hour * 24, nil
	case 'w':
		return time.Duration(v) * time.Hour * 24 * 7, nil
	default:
		return 0, fmt.Errorf("time scale %s is not supported", rootInterval)
	}
}

// outputs result of printer to stdout
func stdout(data []byte, err error) error {
	if err != nil {
		return err
	}

	if _, err := os.Stdout.Write(data); err != nil {
		return err
	}

	return nil
}

// outputs to stderr
func stderr(data string) {
	if !outSilent {
		os.Stderr.WriteString(data)
	}
}

func WithService(
	f func(cmd *cobra.Command, args []string, api Service) error,
) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		conf, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return err
		}

		var api Service

		switch {
		case outSilent:
			api = service.New(conf, silentbar(0))
		default:
			api = newServiceWithSpinner(conf)
		}

		return f(cmd, args, api)
	}
}
