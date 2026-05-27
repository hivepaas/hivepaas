import { useMemo } from "react";

import { Plus } from "lucide-react";
import { PROJECT_SETTINGS_IMPORT_KIND } from "~/projects/data/commands";
import { ProjectBasicAuthQueries } from "~/projects/data/queries";
import { BasicAuthQueries } from "~/settings/data/queries";
import { useCreateOrEditBasicAuthDialog } from "~/settings/dialogs/create-or-edit-basic-auth";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { Button, DataTable } from "@/components/ui";

import { ProjectSettingsImportButton } from "../project-settings-import-button";

import { BasicAuthTableDefs } from "./basic-auth-table.defs";
import type { BasicAuthTableScope } from "./basic-auth-table.types";

function BasicAuthTableView({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const createOrEditDialog = useCreateOrEditBasicAuthDialog();

    const settingsQuery = BasicAuthQueries.useFindManyPaginated(
        {
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "settings",
        },
    );

    const projectQuery = ProjectBasicAuthQueries.useFindManyPaginated(
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
    const { data: { data: basicAuthItems, meta } = DEFAULT_PAGINATED_DATA, isFetching } = query;
    const columns = useMemo(() => BasicAuthTableDefs.columns(scope), [scope]);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        {scope.type === "project" && (
                            <ProjectSettingsImportButton
                                projectId={scope.projectId}
                                settingKind={PROJECT_SETTINGS_IMPORT_KIND.BasicAuth}
                            />
                        )}
                        <Button
                            onClick={() => {
                                createOrEditDialog.actions.open(scope);
                            }}
                        >
                            <Plus className="size-4" />
                            New Basic Auth
                        </Button>
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={basicAuthItems}
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
    scope: BasicAuthTableScope;
}

export function SettingsBasicAuthTable() {
    return <BasicAuthTableView scope={{ type: "settings" }} />;
}

export function ProjectBasicAuthTable({ projectId }: ProjectProps) {
    return <BasicAuthTableView scope={{ type: "project", projectId }} />;
}

interface ProjectProps {
    projectId: string;
}
