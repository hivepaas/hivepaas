import { NodesQueries } from "~/cluster/data/queries";
import { NodesTableDefs } from "~/cluster/module-shared/definitions/tables/nodes/nodes-table.defs";

import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

export function NodesTable() {
    const { pagination, setPagination, sorting, setSorting, search } = useTableState();
    const { data: { data: nodes } = DEFAULT_PAGINATED_DATA, isFetching } = NodesQueries.useFindManyPaginated({
        pagination,
        sorting,
        search,
    });

    return (
        <div className="flex flex-col gap-4">
            <DataTable
                columns={NodesTableDefs.columns}
                data={nodes}
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
