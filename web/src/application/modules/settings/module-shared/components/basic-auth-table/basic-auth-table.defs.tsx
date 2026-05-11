import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingBasicAuth } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import type { BasicAuthTableScope } from "./basic-auth-table.types";
import { BasicAuthMenuCell } from "./building-blocks";

function createColumns(scope: BasicAuthTableScope): ColumnDef<SettingBasicAuth>[] {
    return [
        {
            accessorKey: "name",
            header: "Name",
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
                <BasicAuthMenuCell
                    scope={scope}
                    basicAuth={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const BasicAuthTableDefs = Object.freeze({
    columns: createColumns,
});
