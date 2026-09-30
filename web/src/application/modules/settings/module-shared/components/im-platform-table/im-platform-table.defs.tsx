import { Badge, type BadgeTone } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingImService } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { EImServiceKind } from "@application/shared/enums";

import { ImPlatformEditCell, ImPlatformMenuCell } from "./building-blocks";
import type { ImPlatformTableScope } from "./im-platform-table.types";

function formatKind(kind: SettingImService["kind"]): string {
    switch (kind) {
        case EImServiceKind.Slack:
            return "Slack";
        case EImServiceKind.Discord:
            return "Discord";
        case EImServiceKind.Telegram:
            return "Telegram";
        case EImServiceKind.Lark:
            return "Lark";
        default:
            return kind;
    }
}

function getKindTone(kind: SettingImService["kind"]): BadgeTone {
    switch (kind) {
        case EImServiceKind.Slack:
            return "lime";
        case EImServiceKind.Discord:
            return "blue";
        case EImServiceKind.Telegram:
            return "cyan";
        case EImServiceKind.Lark:
            return "sky";
        default:
            return "neutral";
    }
}

function createColumns(scope: ImPlatformTableScope): ColumnDef<SettingImService>[] {
    return [
        {
            id: "view",
            accessorKey: "inherited",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <ImPlatformEditCell
                    scope={scope}
                    id={original.id}
                />
            ),
            meta: {
                align: "center",
                titleAlign: "center",
            },
        },
        {
            accessorKey: "name",
            header: "Name",
            enableSorting: true,
        },
        {
            accessorKey: "kind",
            header: "Type",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => (
                <div className="flex justify-center">
                    <Badge tone={getKindTone(original.kind)}>{formatKind(original.kind)}</Badge>
                </div>
            ),
        },
        {
            accessorKey: "status",
            header: "Status",
            meta: {
                align: "center",
                titleAlign: "center",
            },
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
                <ImPlatformMenuCell
                    scope={scope}
                    imPlatform={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const ImPlatformTableDefs = Object.freeze({
    columns: createColumns,
});
