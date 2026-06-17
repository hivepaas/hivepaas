import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingAcmeDnsProvider } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";
import { formatAcmeDnsProviderKind } from "~/settings/module-shared/constants/acme-dns-provider.constants";

import { AcmeDnsProviderEditCell, AcmeDnsProviderMenuCell } from "./building-blocks";
import type { AcmeDnsProviderTableScope } from "./acme-dns-provider-table.types";

function createColumns(scope: AcmeDnsProviderTableScope): ColumnDef<SettingAcmeDnsProvider>[] {
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
                <AcmeDnsProviderEditCell
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
            cell: ({ row: { original } }) => formatAcmeDnsProviderKind(original.kind),
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
                <AcmeDnsProviderMenuCell
                    scope={scope}
                    acmeDnsProvider={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const AcmeDnsProviderTableDefs = Object.freeze({
    columns: createColumns,
});
