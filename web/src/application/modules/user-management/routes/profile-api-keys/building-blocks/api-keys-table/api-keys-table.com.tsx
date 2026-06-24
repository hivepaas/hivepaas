import { Plus } from "lucide-react";
import { ProfileApiKeysTableDefs } from "~/user-management/module-shared/definitions/tables/profile-api-keys";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { ProfileQueries } from "@application/shared/data/queries";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { Button, DataTable } from "@/components/ui";

export function ApiKeysTable() {
    const { navigate } = useAppNavigate();
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { data: { data: apiKeys, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        ProfileQueries.useFindManyApiKeysPaginated({
            pagination,
            sorting,
            search,
        });

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <Button
                        onClick={() => {
                            navigate.modules(ROUTE.currentUser.profileApiKeys.create.$route);
                        }}
                    >
                        <Plus /> New API Key
                    </Button>
                }
            />
            <DataTable
                columns={ProfileApiKeysTableDefs.columns}
                data={apiKeys}
                pageSize={pagination.size}
                enablePagination
                manualPagination
                manualSorting
                enableSorting
                totalCount={meta.page.total}
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
