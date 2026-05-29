import { Plus } from "lucide-react";
import { ProjectsQueries } from "~/projects/data/queries";
import { useCreateProjectDialog } from "~/projects/dialogs/create-project";
import { ProjectsTableDefs } from "~/projects/module-shared/definitions/tables/projects/projects-table.defs";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { Button, DataTable } from "@/components/ui";

export function ProjectsTable() {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { data: { data: projects } = DEFAULT_PAGINATED_DATA, isFetching } = ProjectsQueries.useFindManyPaginated({
        pagination,
        sorting,
        search,
    });

    const { actions } = useCreateProjectDialog({
        onClose: () => {
            actions.close();
        },
    });

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <PermissionTooltipAction
                        id={MODULE_IDS.Project}
                        action="write"
                    >
                        {({ isDenied }) => (
                            <Button
                                onClick={() => {
                                    actions.open();
                                }}
                                disabled={isDenied}
                            >
                                <Plus /> New Project
                            </Button>
                        )}
                    </PermissionTooltipAction>
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
