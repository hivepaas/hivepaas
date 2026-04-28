import React, { useMemo } from "react";

import { Button } from "@components/ui";
import { DataTable } from "@components/ui/data-table";
import { Plus } from "lucide-react";
import type { AppStorageMount } from "~/projects/domain";

import { useStorageMounts } from "../../context";

import { createStorageTableColumns } from "./storage-table.defs";

type StorageMountWithId = AppStorageMount & { _id: string };

interface StorageTableProps {
    onAddMount: () => void;
    onEditMount: (mount: StorageMountWithId) => void;
}

function View({ onAddMount, onEditMount }: StorageTableProps) {
    const { mounts } = useStorageMounts();

    const columns = useMemo(
        () =>
            createStorageTableColumns(onEditMount, mount => {
                console.log("delete mount", mount);
            }),
        [onEditMount],
    );

    return (
        <div className="space-y-4">
            <div className="flex justify-between items-center">
                <h3 className="text-lg font-medium">Storage Mounts</h3>
                <Button
                    type="button"
                    variant="outline"
                    onClick={onAddMount}
                >
                    <Plus className="size-4 mr-2" />
                    New Storage
                </Button>
            </div>

            <DataTable
                columns={columns}
                data={mounts}
                enablePagination={false}
                enableSorting={false}
            />
        </div>
    );
}

export const StorageTable = React.memo(View);
