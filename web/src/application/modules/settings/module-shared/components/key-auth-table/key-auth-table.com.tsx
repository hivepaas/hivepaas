import { useMemo } from "react";

import { Plus } from "lucide-react";
import { PROJECT_SETTINGS_IMPORT_KIND } from "~/projects/data/commands";
import { ProjectKeyAuthQueries } from "~/projects/data/queries";
import { KeyAuthQueries } from "~/settings/data/queries";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { ProjectSettingsImportButton } from "../project-settings-import-button";
import { SettingsScopeCreateButton } from "../settings-scope-create-button";

import { KeyAuthTableDefs } from "./key-auth-table.defs";
import type { KeyAuthTableScope } from "./key-auth-table.types";

function KeyAuthTableView({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

    const settingsQuery = KeyAuthQueries.useFindManyPaginated(
        {
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "settings",
        },
    );

    const projectQuery = ProjectKeyAuthQueries.useFindManyPaginated(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            env: scope.type === "project" ? scope.env : undefined,
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "project",
        },
    );

    const query = scope.type === "project" ? projectQuery : settingsQuery;
    const { data: { data: keyAuthItems, meta } = DEFAULT_PAGINATED_DATA, isFetching } = query;
    const columns = useMemo(() => KeyAuthTableDefs.columns(scope), [scope]);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        {scope.type === "project" && (
                            <ProjectSettingsImportButton
                                projectId={scope.projectId}
                                env={scope.env}
                                settingKind={PROJECT_SETTINGS_IMPORT_KIND.KeyAuth}
                            />
                        )}
                        <SettingsScopeCreateButton
                            scope={scope}
                            onClick={() => {
                                navigate.modules(getKeyAuthCreateRoute(scope));
                            }}
                        >
                            <Plus className="size-4" />
                            New Key Auth
                        </SettingsScopeCreateButton>
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={keyAuthItems}
                pageSize={pagination.size}
                manualPagination
                totalCount={meta.page.total}
                manualSorting
                enableSorting
                enablePagination
                isLoading={isFetching}
                onPaginationChange={setPagination}
                onSortingChange={setSorting}
            />
        </div>
    );
}

function getKeyAuthCreateRoute(scope: KeyAuthTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.keyAuth.create.$route(scope.projectId);
    }

    return ROUTE.settings.keyAuth.create.$route;
}

interface Props {
    scope: KeyAuthTableScope;
}

export function SettingsKeyAuthTable() {
    return <KeyAuthTableView scope={{ type: "settings" }} />;
}

export function ProjectKeyAuthTable({ projectId, env }: ProjectProps) {
    return <KeyAuthTableView scope={{ type: "project", projectId, env }} />;
}

interface ProjectProps {
    projectId: string;
    env?: string;
}
