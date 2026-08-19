import { ECommandTemplateKind } from "@application/shared/enums";

export type CommandTemplateFromTemplateSelection = {
    commandType: string;
    commandKind: string;
};

export type CommandTemplateFromTemplateItem = CommandTemplateFromTemplateSelection & {
    label: string;
};

export type CommandTemplateFromTemplateGroup = {
    heading: string;
    commandType: string;
    items: CommandTemplateFromTemplateItem[];
};

export const COMMAND_TEMPLATE_FROM_TEMPLATE_CATALOG = [
    {
        heading: "Backup",
        commandType: ECommandTemplateKind.Backup,
        items: [
            { label: "ClickHouse Dump", commandType: ECommandTemplateKind.Backup, commandKind: "clickhouse-dump" },
            {
                label: "ClickHouse Dump (pipe)",
                commandType: ECommandTemplateKind.Backup,
                commandKind: "clickhouse-dump.pipe",
            },
            { label: "ClickHouse Restore", commandType: ECommandTemplateKind.Backup, commandKind: "clickhouse-restore" },
            {
                label: "ClickHouse Restore (pipe)",
                commandType: ECommandTemplateKind.Backup,
                commandKind: "clickhouse-restore.pipe",
            },
            { label: "Dolt Dump", commandType: ECommandTemplateKind.Backup, commandKind: "dolt-dump" },
            { label: "Dolt Dump (pipe)", commandType: ECommandTemplateKind.Backup, commandKind: "dolt-dump.pipe" },
            { label: "Dolt Restore", commandType: ECommandTemplateKind.Backup, commandKind: "dolt-restore" },
            { label: "Dolt Restore (pipe)", commandType: ECommandTemplateKind.Backup, commandKind: "dolt-restore.pipe" },
            {
                label: "ElasticSearch Dump",
                commandType: ECommandTemplateKind.Backup,
                commandKind: "elasticsearch-dump",
            },
            {
                label: "ElasticSearch Dump (pipe)",
                commandType: ECommandTemplateKind.Backup,
                commandKind: "elasticsearch-dump.pipe",
            },
            {
                label: "ElasticSearch Restore",
                commandType: ECommandTemplateKind.Backup,
                commandKind: "elasticsearch-restore",
            },
            {
                label: "ElasticSearch Restore (pipe)",
                commandType: ECommandTemplateKind.Backup,
                commandKind: "elasticsearch-restore.pipe",
            },
            { label: "InfluxDB Dump", commandType: ECommandTemplateKind.Backup, commandKind: "influx-dump" },
            { label: "InfluxDB Dump (pipe)", commandType: ECommandTemplateKind.Backup, commandKind: "influx-dump.pipe" },
            { label: "InfluxDB Restore", commandType: ECommandTemplateKind.Backup, commandKind: "influx-restore" },
            {
                label: "InfluxDB Restore (pipe)",
                commandType: ECommandTemplateKind.Backup,
                commandKind: "influx-restore.pipe",
            },
            { label: "MariaDB Dump", commandType: ECommandTemplateKind.Backup, commandKind: "mariadb-dump" },
            { label: "MariaDB Dump (pipe)", commandType: ECommandTemplateKind.Backup, commandKind: "mariadb-dump.pipe" },
            { label: "MariaDB Restore", commandType: ECommandTemplateKind.Backup, commandKind: "mariadb" },
            { label: "MariaDB Restore (pipe)", commandType: ECommandTemplateKind.Backup, commandKind: "mariadb.pipe" },
            { label: "MongoDB Dump", commandType: ECommandTemplateKind.Backup, commandKind: "mongodump" },
            { label: "MongoDB Dump (pipe)", commandType: ECommandTemplateKind.Backup, commandKind: "mongodump.pipe" },
            { label: "MongoDB Restore", commandType: ECommandTemplateKind.Backup, commandKind: "mongorestore" },
            {
                label: "MongoDB Restore (pipe)",
                commandType: ECommandTemplateKind.Backup,
                commandKind: "mongorestore.pipe",
            },
            { label: "MySQL Dump", commandType: ECommandTemplateKind.Backup, commandKind: "mysqldump" },
            { label: "MySQL Dump (pipe)", commandType: ECommandTemplateKind.Backup, commandKind: "mysqldump.pipe" },
            { label: "MySQL Restore", commandType: ECommandTemplateKind.Backup, commandKind: "mysql" },
            { label: "MySQL Restore (pipe)", commandType: ECommandTemplateKind.Backup, commandKind: "mysql.pipe" },
            { label: "Neon Dump", commandType: ECommandTemplateKind.Backup, commandKind: "neon-dump" },
            { label: "Neon Dump (pipe)", commandType: ECommandTemplateKind.Backup, commandKind: "neon-dump.pipe" },
            { label: "Neon Restore", commandType: ECommandTemplateKind.Backup, commandKind: "neon-restore" },
            { label: "Neon Restore (pipe)", commandType: ECommandTemplateKind.Backup, commandKind: "neon-restore.pipe" },
            { label: "Postgres Dump", commandType: ECommandTemplateKind.Backup, commandKind: "pg_dump" },
            { label: "Postgres Dump (pipe)", commandType: ECommandTemplateKind.Backup, commandKind: "pg_dump.pipe" },
            { label: "Postgres Restore", commandType: ECommandTemplateKind.Backup, commandKind: "pg_restore" },
            { label: "Postgres Restore (pipe)", commandType: ECommandTemplateKind.Backup, commandKind: "pg_restore.pipe" },
            { label: "Redis Dump", commandType: ECommandTemplateKind.Backup, commandKind: "redis-dump" },
            { label: "Redis Dump (pipe)", commandType: ECommandTemplateKind.Backup, commandKind: "redis-dump.pipe" },
            { label: "Redis Restore", commandType: ECommandTemplateKind.Backup, commandKind: "redis-restore" },
            { label: "Redis Restore (pipe)", commandType: ECommandTemplateKind.Backup, commandKind: "redis-restore.pipe" },
            { label: "SQLite Dump", commandType: ECommandTemplateKind.Backup, commandKind: "sqlite-dump" },
            { label: "SQLite Dump (pipe)", commandType: ECommandTemplateKind.Backup, commandKind: "sqlite-dump.pipe" },
            { label: "SQLite Restore", commandType: ECommandTemplateKind.Backup, commandKind: "sqlite-restore" },
            {
                label: "SQLite Restore (pipe)",
                commandType: ECommandTemplateKind.Backup,
                commandKind: "sqlite-restore.pipe",
            },
            { label: "SQL Server Dump", commandType: ECommandTemplateKind.Backup, commandKind: "sqlserver-dump" },
            {
                label: "SQL Server Dump (pipe)",
                commandType: ECommandTemplateKind.Backup,
                commandKind: "sqlserver-dump.pipe",
            },
            { label: "SQL Server Restore", commandType: ECommandTemplateKind.Backup, commandKind: "sqlserver-restore" },
            {
                label: "SQL Server Restore (pipe)",
                commandType: ECommandTemplateKind.Backup,
                commandKind: "sqlserver-restore.pipe",
            },
        ],
    },
    {
        heading: "Database",
        commandType: ECommandTemplateKind.Database,
        items: [
            {
                label: "ClickHouse Create Database",
                commandType: ECommandTemplateKind.Database,
                commandKind: "clickhouse_create_database",
            },
            {
                label: "MariaDB Create Database",
                commandType: ECommandTemplateKind.Database,
                commandKind: "mariadb_create_database",
            },
            {
                label: "MariaDB Create User",
                commandType: ECommandTemplateKind.Database,
                commandKind: "mariadb_create_user",
            },
            {
                label: "MongoDB Create User",
                commandType: ECommandTemplateKind.Database,
                commandKind: "mongo_create_user",
            },
            {
                label: "MongoDB Drop Database",
                commandType: ECommandTemplateKind.Database,
                commandKind: "mongo_drop_database",
            },
            {
                label: "MySQL Create Database",
                commandType: ECommandTemplateKind.Database,
                commandKind: "mysql_create_database",
            },
            {
                label: "MySQL Create User",
                commandType: ECommandTemplateKind.Database,
                commandKind: "mysql_create_user",
            },
            {
                label: "Postgres Create Database",
                commandType: ECommandTemplateKind.Database,
                commandKind: "pg_create_database",
            },
            {
                label: "Postgres Create User",
                commandType: ECommandTemplateKind.Database,
                commandKind: "pg_create_user",
            },
            {
                label: "Postgres Enable Extensions",
                commandType: ECommandTemplateKind.Database,
                commandKind: "pg_enable_extensions",
            },
            {
                label: "Redis FlushAll Async",
                commandType: ECommandTemplateKind.Database,
                commandKind: "redis_flushall_async",
            },
            {
                label: "Redis FlushDB Async",
                commandType: ECommandTemplateKind.Database,
                commandKind: "redis_flushdb_async",
            },
        ],
    },
    {
        heading: "Data-Ops",
        commandType: ECommandTemplateKind.DataOps,
        items: [
            {
                label: "Alembic Upgrade Head",
                commandType: ECommandTemplateKind.DataOps,
                commandKind: "alembic_upgrade_head",
            },
            {
                label: "ClickHouse Export Parquet",
                commandType: ECommandTemplateKind.DataOps,
                commandKind: "clickhouse_export_parquet",
            },
            {
                label: "ClickHouse Import Parquet",
                commandType: ECommandTemplateKind.DataOps,
                commandKind: "clickhouse_import_parquet",
            },
            { label: "Goose Down", commandType: ECommandTemplateKind.DataOps, commandKind: "goose_down" },
            { label: "Goose Up", commandType: ECommandTemplateKind.DataOps, commandKind: "goose_up" },
            { label: "Migrate Down", commandType: ECommandTemplateKind.DataOps, commandKind: "migrate_down" },
            { label: "Migrate Up", commandType: ECommandTemplateKind.DataOps, commandKind: "migrate_up" },
            {
                label: "MongoDB Export JSON",
                commandType: ECommandTemplateKind.DataOps,
                commandKind: "mongo_export_json",
            },
            {
                label: "MongoDB Import JSON",
                commandType: ECommandTemplateKind.DataOps,
                commandKind: "mongo_import_json",
            },
            {
                label: "MySQL Export CSV",
                commandType: ECommandTemplateKind.DataOps,
                commandKind: "mysql_export_csv",
            },
            {
                label: "MySQL Import CSV",
                commandType: ECommandTemplateKind.DataOps,
                commandKind: "mysql_import_csv",
            },
            {
                label: "Postgres Export CSV",
                commandType: ECommandTemplateKind.DataOps,
                commandKind: "pg_export_csv",
            },
            {
                label: "Postgres Import CSV",
                commandType: ECommandTemplateKind.DataOps,
                commandKind: "pg_import_csv",
            },
            {
                label: "Prisma Migrate Deploy",
                commandType: ECommandTemplateKind.DataOps,
                commandKind: "prisma_migrate_deploy",
            },
        ],
    },
    {
        heading: "Testing",
        commandType: ECommandTemplateKind.Testing,
        items: [
            {
                label: "ClickHouse Benchmark",
                commandType: ECommandTemplateKind.Testing,
                commandKind: "clickhouse_benchmark",
            },
            { label: "Hey Load Test", commandType: ECommandTemplateKind.Testing, commandKind: "hey_load_test" },
            { label: "Iperf3 Bandwidth", commandType: ECommandTemplateKind.Testing, commandKind: "iperf3_bandwidth" },
            { label: "K6 Run", commandType: ECommandTemplateKind.Testing, commandKind: "k6_run" },
            {
                label: "MongoDB Bench Insert",
                commandType: ECommandTemplateKind.Testing,
                commandKind: "mongo_bench_insert",
            },
            { label: "MySQL Slap Run", commandType: ECommandTemplateKind.Testing, commandKind: "mysqlslap_run" },
            {
                label: "MySQL Sysbench OLTP",
                commandType: ECommandTemplateKind.Testing,
                commandKind: "sysbench_mysql_oltp",
            },
            {
                label: "Postgres Bench Init",
                commandType: ECommandTemplateKind.Testing,
                commandKind: "pgbench_init",
            },
            { label: "Postgres Bench Run", commandType: ECommandTemplateKind.Testing, commandKind: "pgbench_run" },
            { label: "Redis Benchmark", commandType: ECommandTemplateKind.Testing, commandKind: "redis_benchmark" },
            {
                label: "Redis Benchmark Ping",
                commandType: ECommandTemplateKind.Testing,
                commandKind: "redis_benchmark_ping",
            },
        ],
    },
    {
        heading: "Diagnostics",
        commandType: ECommandTemplateKind.Diagnostics,
        items: [
            {
                label: "ClickHouse Processes",
                commandType: ECommandTemplateKind.Diagnostics,
                commandKind: "clickhouse_processes",
            },
            {
                label: "ElasticSearch Cluster Health",
                commandType: ECommandTemplateKind.Diagnostics,
                commandKind: "elasticsearch_cluster_health",
            },
            { label: "MariaDB Ping", commandType: ECommandTemplateKind.Diagnostics, commandKind: "mariadb_ping" },
            {
                label: "MariaDB Process List",
                commandType: ECommandTemplateKind.Diagnostics,
                commandKind: "mariadb_processlist",
            },
            {
                label: "MariaDB Table Size",
                commandType: ECommandTemplateKind.Diagnostics,
                commandKind: "mariadb_table_size",
            },
            { label: "MongoDB Stats", commandType: ECommandTemplateKind.Diagnostics, commandKind: "mongo_db_stats" },
            { label: "MongoDB Ping", commandType: ECommandTemplateKind.Diagnostics, commandKind: "mongo_ping" },
            {
                label: "MongoDB Server Status",
                commandType: ECommandTemplateKind.Diagnostics,
                commandKind: "mongo_server_status",
            },
            { label: "MySQL Ping", commandType: ECommandTemplateKind.Diagnostics, commandKind: "mysql_ping" },
            {
                label: "MySQL Process List",
                commandType: ECommandTemplateKind.Diagnostics,
                commandKind: "mysql_processlist",
            },
            { label: "MySQL Table Size", commandType: ECommandTemplateKind.Diagnostics, commandKind: "mysql_table_size" },
            { label: "Postgres Ping", commandType: ECommandTemplateKind.Diagnostics, commandKind: "pg_ping" },
            {
                label: "Postgres Stat Activity",
                commandType: ECommandTemplateKind.Diagnostics,
                commandKind: "pg_stat_activity",
            },
            { label: "Postgres Table Size", commandType: ECommandTemplateKind.Diagnostics, commandKind: "pg_table_size" },
            { label: "Redis BigKeys", commandType: ECommandTemplateKind.Diagnostics, commandKind: "redis_bigkeys" },
            { label: "Redis Info", commandType: ECommandTemplateKind.Diagnostics, commandKind: "redis_info" },
            { label: "Redis Ping", commandType: ECommandTemplateKind.Diagnostics, commandKind: "redis_ping" },
        ],
    },
    {
        heading: "Maintenance",
        commandType: ECommandTemplateKind.Maintenance,
        items: [
            {
                label: "ClickHouse Optimize Table",
                commandType: ECommandTemplateKind.Maintenance,
                commandKind: "clickhouse_optimize_table",
            },
            {
                label: "Docker Logs Cleanup",
                commandType: ECommandTemplateKind.Maintenance,
                commandKind: "docker_logs_cleanup",
            },
            { label: "Docker Prune All", commandType: ECommandTemplateKind.Maintenance, commandKind: "docker_prune_all" },
            {
                label: "ElasticSearch Clear Cache",
                commandType: ECommandTemplateKind.Maintenance,
                commandKind: "elasticsearch_clearcache",
            },
            {
                label: "ElasticSearch Force Merge",
                commandType: ECommandTemplateKind.Maintenance,
                commandKind: "elasticsearch_forcemerge",
            },
            { label: "MariaDB Analyze", commandType: ECommandTemplateKind.Maintenance, commandKind: "mariadb_analyze" },
            { label: "MariaDB Optimize", commandType: ECommandTemplateKind.Maintenance, commandKind: "mariadb_optimize" },
            { label: "MongoDB Compact", commandType: ECommandTemplateKind.Maintenance, commandKind: "mongo_compact" },
            { label: "MySQL Analyze", commandType: ECommandTemplateKind.Maintenance, commandKind: "mysql_analyze" },
            { label: "MySQL Optimize", commandType: ECommandTemplateKind.Maintenance, commandKind: "mysql_optimize" },
            {
                label: "Postgres Reindex Database",
                commandType: ECommandTemplateKind.Maintenance,
                commandKind: "pg_reindex_database",
            },
            {
                label: "Postgres Vacuum Analyze",
                commandType: ECommandTemplateKind.Maintenance,
                commandKind: "pg_vacuum_analyze",
            },
            { label: "Redis Bgsave", commandType: ECommandTemplateKind.Maintenance, commandKind: "redis_bgsave" },
            {
                label: "Redis Memory Purge",
                commandType: ECommandTemplateKind.Maintenance,
                commandKind: "redis_memory_purge",
            },
        ],
    },
] as const satisfies readonly CommandTemplateFromTemplateGroup[];
