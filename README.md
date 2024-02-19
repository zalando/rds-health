<p align="center">
  <img src="./doc/rds-health.png" height="240" />
  <h3 align="center">AWS RDS Health</h3>
  <p align="center"><strong>discover anomalies, performance issues and optimization within AWS RDS</strong></p>

  <p align="center">
    <!-- Discussion -->
    <a href="https://github.com/zalando/rds-health/discussions">
      <img alt="GitHub Discussions" src="https://img.shields.io/github/discussions/zalando/rds-health?logo=github">
    </a>
    <!-- Version -->
    <a href="https://github.com/zalando/rds-health/releases">
      <img src="https://img.shields.io/github/v/tag/zalando/rds-health?label=version" />
    </a>
    <!-- Build Status -->
    <a href="https://github.com/zalando/rds-health/actions/">
      <img src="https://github.com/zalando/rds-health/workflows/test/badge.svg" />
    </a>
    <!-- GitHub -->
    <a href="http://github.com/zalando/rds-health">
      <img src="https://img.shields.io/github/last-commit/zalando/rds-health.svg" />
    </a>
    <!-- Coverage
    <a href="https://coveralls.io/github/zalando/rds-health?branch=main">
      <img src="https://coveralls.io/repos/github/zalando/rds-health/badge.svg?branch=main" />
    </a>
    -->
  </p>
</p>


# AWS RDS Health

`rds-health` is a command-line utility to check "health" of AWS RDS instances, clusters using [12 rules](./doc/health-rules.md). The utility interactively analyses database metrics to discover anomalies, performance issues and detects possible optimizations.


## Quick Example

Let's get your start with `rds-health`. These few simple steps explain how to run a first health check.

### Install

Easiest way to install the latest version of utility using binary release, which are available
either from [Homebrew](https://brew.sh/) taps or [GitHub](https://github.com/zalando/rds-health/releases) for multiple platforms.

```bash
## Install using brew
brew tap zalando/rds-health https://github.com/zalando/rds-health
brew install -q rds-health

## use `brew upgrade` to upgrade to latest version 
```

Alternatively, you can install application from source code but it requires [Golang](https://go.dev/) to be installed.

```bash
go install github.com/zalando/rds-health@latest
```


### Configure & Discover

The `rds-health` utility conducts analysis of AWS RDS instances using time-series metrics collected by AWS Performance Insights. It is essential requirement:

> [AWS Performance Insights MUST be switched on for your database instances](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_PerfInsights.Enabling.html).

Like any other CLI, `rds-health` requires credential to access your AWS Account. It is sufficient to provision read-only credentials. See official AWS guide on [configure the CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html).

Please watch out the region settings. The explicit definition of region is required through environment variable `AWS_DEFAULT_REGION=eu-central-1` if your aws configuration profile misses the default value. 

Start with discovery of your deployments once all configuration is done.

```
rds-health list

AZ ENGINE            VSN    INSTANCE        CPU     MEM  STORAGE TYPE   RO NAME
   aurora-postgresql 14.7                                                  my-cluster-1
1c aurora-postgresql 14.7   db.r5.xlarge     4x  32 GiB  100 GiB aurora    my-cluster-1-node-a
1a aurora-postgresql 14.7   db.r5.xlarge     4x  32 GiB  100 GiB aurora ro my-cluster-1-node-b
   aurora-postgresql 13.8                                                  my-cluster-2
1b aurora-postgresql 13.8   db.t4g.medium    2x   4 GiB    1 GiB aurora    my-cluster-2-node-a
1a aurora-postgresql 13.8   db.t4g.medium    2x   4 GiB    1 GiB aurora ro my-cluster-2-node-b
...

1a postgres          14.7   db.m5.large      2x   8 GiB  400 GiB gp2       my-database-1
1b postgres          14.7   db.t3.medium     2x   4 GiB   40 GiB gp2       my-database-2
...

(use "rds-health check" to check health status of instances)
```


### Check Health

The health utility has defined [**12 rules**](./doc/health-rules.md) to be checked. For each rule, the utility reports `STATUS` (passed, failed), relative quantity of failed samples `%` of time the rules is passed/failed, `MIN`, `AVG` and `MAX` values across all measurements. In order to reduce number of false positives, the utility applies softening on raw data to remove outliers. 

```
rds-health check -t 7d -n my-database-1

STATUS       %            MIN            AVG            MAX	 ID CHECK
FAILED  32.14%           0.03          13.33         250.61	 D3: storage i/o latency
WARNED 100.00%           4.10           4.34           4.69	 P4: db transactions (xact_commit)
FAILED 100.00%           1.04           1.06           1.61	 P5: sql efficiency

FAIL my-database-1

(use "rds-health check -v -n my-database-1" to see full report)
```

The utility deliberately used "min-max" aggregation technique per discrete time interval instead of percentiles. It is derived from AWS Performance Insights capability that persists _the minimum_ and _the maximum_ values of each interval along with _the average_ value. So that `rds-health` utility does not either uses percentiles. It sounds as contradicting with best practices of system monitoring where percentiles become the primary service level indicators. However, there are no math for meaningfully aggregating percentiles. Once telemetry system calculated percentile and discarded the raw data, it is not possible aggregate the summarized percentiles into anything useful. Averaging percentile leads to bogus result. Min-Max analysis is only an alternative technique applicable here that get an observability of the full range of the data.

The utility obtains [database metrics](./internal/rules/metrics.go) as a time-series data. AWS returns these time series as aggregated discrete value on fixed time interval (e.g. 1s, 1m, 5m or 1h). For each interval, utility runs _min-max_ analysis and reports the result. Note together with analysis of "raw data", the utility soften the time-series by filtering the outliers (e.g. night time, busy hours), which helps to get better perspective on typical workload. 


### Capacity Planning

The capacity planning requires a comprehensive view on the workload conducted by the database instance. The health utility provides a single command to fetch essential metrics: the "hardware" configuration (cpu, memory, storage, instance type); executed transactions, read/write tuples, disk I/O, etc.

```
rds-health show -t 7d -n my-database-1

UNIT            MIN            AVG            MAX
 tps           4.10           4.34           4.66 db transactions (xact_commit)
iops          21.52          22.45          34.62 tup_fetched (rows returned by query)
iops        2111.74        2113.84        2178.75 tup_returned (rows read from storage)
iops           0.00           0.06           0.12 tup_inserted (rows inserted to db)
iops           0.00           0.00           0.05 tup_updated (rows updated at db)
iops           0.00           0.00           0.02 tup_deleted (rows deleted from db)
   %           4.10           4.64           6.10 cpu utilization
   %           0.10           0.14           0.85 cpu await for storage
iops           0.00           0.09           0.60 storage read i/o
iops           3.84           5.70          10.31 storage write i/o
iops           0.00           0.00           0.00 blk_read
iops          90.82          93.80         115.03 blks_hit (cache hits)
iops           0.00           0.15           3.33 buffers_checkpoint
  ms           5.00           6.92           9.00 checkpoint_sync_latency
  KB      299018.00      325842.03      412094.00 free memory
  KB     4788554.00     4803596.83     4813992.00 filesys caching memory
  KB    10015684.00    10015718.07    10016248.00 used storage space

my-database-1 (db.m5.large, postgres v14.7)
```

### Next Steps

Run help system to discover all other features

```
rds-health help
```

## How To Contribute

The library is [MIT](./LICENSE.md) licensed and accepts contributions via GitHub pull requests. See [contributing guidelines](./CONTRIBUTING.md)


## License

[![See LICENSE](https://img.shields.io/github/license/zalando/rds-health.svg?style=for-the-badge)](./LICENSE.md)