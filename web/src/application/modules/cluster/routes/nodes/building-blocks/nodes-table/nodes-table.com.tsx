import { Plus } from "lucide-react";
import { NodesQueries } from "~/cluster/data/queries";
import { NodesTableDefs } from "~/cluster/module-shared/definitions/tables/nodes/nodes-table.defs";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { useJoinNewNodeDialog, useSetManagerNodesDialog } from "@application/modules/cluster/dialogs";

import { Button, DataTable } from "@/components/ui";

export function NodesTable() {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { data: { data: nodes } = DEFAULT_PAGINATED_DATA, isFetching } = NodesQueries.useFindManyPaginated({
        pagination,
        sorting,
        search,
    });

    const joinNewNodeDialog = useJoinNewNodeDialog({
        onClose: () => {
            joinNewNodeDialog.actions.close();
        },
    });

    const setManagerNodesDialog = useSetManagerNodesDialog({
        onClose: () => {
            setManagerNodesDialog.actions.close();
        },
    });

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <PermissionTooltipAction
                        id={MODULE_IDS.Cluster}
                        action="write"
                    >
                        {({ isDenied }) => (
                            <div className="flex items-center gap-2">
                                <Button
                                    variant="outline"
                                    onClick={setManagerNodesDialog.actions.open}
                                    disabled={isDenied}
                                >
                                    Set Manager Nodes
                                </Button>
                                <Button
                                    onClick={joinNewNodeDialog.actions.open}
                                    disabled={isDenied}
                                >
                                    <Plus /> Join Node
                                </Button>
                            </div>
                        )}
                    </PermissionTooltipAction>
                }
            />
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
