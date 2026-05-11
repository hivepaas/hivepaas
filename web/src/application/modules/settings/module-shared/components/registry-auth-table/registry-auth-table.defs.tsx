import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingRegistryAuth } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { RegistryAuthMenuCell } from "./building-blocks";
import type { RegistryAuthTableScope } from "./registry-auth-table.types";

function createColumns(scope: RegistryAuthTableScope): ColumnDef<SettingRegistryAuth>[] {
    return [
        {
            accessorKey: "name",
            header: "Name",
            enableSorting: true,
        },
        {
            accessorKey: "address",
            header: "Address",
            enableSorting: true,
        },
        {
            accessorKey: "username",
            header: "Username",
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
                <RegistryAuthMenuCell
                    scope={scope}
                    registryAuth={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const RegistryAuthTableDefs = Object.freeze({
    columns: createColumns,
});
