import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { ChevronDown, FileCode2, Plus } from "lucide-react";
import { Link } from "react-router";
import { ProjectsQueries } from "~/projects/data/queries";
import { useCreateProjectDialog } from "~/projects/dialogs/create-project";
import { ProjectsTableDefs } from "~/projects/module-shared/definitions/tables/projects/projects-table.defs";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { Button, DataTable } from "@/components/ui";

export function ProjectsTable() {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { data: { data: projects, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        ProjectsQueries.useFindManyPaginated({
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
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button disabled={isDenied}>
                                        <Plus /> New Project <ChevronDown className="size-4" />
                                    </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent
                                    align="end"
                                    className="min-w-[220px]"
                                >
                                    <DropdownMenuItem
                                        onSelect={() => {
                                            actions.open();
                                        }}
                                    >
                                        <Plus className="size-4" /> Empty project
                                    </DropdownMenuItem>
                                    <DropdownMenuItem asChild>
                                        <Link to={ROUTE.projects.newFromCompose.$route}>
                                            <FileCode2 className="size-4" /> From Docker Compose
                                        </Link>
                                    </DropdownMenuItem>
                                </DropdownMenuContent>
                            </DropdownMenu>
                        )}
                    </PermissionTooltipAction>
                }
            />
            <DataTable
                columns={ProjectsTableDefs.columns}
                data={projects}
                pageSize={pagination.size}
                enablePagination
                manualPagination
                totalCount={meta.page.total}
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
