import { ECommandTemplateKind } from "@application/shared/enums";

export type CommandPipeFromTemplateSelection = {
    commandType: string;
    commandKind: string;
};

export type CommandPipeFromTemplateItem = CommandPipeFromTemplateSelection & {
    label: string;
};

export type CommandPipeFromTemplateGroup = {
    heading: string;
    commandType: string;
    items: CommandPipeFromTemplateItem[];
};

export const COMMAND_PIPE_FROM_TEMPLATE_CATALOG = [
    {
        heading: "Backup",
        commandType: ECommandTemplateKind.Backup,
        items: [
            { label: "ClickHouse", commandType: ECommandTemplateKind.Backup, commandKind: "clickhouse" },
            { label: "Dolt", commandType: ECommandTemplateKind.Backup, commandKind: "dolt" },
            { label: "ElasticSearch", commandType: ECommandTemplateKind.Backup, commandKind: "elasticsearch" },
            { label: "InfluxDB", commandType: ECommandTemplateKind.Backup, commandKind: "influx" },
            { label: "MariaDB", commandType: ECommandTemplateKind.Backup, commandKind: "mariadb" },
            { label: "MongoDB", commandType: ECommandTemplateKind.Backup, commandKind: "mongodb" },
            { label: "MySQL", commandType: ECommandTemplateKind.Backup, commandKind: "mysql" },
            { label: "Neon", commandType: ECommandTemplateKind.Backup, commandKind: "neon" },
            { label: "Postgres", commandType: ECommandTemplateKind.Backup, commandKind: "postgres" },
            { label: "Redis", commandType: ECommandTemplateKind.Backup, commandKind: "redis" },
            { label: "SQLite", commandType: ECommandTemplateKind.Backup, commandKind: "sqlite" },
            { label: "SQL Server", commandType: ECommandTemplateKind.Backup, commandKind: "sqlserver" },
        ],
    },
] as const satisfies readonly CommandPipeFromTemplateGroup[];
