import { useMemo, useState } from "react";

import { SystemBackupFileQueries } from "~/system-settings/data";
import type { SystemBackupFile } from "~/system-settings/domain";
import { SystemBackupFilesTableDefs } from "~/system-settings/module-shared";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { BackupFileInfoDialog } from "../building-blocks";

export function SystemSettingsDataBackupBackupFilesRoute() {
    const [selectedFileID, setSelectedFileID] = useState<string>();
    const [infoDialogOpen, setInfoDialogOpen] = useState(false);
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();

    const { data: { data: backupFiles, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        SystemBackupFileQueries.useFindManyPaginated({
            pagination,
            sorting,
            search,
        });

    const columns = useMemo(() => {
        return SystemBackupFilesTableDefs.columns((file: SystemBackupFile) => {
            setSelectedFileID(file.id);
            setInfoDialogOpen(true);
        });
    }, []);

    return (
        <div className="flex flex-col gap-4">
            <TableActions search={{ value: search, onChange: setSearch }} />

            <DataTable
                columns={columns}
                data={backupFiles}
                pageSize={pagination.size}
                enablePagination
                manualPagination
                totalCount={meta.page.total}
                showPageSizeSelector={false}
                enableSorting
                manualSorting
                isLoading={isFetching}
                onPaginationChange={setPagination}
                onSortingChange={setSorting}
            />

            <BackupFileInfoDialog
                fileID={selectedFileID}
                open={infoDialogOpen}
                onOpenChange={setInfoDialogOpen}
            />
        </div>
    );
}
