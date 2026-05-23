import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingSSHKey } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { SSHKeyEditCell, SSHKeyMenuCell } from "./building-blocks";
import type { SSHKeyTableScope } from "./ssh-key-table.types";

function createColumns(scope: SSHKeyTableScope): ColumnDef<SettingSSHKey>[] {
    return [
        {
            id: "view",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <SSHKeyEditCell
                    scope={scope}
                    id={original.id}
                    inherited={original.inherited}
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
            accessorKey: "targets",
            header: "Targets",
            cell: ({ row: { original } }) => {
                if (!original.targets?.length) return "-";
                return original.targets.join(", ");
            },
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
                <SSHKeyMenuCell
                    scope={scope}
                    sshKey={original}
                />
            ),
            meta: { align: "right" },
        },
    ];
}

export const SSHKeyTableDefs = Object.freeze({
    columns: createColumns,
});
