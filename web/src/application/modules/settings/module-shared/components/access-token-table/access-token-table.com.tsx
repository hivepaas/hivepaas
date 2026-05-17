import { useMemo } from "react";

import { Plus } from "lucide-react";
import { ProjectAccessTokenQueries } from "~/projects/data/queries";
import { AccessTokenQueries } from "~/settings/data/queries";
import { useCreateOrEditAccessTokenDialog } from "~/settings/dialogs/create-or-edit-access-token";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { Button, DataTable } from "@/components/ui";

import { AccessTokenTableDefs } from "./access-token-table.defs";
import type { AccessTokenTableScope } from "./access-token-table.types";

function AccessTokenTableView({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const createOrEditDialog = useCreateOrEditAccessTokenDialog();

    const settingsQuery = AccessTokenQueries.useFindManyPaginated(
        { pagination, sorting, search },
        { enabled: scope.type === "settings" },
    );

    const projectQuery = ProjectAccessTokenQueries.useFindManyPaginated(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            pagination,
            sorting,
            search,
        },
        { enabled: scope.type === "project" },
    );

    const query = scope.type === "project" ? projectQuery : settingsQuery;
    const { data: { data: accessTokens, meta } = DEFAULT_PAGINATED_DATA, isFetching } = query;
    const columns = useMemo(() => AccessTokenTableDefs.columns(scope), [scope]);

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
                        New Access Token
                    </Button>
                }
            />
            <DataTable
                columns={columns}
                data={accessTokens}
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
    scope: AccessTokenTableScope;
}

export function SettingsAccessTokenTable() {
    return <AccessTokenTableView scope={{ type: "settings" }} />;
}

export function ProjectAccessTokenTable({ projectId }: ProjectProps) {
    return <AccessTokenTableView scope={{ type: "project", projectId }} />;
}

interface ProjectProps {
    projectId: string;
}
