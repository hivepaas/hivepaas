import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingImService } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { EImServiceKind } from "@application/shared/enums";

import { ImPlatformMenuCell } from "./building-blocks";
import type { ImPlatformTableScope } from "./im-platform-table.types";

function formatKind(kind: SettingImService["kind"]): string {
    switch (kind) {
        case EImServiceKind.Slack:
            return "Slack";
        case EImServiceKind.Discord:
            return "Discord";
        default:
            return kind;
    }
}

function getKindClassName(kind: SettingImService["kind"]): string {
    switch (kind) {
        case EImServiceKind.Slack:
            return "bg-lime-300 text-white";
        case EImServiceKind.Discord:
            return "bg-blue-300 text-white";
        default:
            return "bg-muted text-muted-foreground";
    }
}

function createColumns(scope: ImPlatformTableScope): ColumnDef<SettingImService>[] {
    return [
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
                    <span
                        className={`inline-flex items-center rounded-md px-2.5 py-1 text-xs font-medium ${getKindClassName(
                            original.kind,
                        )}`}
                    >
                        {formatKind(original.kind)}
                    </span>
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
                    {scope.type === "project" && original.inherited && (
                        <span className="inline-flex items-center rounded-md bg-purple-500 px-2 py-1 text-xs font-medium text-white">
                            Inherited
                        </span>
                    )}
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
