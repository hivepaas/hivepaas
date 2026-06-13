import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingSslProvider } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";
import { formatSslProviderKind } from "~/settings/module-shared/constants/ssl-provider.constants";

import { SslProviderEditCell, SslProviderMenuCell } from "./building-blocks";
import type { SslProviderTableScope } from "./ssl-provider-table.types";

function createColumns(scope: SslProviderTableScope): ColumnDef<SettingSslProvider>[] {
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
                <SslProviderEditCell
                    scope={scope}
                    id={original.id}
                    inherited={original.inherited}
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
            header: "Provider",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => (
                <div className="flex justify-center">
                    <Badge className="bg-emerald-300 text-white">{formatSslProviderKind(original.kind)}</Badge>
                </div>
            ),
        },
        {
            accessorKey: "email",
            header: "E-mail",
            enableSorting: true,
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
                <SslProviderMenuCell
                    scope={scope}
                    sslProvider={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const SslProviderTableDefs = Object.freeze({
    columns: createColumns,
});
