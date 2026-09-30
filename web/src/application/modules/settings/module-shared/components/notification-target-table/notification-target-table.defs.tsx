import { Badge, type BadgeTone } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingNotification } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { NotificationTargetEditCell, NotificationTargetMenuCell } from "./building-blocks";
import type { NotificationTargetTableScope } from "./notification-target-table.types";

/** The color of a target's badge, by the kind of target. */
const TARGET_TONES: Record<string, BadgeTone> = {
    email: "blue",
    slack: "purple",
    discord: "indigo",
    telegram: "sky",
    lark: "teal",
};

/** A kind of target added later, until it is given a color of its own. */
const OTHER_TARGET_TONE: BadgeTone = "amber";

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
                    tone={TARGET_TONES[target] ?? OTHER_TARGET_TONE}
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
                    {scope.type === "project" && original.inherited && <Badge tone="purple">Inherited</Badge>}
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
