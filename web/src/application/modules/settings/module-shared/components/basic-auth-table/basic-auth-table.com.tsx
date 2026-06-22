import { useMemo } from "react";

import { Plus } from "lucide-react";
import { PROJECT_SETTINGS_IMPORT_KIND } from "~/projects/data/commands";
import { ProjectBasicAuthQueries } from "~/projects/data/queries";
import { BasicAuthQueries } from "~/settings/data/queries";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { ProjectSettingsImportButton } from "../project-settings-import-button";
import { SettingsScopeCreateButton } from "../settings-scope-create-button";

import { BasicAuthTableDefs } from "./basic-auth-table.defs";
import type { BasicAuthTableScope } from "./basic-auth-table.types";

function BasicAuthTableView({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

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
                        <SettingsScopeCreateButton
                            scope={scope}
                            onClick={() => {
                                navigate.modules(getBasicAuthCreateRoute(scope));
                            }}
                        >
                            <Plus className="size-4" />
                            New Basic Auth
                        </SettingsScopeCreateButton>
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

function getBasicAuthCreateRoute(scope: BasicAuthTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.basicAuth.create.$route(scope.projectId);
    }

    return ROUTE.settings.basicAuth.create.$route;
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
