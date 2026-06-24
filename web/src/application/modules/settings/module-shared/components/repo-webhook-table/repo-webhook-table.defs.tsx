import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingRepoWebhook } from "~/settings/domain";
import { RepoWebhookKindBadge, SettingStatusBadge } from "~/settings/module-shared/components";

import { RepoWebhookEditCell, RepoWebhookMenuCell } from "./building-blocks";
import type { RepoWebhookTableScope } from "./repo-webhook-table.types";

function createColumns(scope: RepoWebhookTableScope): ColumnDef<SettingRepoWebhook>[] {
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
                <RepoWebhookEditCell
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
            cell: ({ row: { original } }) => <RepoWebhookKindBadge kind={original.kind} />,
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
                        <Badge className="bg-purple-500 text-white">Inherited</Badge>
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
                <RepoWebhookMenuCell
                    scope={scope}
                    repoWebhook={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const RepoWebhookTableDefs = Object.freeze({
    columns: createColumns,
});
