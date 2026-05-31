import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { AppScheduledJob } from "~/projects/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { EditCell, MenuCell, ViewTasksCell } from "./building-blocks";

function formatDate(date: Date | null | undefined) {
    if (!date) {
        return "-";
    }

    return format(date, "yyyy-MM-dd HH:mm:ss");
}

function formatSchedule(scheduledJob: AppScheduledJob) {
    if (scheduledJob.schedule.interval) {
        return `every ${scheduledJob.schedule.interval}`;
    }

    if (scheduledJob.schedule.cronExpr) {
        return `cron: ${scheduledJob.schedule.cronExpr}`;
    }

    return "-";
}

function createColumns(projectId: string, appId: string): ColumnDef<AppScheduledJob>[] {
    return [
        {
            id: "view",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <EditCell
                    projectId={projectId}
                    appId={appId}
                    scheduledJob={original}
                />
            ),
            meta: {
                align: "center",
                titleAlign: "center",
            },
        },
        {
            id: "tasks",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 120,
            size: 120,
            cell: () => <ViewTasksCell />,
            meta: {
                align: "left",
                titleAlign: "left",
            },
        },
        {
            accessorKey: "name",
            header: "Name",
        },
        {
            accessorKey: "schedule",
            header: "Schedule",
            enableSorting: false,
            cell: ({ row: { original } }) => formatSchedule(original),
        },
        {
            accessorKey: "nextRuns",
            header: "Next Run",
            enableSorting: false,
            cell: ({ row: { original } }) => formatDate(original.nextRuns[0]),
        },
        {
            accessorKey: "status",
            header: "Status",
            cell: ({ row: { original } }) => <SettingStatusBadge status={original.status} />,
            meta: {
                align: "center",
                titleAlign: "center",
            },
        },
        {
            accessorKey: "expireAt",
            header: "Expire At",
            cell: ({ row: { original } }) => formatDate(original.expireAt),
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row: { original } }) => (
                <MenuCell
                    projectId={projectId}
                    appId={appId}
                    scheduledJob={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const AppScheduledJobsTableDefs = Object.freeze({
    columns: createColumns,
});
