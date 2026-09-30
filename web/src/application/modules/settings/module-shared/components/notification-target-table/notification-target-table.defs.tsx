import { Badge } from "@components/ui/badge";
import { cn } from "@lib/utils";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingNotification } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { NotificationTargetEditCell, NotificationTargetMenuCell } from "./building-blocks";
import type { NotificationTargetTableScope } from "./notification-target-table.types";

/** The colors of a target's badge, by the kind of target. */
const TARGET_CLASS_NAMES: Record<string, string> = {
    email: "bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/30",
    slack: "bg-purple-500/10 text-purple-700 dark:text-purple-400 border-purple-500/30",
    discord: "bg-indigo-500/10 text-indigo-700 dark:text-indigo-400 border-indigo-500/30",
    telegram: "bg-sky-500/10 text-sky-700 dark:text-sky-400 border-sky-500/30",
    lark: "bg-teal-500/10 text-teal-700 dark:text-teal-400 border-teal-500/30",
};

/** A kind of target added later, until it is given a color of its own. */
const OTHER_TARGET_CLASS_NAME = "bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/30";

function getTargets(notification: SettingNotification): string[] {
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
    if (notification.viaTelegram?.enabled) {
        targets.push("telegram");
    }
    if (notification.viaLark?.enabled) {
        targets.push("lark");
    }

    return targets;
}

function TargetBadges({ notification }: { notification: SettingNotification }) {
    const targets = getTargets(notification);

    if (targets.length === 0) {
        return "-";
    }

    return (
        <div className="flex flex-wrap gap-1.5">
            {targets.map(target => (
                <Badge
                    key={target}
                    variant="outline"
                    className={cn("text-xs font-medium", TARGET_CLASS_NAMES[target] ?? OTHER_TARGET_CLASS_NAME)}
                >
                    {target}
                </Badge>
            ))}
        </div>
    );
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
            cell: ({ row: { original } }) => <TargetBadges notification={original} />,
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
