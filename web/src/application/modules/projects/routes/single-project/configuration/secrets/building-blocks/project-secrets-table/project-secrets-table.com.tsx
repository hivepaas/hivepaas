import { useMemo } from "react";

import { Plus } from "lucide-react";
import { ProjectSecretsQueries } from "~/projects/data/queries";
import { ProjectSecretsTableDefs } from "~/projects/module-shared/definitions/tables/project-secrets";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { Button, DataTable } from "@/components/ui";

export function ProjectSecretsTable({ projectId }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

    const { data: { data: secrets } = DEFAULT_PAGINATED_DATA, isFetching } = ProjectSecretsQueries.useFindManyPaginated(
        {
            projectID: projectId,
            pagination,
            sorting,
            search,
        },
    );

    const columns = useMemo(() => ProjectSecretsTableDefs.columns(projectId), [projectId]);

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
                                    navigate.modules(
                                        ROUTE.projects.single.providerConfiguration.secrets.create.$route(projectId),
                                    );
                                }}
                                type="button"
                                color="primary"
                                disabled={isDenied}
                            >
                                <Plus /> New Secret
                            </Button>
                        )}
                    </PermissionTooltipAction>
                }
            />
            <DataTable
                columns={columns}
                data={secrets}
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
