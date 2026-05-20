import React, { useMemo, useState } from "react";

import { Button, Input } from "@components/ui";
import { DataTable } from "@components/ui/data-table";
import { Plus, SearchIcon } from "lucide-react";
import type { AppStorageMount } from "~/projects/domain";

import { useStorageMounts } from "../../context";

import { createStorageTableColumns } from "./storage-table.defs";

type StorageMountWithId = AppStorageMount & { _id: string };

interface StorageTableProps {
    onAddMount: () => void;
    onEditMount: (mount: StorageMountWithId) => void;
    onDeleteMount: (mount: StorageMountWithId) => Promise<void> | void;
}

function View({ onAddMount, onEditMount, onDeleteMount }: StorageTableProps) {
    const { mounts } = useStorageMounts();
    const [internalSearch, setInternalSearch] = useState("");

    const columns = useMemo(() => createStorageTableColumns(onEditMount, onDeleteMount), [onDeleteMount, onEditMount]);

    const filteredMounts = useMemo(() => {
        return mounts.filter(
            mount =>
                (mount.type?.toLowerCase().includes(internalSearch.toLowerCase()) ?? false) ||
                (mount.target?.toLowerCase().includes(internalSearch.toLowerCase()) ?? false),
        );
    }, [mounts, internalSearch]);

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between gap-2">
                <div className="relative">
                    <div className="text-muted-foreground pointer-events-none absolute inset-y-0 left-0 flex items-center justify-center pl-3 peer-disabled:opacity-50">
                        <SearchIcon className="size-4" />
                        <span className="sr-only">Search</span>
                    </div>
                    <Input
                        value={internalSearch}
                        onChange={e => {
                            setInternalSearch(e.target.value);
                        }}
                        type="search"
                        placeholder="Search"
                        className="peer px-9 [&::-webkit-search-cancel-button]:appearance-none [&::-webkit-search-decoration]:appearance-none [&::-webkit-search-results-button]:appearance-none [&::-webkit-search-results-decoration]:appearance-none"
                    />
                </div>
                <Button
                    type="button"
                    color="primary"
                    onClick={onAddMount}
                >
                    <Plus className="size-4" />
                    New Storage
                </Button>
            </div>

            <DataTable
                columns={columns}
                data={filteredMounts}
                enablePagination={false}
                enableSorting={false}
            />
        </div>
    );
}

export const StorageTable = React.memo(View);
