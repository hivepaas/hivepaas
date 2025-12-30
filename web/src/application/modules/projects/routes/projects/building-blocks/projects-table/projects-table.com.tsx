import { Plus } from "lucide-react";
import { ProjectsQueries } from "~/projects/data/queries";
import { ProjectsTableDefs } from "~/projects/module-shared/definitions/tables/projects/projects-table.defs";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { Button, DataTable } from "@/components/ui";

export function ProjectsTable() {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { data: { data: projects } = DEFAULT_PAGINATED_DATA, isFetching } = ProjectsQueries.useFindManyPaginated({
        pagination,
        sorting,
        search,
    });

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <Button>
                        <Plus /> Create Project
                    </Button>
                }
            />
            <DataTable
                columns={ProjectsTableDefs.columns}
                data={projects}
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

