import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import { ERegistryAuthKind, type SettingRegistryAuth } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { RegistryAuthEditCell, RegistryAuthMenuCell } from "./building-blocks";
import type { RegistryAuthTableScope } from "./registry-auth-table.types";

function createColumns(scope: RegistryAuthTableScope): ColumnDef<SettingRegistryAuth>[] {
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
                <RegistryAuthEditCell
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
            accessorKey: "address",
            header: "Address",
            enableSorting: true,
            cell: ({ row: { original } }) =>
                original.address ? (
                    <Badge
                        tone="neutral"
                        className="font-mono max-w-full whitespace-normal break-all"
                    >
                        {original.address}
                    </Badge>
                ) : (
                    "-"
                ),
        },
        {
            accessorKey: "username",
            header: "Username",
            enableSorting: true,
            cell: ({ row: { original } }) =>
                original.kind === ERegistryAuthKind.AwsEcr ? (
                    <Badge tone="amber">Amazon ECR</Badge>
                ) : (
                    original.username || "-"
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
