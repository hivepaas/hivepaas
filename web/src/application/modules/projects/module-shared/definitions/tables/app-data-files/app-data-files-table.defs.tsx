import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { AppDataFile } from "~/projects/domain";

import { getFriendlyDataSize } from "@application/shared/utils/data-size";

import { AppDataFileMenuCell } from "./building-blocks";

const MAX_STORAGE_LABEL_LENGTH = 20;

function getStorageLabel(file: AppDataFile): string {
    const label = file.storageType === "local" ? "local" : (file.storage?.name ?? "-");

    if (label.length <= MAX_STORAGE_LABEL_LENGTH) {
        return label;
    }

    return `${label.slice(0, MAX_STORAGE_LABEL_LENGTH - 3)}...`;
}

function createColumns(projectId: string, appId: string): ColumnDef<AppDataFile>[] {
    return [
        {
            accessorKey: "name",
            header: "Name",
        },
        {
            accessorKey: "type",
            header: "Type",
            enableSorting: false,
            cell: ({ row: { original } }) => <Badge className="bg-sky-500 text-white">{original.type}</Badge>,
            meta: {
                align: "center",
                titleAlign: "center",
            },
        },
        {
            accessorKey: "kind",
            header: "Kind",
            enableSorting: false,
            cell: ({ row: { original } }) => {
                if (!original.kind) {
                    return "-";
                }

                return <Badge className="bg-indigo-500 text-white">{original.kind}</Badge>;
            },
            meta: {
                align: "center",
                titleAlign: "center",
            },
        },
        {
            accessorKey: "sizeBytes",
            header: "Size",
            enableSorting: false,
            cell: ({ row: { original } }) => getFriendlyDataSize(original.sizeBytes) || "-",
        },
        {
            id: "storage",
            header: "Storage",
            enableSorting: false,
            cell: ({ row: { original } }) => getStorageLabel(original),
        },
        {
            accessorKey: "createdAt",
            header: "Created At",
            enableSorting: false,
            cell: ({ row: { original } }) => format(original.createdAt, "yyyy-MM-dd HH:mm:ss"),
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row: { original } }) => (
                <AppDataFileMenuCell
                    projectId={projectId}
                    appId={appId}
                    dataFile={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const AppDataFilesTableDefs = Object.freeze({
    columns: createColumns,
});
