import { useMemo } from "react";

import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { AppDataFilesQueries } from "~/projects/data/queries";
import { AppDataFilesTableDefs } from "~/projects/module-shared/definitions/tables/app-data-files";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

export function AppDataFilesRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { data: { data: dataFiles, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        AppDataFilesQueries.useFindManyPaginated(
            {
                projectID: projectId,
                appID: appId,
                pagination,
                sorting,
                search,
            },
            APP_CONFIGURATION_QUERY_OPTIONS,
        );

    const columns = useMemo(() => AppDataFilesTableDefs.columns(projectId, appId), [projectId, appId]);

    return (
        <div className="flex flex-col gap-6">
            <TableActions search={{ value: search, onChange: setSearch }} />

            <DataTable
                columns={columns}
                data={dataFiles}
                isLoading={isFetching}
                pageSize={pagination.size}
                manualPagination
                totalCount={meta.page.total}
                onPaginationChange={setPagination}
                manualSorting
                onSortingChange={setSorting}
                enablePagination
                showPageSizeSelector={false}
            />
        </div>
    );
}
