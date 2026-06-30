import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { AppSecret } from "~/projects/domain";
import { ProjectSecretStatusBadge } from "~/projects/module-shared/components";

import { EditCell, MenuCell } from "./building-blocks";

function createColumns(projectId: string, appId: string): ColumnDef<AppSecret>[] {
    return [
        {
            id: "view",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => {
                if (original.inherited) {
                    return null;
                }

                return (
                    <EditCell
                        projectId={projectId}
                        appId={appId}
                        secret={original}
                    />
                );
            },
            meta: {
                align: "center",
                titleAlign: "center",
            },
        },
        {
            accessorKey: "name",
            header: "Name",
        },
        {
            header: "Status",
            cell: ({ row: { original } }) => {
                const { status } = original;
                return (
                    <div className="flex items-center justify-center gap-2">
                        <ProjectSecretStatusBadge status={status} />
                        {original.inherited && <Badge className="bg-purple-500 text-white">Inherited</Badge>}
                    </div>
                );
            },
            meta: {
                align: "center",
                titleAlign: "center",
            },
        },
        {
            header: "Type",
            cell: ({ row: { original } }) => {
                return original.base64 ? "binary" : "text";
            },
            meta: {
                align: "center",
                titleAlign: "center",
            },
        },
        {
            header: "Mountpoint",
            cell: ({ row: { original } }) => {
                return original.swarmRef?.file?.name ?? "-";
            },
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
            cell: ({ row: { original } }) => {
                // Inherited secrets are read-only
                if (original.inherited) {
                    return null;
                }

                return (
                    <MenuCell
                        projectId={projectId}
                        appId={appId}
                        secret={original}
                    />
                );
            },
            meta: {
                align: "right",
            },
        },
    ];
}

export const AppSecretsTableDefs = Object.freeze({
    columns: createColumns,
});
