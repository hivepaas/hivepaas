import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { AppHealthCheck } from "~/projects/domain";
import { EAppHealthCheckType } from "~/projects/module-shared/enums";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { EditCell, MenuCell, ViewTasksCell } from "./building-blocks";

function HealthCheckTypeBadge({ type }: { type: AppHealthCheck["healthcheckType"] }) {
    if (type === EAppHealthCheckType.REST) {
        return <Badge className="bg-yellow-400 text-white hover:bg-yellow-400/90">REST</Badge>;
    }

    return <Badge className="bg-purple-500 text-white hover:bg-purple-500/90">GRPC</Badge>;
}

function createColumns(projectId: string, appId: string): ColumnDef<AppHealthCheck>[] {
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
                    healthCheck={original}
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
            accessorKey: "interval",
            header: "Schedule",
            cell: ({ row: { original } }) => `every ${original.interval}`,
        },
        {
            accessorKey: "healthcheckType",
            header: "Type",
            cell: ({ row: { original } }) => <HealthCheckTypeBadge type={original.healthcheckType} />,
            meta: {
                align: "center",
                titleAlign: "center",
            },
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
            cell: ({ row: { original } }) => {
                if (!original.expireAt) {
                    return "-";
                }

                return format(original.expireAt, "yyyy-MM-dd HH:mm:ss");
            },
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row: { original } }) => (
                <MenuCell
                    projectId={projectId}
                    appId={appId}
                    healthCheck={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const AppHealthChecksTableDefs = Object.freeze({
    columns: createColumns,
});
