import { useMemo } from "react";

import { Plus } from "lucide-react";
import { ProjectRegistryAuthQueries } from "~/projects/data/queries";
import { RegistryAuthQueries } from "~/settings/data/queries";
import { useCreateOrEditRegistryAuthDialog } from "~/settings/dialogs/create-or-edit-registry-auth";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { Button, DataTable } from "@/components/ui";

import { RegistryAuthTableDefs } from "./registry-auth-table.defs";
import type { RegistryAuthTableScope } from "./registry-auth-table.types";

function RegistryAuthTableView({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const createOrEditDialog = useCreateOrEditRegistryAuthDialog();

    const settingsQuery = RegistryAuthQueries.useFindManyPaginated(
        {
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "settings",
        },
    );

    const projectQuery = ProjectRegistryAuthQueries.useFindManyPaginated(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "project",
        },
    );

    const query = scope.type === "project" ? projectQuery : settingsQuery;
    const { data: { data: registryAuthItems, meta } = DEFAULT_PAGINATED_DATA, isFetching } = query;
    const columns = useMemo(() => RegistryAuthTableDefs.columns(scope), [scope]);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <Button
                        onClick={() => {
                            createOrEditDialog.actions.open(scope);
                        }}
                    >
                        <Plus className="size-4" />
                        New Registry Auth
                    </Button>
                }
            />
            <DataTable
                columns={columns}
                data={registryAuthItems}
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

interface Props {
    scope: RegistryAuthTableScope;
}

export function SettingsRegistryAuthTable() {
    return <RegistryAuthTableView scope={{ type: "settings" }} />;
}

export function ProjectRegistryAuthTable({ projectId }: ProjectProps) {
    return <RegistryAuthTableView scope={{ type: "project", projectId }} />;
}

interface ProjectProps {
    projectId: string;
}
