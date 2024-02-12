//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package rules

// Operating System
var (
	OsCpuUtil = estimator{
		id:   "C1",
		name: "os.cpuUtilization.total",
		unit: "%",
		info: "cpu utilization",
		desc: `
			Should be worrying if value is higher than 40%.
			Typical database workloads is bound to memory or storage, high CPU is anomaly.
		`,
	}

	OsCpuWait = estimator{
		id:   "C2",
		name: "os.cpuUtilization.wait",
		unit: "%",
		info: "cpu await for storage",
		desc: `
		  Any value above 5%% - 10%% shows nonoptimal disk configuration.
			High value is the indicated of database instance to be bounded by the storage capacity. 
		`,
	}

	OsSwapIn = estimator{
		id:   "M1",
		name: "os.swap.in",
		unit: "KB",
		info: "swapped in from disk",
		desc: `
			Any intensive activities indicates that system swaps, OOM symptoms
		`,
	}

	OsSwapOut = estimator{
		id:   "M2",
		name: "os.swap.out",
		unit: "KB",
		info: "swapped out to disk",
		desc: `
			Any intensive activities indicates that system swaps, OOM symptoms
		`,
	}

	OsMemoryTotal = estimator{
		name: "os.memory.total",
		unit: "KB",
		info: "total memory",
		desc: `
			The hard limit system can use memory
		`,
	}

	OsMemoryFree = estimator{
		name: "os.memory.free",
		unit: "KB",
		info: "free memory",
		desc: `
			Higher is better, memory is still available for the application
		`,
	}

	OsMemoryCached = estimator{
		name: "os.memory.cached",
		unit: "KB",
		info: "filesys caching memory",
		desc: `
			Higher is better, amount of memory used to cache file system I/O 
		`,
	}

	OsFileSysTotal = estimator{
		name: "os.fileSys.total",
		unit: "KB",
		info: "total storage space",
		desc: `
			Storage space allocated to the instace
		`,
	}

	OsFileSysUsed = estimator{
		name: "os.fileSys.used",
		unit: "KB",
		info: "used storage space",
		desc: `
			Storage space used by the datasets
		`,
	}
)

