import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import { EyeIcon } from "lucide-react";
import { SettingStatusBadge } from "~/settings/module-shared/components";
import type { SystemBackupFile } from "~/system-settings/domain";

import { Button } from "@/components/ui";

import { BackupFileMenuCell } from "./building-blocks";
import { formatBackupFileSize, getBackupFileStorageLabel } from "./utils";

function createColumns(onView: (file: SystemBackupFile) => void): ColumnDef<SystemBackupFile>[] {
    return [
        {
            id: "view",
            header: "",
            minSize: 64,
            size: 64,
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => {
                return (
                    <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-link hover:opacity-50"
                        onClick={() => {
                            onView(original);
                        }}
                    >
                        <EyeIcon className="size-5" />
                        <span className="sr-only">View backup file info</span>
                    </Button>
                );
            },
        },
        {
            accessorKey: "name",
            header: "Name",
        },
        {
            accessorKey: "size",
            header: "Size",
            cell: ({ row: { original } }) => formatBackupFileSize(original.size),
        },
        {
            accessorKey: "storage",
            header: "Storage",
            cell: ({ row: { original } }) => getBackupFileStorageLabel(original),
        },
        {
            accessorKey: "status",
            header: "Status",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => <SettingStatusBadge status={original.status} />,
        },
        {
            accessorKey: "createdAt",
            header: "Created At",
            cell: ({ row: { original } }) => format(original.createdAt, "yyyy-MM-dd HH:mm:ss"),
        },
        {
            id: "actions",
            header: "",
            minSize: 64,
            size: 64,
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => <BackupFileMenuCell file={original} />,
        },
    ];
}

export const SystemBackupFilesTableDefs = Object.freeze({
    columns: createColumns,
});
