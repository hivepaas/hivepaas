import { useMemo } from "react";
import { ProjectAppsQueries } from "~/projects/data/queries";
import { ProjectAppsTableDefs } from "~/projects/module-shared/definitions/tables/project-apps";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

export function ProjectAppsTable({ projectId }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();

    const { data: { data: apps } = DEFAULT_PAGINATED_DATA, isFetching } = ProjectAppsQueries.useFindManyPaginated({
        projectID: projectId,
        pagination,
        sorting,
        search,
    });

    const columns = useMemo(() => ProjectAppsTableDefs.columns(projectId), [projectId]);

    return (
        <div className="flex flex-col gap-4">
            <TableActions search={{ value: search, onChange: setSearch }} />
            <DataTable
                columns={columns}
                data={apps}
                pageSize={pagination.size}
                enablePagination
                manualSorting
                enableSorting
                isLoading={isFetching}
                onPaginationChange={value => {
                    setPagination(value);
                }}
                onSortingChange={value => {
                    setSorting(value);
                }}
            />
        </div>
    );
}

interface Props {
    projectId: string;
}
