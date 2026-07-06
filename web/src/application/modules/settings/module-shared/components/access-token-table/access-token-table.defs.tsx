import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingAccessToken } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { EAccessTokenKind } from "@application/shared/enums";

import type { AccessTokenTableScope } from "./access-token-table.types";
import { AccessTokenEditCell, AccessTokenMenuCell } from "./building-blocks";

function getAccessTokenKindClassName(kind: SettingAccessToken["kind"]): string {
    switch (kind) {
        case EAccessTokenKind.Bitbucket:
            return "bg-blue-600 text-white";
        case EAccessTokenKind.Cloudflare:
            return "bg-orange-500 text-white";
        case EAccessTokenKind.Github:
            return "bg-sky-500 text-white";
        case EAccessTokenKind.Gitlab:
            return "bg-indigo-500 text-white";
        case EAccessTokenKind.Gitea:
            return "bg-fuchsia-500 text-white";
        case EAccessTokenKind.Gogs:
            return "bg-amber-400 text-white";
        default:
            return "bg-violet-500 text-white";
    }
}

function createColumns(scope: AccessTokenTableScope): ColumnDef<SettingAccessToken>[] {
    return [
        {
            id: "view",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <AccessTokenEditCell
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
            accessorKey: "kind",
            header: "Type",
            cell: ({ row: { original } }) => {
                if (!original.kind) return "-";
                return <Badge className={getAccessTokenKindClassName(original.kind)}>{original.kind}</Badge>;
            },
        },
        {
            accessorKey: "user",
            header: "User",
            enableSorting: true,
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
                <AccessTokenMenuCell
                    scope={scope}
                    accessToken={original}
                />
            ),
            meta: { align: "right" },
        },
    ];
}

export const AccessTokenTableDefs = Object.freeze({
    columns: createColumns,
});