var (
	DbStorageReadIO = estimator{
		id:   "D1",
		name: "os.diskIO.rdsdev.readIOsPS",
		unit: "iops",
		info: "storage read i/o",
		desc: `
			The number shall be aligned with IOPS provisioned for the RDS instance.
			A very low value shows that the entire dataset is served from memory.
		`,
	}

	DbStorageWriteIO = estimator{
		id:   "D2",
		name: "os.diskIO.rdsdev.writeIOsPS",
		unit: "iops",
		info: "storage write i/o",
		desc: `
			The number shall be aligned with IOPS provisioned for the RDS instance.
			High number shows that the workload is a write bound one.
		`,
	}

	DbStorageAwait = estimator{
		id:   "D3",
		name: "os.diskIO.rdsdev.await",
		unit: "ms",
		info: "storage i/o latency",
		desc: `
		  The time used by the storage to fulfill I/O.
		  Any value above 10ms requires improvement to the storage system.
		  Any value above 4 - 5ms requires validation that defined SLO is not impacted.
		`,
	}

	DbDataBlockCacheHit = estimator{
		name: "db.Cache.blks_hit",
		unit: "iops",
		info: "blks_hit (cache hits)",
		desc: `
			Data block found from cache, db is not doing physical I/O. higher is better.
		`,
	}

	DbDataBlockReadIO = estimator{
		name: "db.IO.blk_read",
		unit: "iops",
		info: "blk_read",
		desc: `
			Number of blocks read from physical storage.
			The value shall be aligned with IOPS provisioned for the RDS instance. 
		`,
	}

	DbDataBlockCacheHitRatio = calculator{
		id:   "P1",
		lhm:  "db.Cache.blks_hit",
		rhm:  "db.IO.blk_read",
		fop:  func(lhm, rhm float64) float64 { return 100 * lhm / (lhm + rhm) },
		unit: "%",
		info: "db cache hit ratio",
		desc: `
			Any values below 80 % show that database have insufficient amount of
			shared buffers or physical RAM. Data required for top-called queries
			don't fit into memory, and database has to read it from disk.
		`,
	}

	DbDataBlockReadTime = estimator{
		id:   "P2",
		name: "db.IO.blk_read_time",
		unit: "ms",
		info: "db blocks read latency",
		desc: `
			The time spent by database reading blocks.
		`,
	}

	DbBuffersCheckpoints = estimator{
		name: "db.Checkpoint.buffers_checkpoint",
		unit: "iops",
		info: "buffers_checkpoint",
		desc: `
			Number of blocks written by database to physical storage.
			The value shall be aligned with IOPS provisioned for the RDS instance. 
		`,
	}

	DbBuffersCheckpointsTime = estimator{
		name: "db.Checkpoint.checkpoint_sync_latency",
		unit: "ms",
		info: "checkpoint_sync_latency",
		desc: `
		The time spent by database syncing data to disk. 
		`,
	}

	DbDeadlocks = estimator{
		id:   "P3",
		name: "db.Concurrency.deadlocks",
		unit: "tps",
		info: "db deadlocks",
		desc: `
			Number of deadlocks detected in this database. Ideally shall be 0.
			Requires evaluation of application logic if number is high.
		`,
	}

	DbBlockedTransactions = estimator{
		name: "db.Transactions.blocked_transactions",
		unit: "tps",
		info: "blocked_transactions",
		desc: `
			Number of transactions waiting for row lock.
			The high number requires concurrency optimisation only if SLO is impaired.
		`,
	}

	DbRollbacks = estimator{
		name: "db.Transactions.xact_rollback",
		unit: "tps",
		info: "xact_rollback",
		desc: `
			High number indicates issue with the transaction logic in the app (conflicts)
		`,
	}

	DbXactCommit = estimator{
		id:   "P4",
		name: "db.Transactions.xact_commit",
		unit: "tps",
		info: "db transactions (xact_commit)",
		desc: `
			Informative metric shows workload conducted by database.
			The metric contains both read and write queries.
			Please remember that every statement that is not run in a transaction block actually runs in its own little transaction, so it will cause a commit
		`,
	}

	SqlTuplesFetched = estimator{
		name: "db.SQL.tup_fetched",
		unit: "iops",
		info: "tup_fetched (rows returned by query)",
		desc: `
		   Number of rows (tuples) returned by database engine to the client
		`,
	}

	SqlTuplesReturned = estimator{
		name: "db.SQL.tup_returned",
		unit: "iops",
		info: "tup_returned (rows read from storage)",
		desc: `
		  Number of rows (tuples) read from the physical storage to database engine for post-processing.
			A high ratio of tup_returned / tup_fetched indicates on existence of inefficient queries.
		`,
	}

	SqlEfficiency = calculator{
		id:   "P5",
		lhm:  "db.SQL.tup_fetched",
		rhm:  "db.SQL.tup_returned",
		fop:  func(lhm, rhm float64) float64 { return 100 * lhm / rhm },
		unit: "%",
		info: "sql efficiency",
		desc: `
			SQL efficiency shows the percentage of rows fetched by the client vs
			rows returned from the storage. The metric does not necessarily show any
			performance issue with databases but high ratio of returned vs fetched
			rows should trigger the question about optimization of SQL queries,
			schema or indexes. 
			
			For example, If you do "select count(*) from million_row_table",
			one million rows will be returned, but only one row will be fetched.
		`,
	}

	SqlTuplesInserted = estimator{
		name: "db.SQL.tup_inserted",
		unit: "iops",
		info: "tup_inserted (rows inserted to db)",
		desc: `
			Number of rows inserted/updated/deleted in this database.
			Read-mostly workload should not have a high number.
			Any spikes should raise a question, what is going on here.
		`,
	}

	SqlTuplesUpdated = estimator{
		name: "db.SQL.tup_updated",
		unit: "iops",
		info: "tup_updated (rows updated at db)",
		desc: `
			Number of rows inserted/updated/deleted in this database.
			Read-mostly workload should not have a high number.
			Any spikes should raise a question, what is going on here.
		`,
	}

	SqlTuplesDeleted = estimator{
		name: "db.SQL.tup_deleted",
		unit: "iops",
		info: "tup_deleted (rows deleted from db)",
		desc: `
			Number of rows inserted/updated/deleted in this database.
			Read-mostly workload should not have a high number.
			Any spikes should raise a question, what is going on here.
		`,
	}

	DbTempBytes = estimator{
		name: "db.Temp.temp_bytes",
		unit: "B",
		info: "size of temp tables",
		desc: `
			Total amount of data written to temporary files by queries in this instance.
		`,
	}
)
