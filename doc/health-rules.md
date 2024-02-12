# Database health rules

The command-line utility checks the health of AWS RDS.
The utility uses a rules defined by the following checklist.


## C1: cpu utilization

**Metric**: os.cpuUtilization.total (%)

**Condition**: `max cpu util` < 60% and `avg cpu util` < 40%

We should worrying if value is higher than 40%. Typical database workloads is bound to memory or storage, high CPU is anomaly that requires further investigation.

## C2: cpu await for storage

**Metric**: os.cpuUtilization.wait (%)

**Condition**: `max cpu await` < 10% and `avg cpu await` < 8%

Any value above 5%% - 10%% shows suboptimal disk configuration. High value is the indicated of database instance to be bounded by the storage capacity. Highly likely the storage needs to be scaled.

## M1: swapped in from disk

**Metric**: os.swap.in (KB/s)

**Condition**: `max swap in` < 1KB/s and `avg swap in` < 1KB/s

Any intensive activities indicates that system is swapping. It is an indication about having low memory.

## M2: swapped out to disk

**Metric**: os.swap.out (KB/s)

**Condition**: `max swap out` < 1KB/s and `avg swap out` < 1KB/s

Any intensive activities indicates that system is swapping. It is an indication about having low memory.


## D1: storage read i/o

**Metric**: os.diskIO.rdsdev.readIOsPS (IOPS)

**Condition**: `max storage read` < 300 IOPS and `avg storage read` < 100 IOPS

The number shall be aligned with the storage architecture deployed for the database instance. Each instance has a limit of IOPS it can do. With the GP2 volume type, IOPS are provisioned by volume size, 3 IOPS per GB of storage with a minimum of 100 IOPS. IO volume types has explicit value.

A very low value shows that the entire dataset is served from memory. In this case, align the storage capacity with the overall database workload so that storage capacity is enough to handle 

## D2: storage write i/o

**Metric**: os.diskIO.rdsdev.writeIOsPS (IOPS)

**Condition**: `max storage write` < 300 IOPS and `avg storage write` < 100 IOPS

The number shall be aligned with the storage architecture deployed for the database instance. Each instance has a limit of IOPS it can do. With the GP2 volume type, IOPS are provisioned by volume size, 3 IOPS per GB of storage with a minimum of 100 IOPS. IO volume types has explicit value.

High number shows that the workload is write-mostly and potentially bound to the disk storage.

## D3: storage i/o latency

**Metric**: os.diskIO.rdsdev.await (ms)

**Condition**: `max storage latency` < 20 ms and `avg storage latency` < 10 ms

The metric reflect a time used by the storage to fulfill the database queries. High latency on the storage implies a high latency of SQL queries. 

Please be aware that latency above 10ms requires improvement to the storage system. A typically disk latency should be less than 4 - 5 ms. Please validate that application SLOs are not impacted if application latency above 5 ms.


## P1: database cache hit ratio

**Metric**: db.Cache.blks_hit / (db.Cache.blks_hit + db.IO.blk_read)

**Condition**: `min db cache hit ratio` > 80 %and `avg db cache hit ratio` > 90 %

The database does reading and writing of tables data in blocks. Default page size of PostgreSQL is 8192 bytes. Default IO block size in Linux is 4096 bytes. The number of block read by database from the physical storage has to be aligned with storage capacity provisioned to database instance. Database caches these blocks in the memory to optimize the application performance. When clients request data, database checks cached memory and if there are no relevant data there it has to read it from disk, thus queries become slower.  

Any values below 80 % show that database have insufficient amount of shared buffers or physical RAM. Data required for top-called queries don't fit into memory, and database has to read it from disk.


## P2: database blocks read latency

**Metric**: db.IO.blk_read_time (ms)

**Condition**: `max db blocks read latency` < 20 ms and `avg db blocks read latency` < 10 ms

The metric reflect a time used by the database to read blocks from the storage. High latency on the storage implies a high latency of SQL queries. 

Please be aware that latency above 10ms requires validation on the impact of application SLOs and improvement to the storage system.


## P3: database deadlocks

**Metric**: db.Concurrency.deadlocks (tps)

**Condition**: `max db deadlocks` == 0 and `avg db deadlocks` == 0

Number of deadlocks detected in this database. Ideally, it shall be 0  shall be 0. The application schema and I/O logic requires evaluation if number is high.


## P4: database transactions

**Metric**: db.Transactions.xact_commit (tps)

**Condition**: `min db tx` > 3 tps and `avg db tx` > 5 tps

Number of transaction executed by database. The low number indicates that database instance is standby.


## P5: SQL efficiency 

**Metric**: db.SQL.tup_fetched / db.SQL.tup_returned

**Condition**: `min sql efficiency` > 10 % and `avg sql efficiency` > 20 %

SQL efficiency shows the percentage of rows fetched by the client vs rows returned from the storage. The metric does not necessarily show any performance issue with databases but high ratio of returned vs fetched rows should trigger the question about optimization of SQL queries, schema or indexes. 
			
For example, If you do `select count(*) from million_row_table`, one million rows will be returned, but only one row will be fetched.
