import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingNotification } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { NotificationTargetEditCell, NotificationTargetMenuCell } from "./building-blocks";
import type { NotificationTargetTableScope } from "./notification-target-table.types";

function getTargets(notification: SettingNotification): string {
    const targets: string[] = [];

    if (notification.viaEmail?.enabled) {
        targets.push("email");
    }
    if (notification.viaSlack?.enabled) {
        targets.push("slack");
    }
    if (notification.viaDiscord?.enabled) {
        targets.push("discord");
    }

    return targets.length > 0 ? targets.join(", ") : "-";
}

function createColumns(scope: NotificationTargetTableScope): ColumnDef<SettingNotification>[] {
    return [
        {
            id: "view",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <NotificationTargetEditCell
                    scope={scope}
                    id={original.id}
                    inherited={original.inherited}
                />
            ),
            meta: { align: "center", titleAlign: "center" },
        },
        {
            accessorKey: "name",
            header: "Name",
            enableSorting: true,
        },
        {
            id: "targets",
            header: "Targets",
            cell: ({ row: { original } }) => getTargets(original),
        },
        {
            accessorKey: "status",
            header: "Status",
            meta: { align: "center", titleAlign: "center" },
            cell: ({ row: { original } }) => (
                <div className="flex items-center justify-center gap-2">
                    <SettingStatusBadge status={original.status} />
                    {scope.type === "project" && original.inherited && (
                        <Badge className="bg-purple-500 text-white">Inherited</Badge>
                    )}
                </div>
            ),
        },
        {
            accessorKey: "expireAt",
            header: "Expire At",
            cell: ({ row: { original } }) => {
                if (!original.expireAt) return "-";
                return format(original.expireAt, "yyyy-MM-dd HH:mm:ss");
            },
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row: { original } }) => (
                <NotificationTargetMenuCell
                    scope={scope}
                    notificationTarget={original}
                />
            ),
            meta: { align: "right" },
        },
    ];
}

export const NotificationTargetTableDefs = Object.freeze({
    columns: createColumns,
});
