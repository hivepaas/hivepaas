import { Badge, type BadgeTone } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingAccessToken } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { EAccessTokenKind } from "@application/shared/enums";

import type { AccessTokenTableScope } from "./access-token-table.types";
import { AccessTokenEditCell, AccessTokenMenuCell } from "./building-blocks";

function getAccessTokenKindTone(kind: SettingAccessToken["kind"]): BadgeTone {
    switch (kind) {
        case EAccessTokenKind.Bitbucket:
            return "blue";
        case EAccessTokenKind.Cloudflare:
            return "orange";
        case EAccessTokenKind.Github:
            return "sky";
        case EAccessTokenKind.Gitlab:
            return "indigo";
        case EAccessTokenKind.Gitea:
            return "fuchsia";
        case EAccessTokenKind.Gogs:
            return "amber";
        default:
            return "violet";
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
                return <Badge tone={getAccessTokenKindTone(original.kind)}>{original.kind}</Badge>;
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
