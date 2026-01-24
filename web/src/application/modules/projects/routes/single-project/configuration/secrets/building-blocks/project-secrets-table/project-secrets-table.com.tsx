import { useMemo } from "react";
import { Plus } from "lucide-react";
import { ProjectSecretsQueries } from "~/projects/data/queries";
import { useCreateOrEditProjectSecretDialog } from "~/projects/dialogs/create-or-edit-project-secret";
import { ProjectSecretsTableDefs } from "~/projects/module-shared/definitions/tables/project-secrets";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { Button, DataTable } from "@/components/ui";

export function ProjectSecretsTable({ projectId }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { actions } = useCreateOrEditProjectSecretDialog({
        onClose: () => {
            actions.close();
        },
    });

    const { data: { data: secrets } = DEFAULT_PAGINATED_DATA, isFetching } =
        ProjectSecretsQueries.useFindManyPaginated({
            projectID: projectId,
            pagination,
            sorting,
            search,
        });

    const columns = useMemo(() => ProjectSecretsTableDefs.columns(projectId), [projectId]);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <Button
                        onClick={() => {
                            actions.open(projectId);
                        }}
                    >
                        <Plus /> Create Secret
                    </Button>
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
