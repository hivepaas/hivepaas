import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingCloudStorage } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { CloudStorageEditCell, CloudStorageMenuCell } from "./building-blocks";
import type { CloudStorageTableScope } from "./cloud-storage-table.types";

function createColumns(scope: CloudStorageTableScope): ColumnDef<SettingCloudStorage>[] {
    return [
        {
            id: "view",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <CloudStorageEditCell
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
            header: "Provider",
            cell: ({ row: { original } }) => {
                if (!original.kind) return "-";
                return <Badge className="bg-blue-500 text-white">{original.kind}</Badge>;
            },
        },
        {
            accessorKey: "s3.region",
            header: "Region",
            cell: ({ row: { original } }) => original.s3.region || "-",
        },
        {
            accessorKey: "s3.bucket",
            header: "Bucket",
            cell: ({ row: { original } }) => original.s3.bucket || "-",
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
                <CloudStorageMenuCell
                    scope={scope}
                    cloudStorage={original}
                />
            ),
            meta: { align: "right" },
        },
    ];
}

export const CloudStorageTableDefs = Object.freeze({
    columns: createColumns,
});
