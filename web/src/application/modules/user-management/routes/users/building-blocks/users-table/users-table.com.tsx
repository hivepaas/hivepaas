import { Plus } from "lucide-react";
import { UsersQueries } from "~/user-management/data/queries";
import { UsersTableDefs } from "~/user-management/module-shared/definitions/tables";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { Button, DataTable } from "@/components/ui";

export function UsersTable() {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { data: { data: users } = DEFAULT_PAGINATED_DATA, isFetching } = UsersQueries.useFindManyPaginated({
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
                        <Plus /> Invite User
                    </Button>
                }
            />
            <DataTable
                columns={UsersTableDefs.columns}
                data={users}
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
