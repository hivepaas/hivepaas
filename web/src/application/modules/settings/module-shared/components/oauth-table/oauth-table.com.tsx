import { useMemo } from "react";

import { Plus } from "lucide-react";
import { OAuthQueries } from "~/settings/data/queries";
import { useCreateOrEditOAuthDialog } from "~/settings/dialogs/create-or-edit-oauth";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { Button, DataTable } from "@/components/ui";

import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { OAuthTableDefs } from "./oauth-table.defs";

export function SettingsOAuthTable() {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const createOrEditDialog = useCreateOrEditOAuthDialog();

    const { data: { data: oauthItems, meta } = DEFAULT_PAGINATED_DATA, isFetching } = OAuthQueries.useFindManyPaginated(
        { pagination, sorting, search },
    );
    const columns = useMemo(() => OAuthTableDefs.columns(), []);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <PermissionTooltipAction
                        id={MODULE_IDS.Settings}
                        action="write"
                    >
                        {({ isDenied }) => (
                            <Button
                                onClick={() => {
                                    createOrEditDialog.actions.open();
                                }}
                                disabled={isDenied}
                            >
                                <Plus className="size-4" />
                                New OAuth
                            </Button>
                        )}
                    </PermissionTooltipAction>
                }
            />
            <DataTable
                columns={columns}
                data={oauthItems}
                pageSize={pagination.size}
                manualPagination
                totalCount={meta.page.total}
                manualSorting
                enableSorting
                enablePagination
                isLoading={isFetching}
                onPaginationChange={setPagination}
                onSortingChange={setSorting}
                showPageSizeSelector={false}
            />
        </div>
    );
}
