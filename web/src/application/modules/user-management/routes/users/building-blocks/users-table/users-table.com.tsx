import { Plus } from "lucide-react";
import { UsersQueries } from "~/user-management/data/queries";
import { useInviteUserDialog } from "~/user-management/dialogs";
import { UsersTableDefs } from "~/user-management/module-shared/definitions/tables";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { Button, DataTable } from "@/components/ui";

export function UsersTable() {
    const inviteUserDialog = useInviteUserDialog({
        onClose: () => {
            inviteUserDialog.actions.close();
        },
    });
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
                    <PermissionTooltipAction
                        id={MODULE_IDS.User}
                        action="write"
                    >
                        {({ isDenied }) => (
                            <Button
                                onClick={inviteUserDialog.actions.open}
                                disabled={isDenied}
                            >
                                <Plus /> Invite User
                            </Button>
                        )}
                    </PermissionTooltipAction>
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
