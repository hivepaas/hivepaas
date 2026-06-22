import { useMemo } from "react";

import { Plus } from "lucide-react";
import { PROJECT_SETTINGS_IMPORT_KIND } from "~/projects/data/commands";
import { ProjectAccessTokenQueries } from "~/projects/data/queries";
import { AccessTokenQueries } from "~/settings/data/queries";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { ProjectSettingsImportButton } from "../project-settings-import-button";
import { SettingsScopeCreateButton } from "../settings-scope-create-button";

import { AccessTokenTableDefs } from "./access-token-table.defs";
import type { AccessTokenTableScope } from "./access-token-table.types";

function AccessTokenTableView({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

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
                    <div className="flex flex-wrap gap-3">
                        {scope.type === "project" && (
                            <ProjectSettingsImportButton
                                projectId={scope.projectId}
                                settingKind={PROJECT_SETTINGS_IMPORT_KIND.AccessToken}
                            />
                        )}
                        <SettingsScopeCreateButton
                            scope={scope}
                            onClick={() => {
                                navigate.modules(getAccessTokenCreateRoute(scope));
                            }}
                        >
                            <Plus className="size-4" />
                            New Access Token
                        </SettingsScopeCreateButton>
                    </div>
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

function getAccessTokenCreateRoute(scope: AccessTokenTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.accessTokens.create.$route(scope.projectId);
    }

    return ROUTE.settings.accessTokens.create.$route;
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
