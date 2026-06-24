import { useMemo } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Plus } from "lucide-react";
import { PROJECT_SETTINGS_IMPORT_KIND } from "~/projects/data/commands";
import { ProjectGithubAppQueries } from "~/projects/data/queries";
import { GithubAppQueries } from "~/settings/data/queries";
import { useProvisionGithubAppDialog } from "~/settings/dialogs/provision-github-app";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { ProjectSettingsImportButton } from "../project-settings-import-button";
import { SettingsScopeCreateButton } from "../settings-scope-create-button";

import { GithubAppTableDefs } from "./github-app-table.defs";
import type { GithubAppTableScope } from "./github-app-table.types";

function GithubAppTableView({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const provisionDialog = useProvisionGithubAppDialog();
    const { navigate } = useAppNavigate();

    const settingsQuery = GithubAppQueries.useFindManyPaginated(
        {
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "settings",
        },
    );

    const projectQuery = ProjectGithubAppQueries.useFindManyPaginated(
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
    const { data: { data: githubApps, meta } = DEFAULT_PAGINATED_DATA, isFetching } = query;
    const columns = useMemo(() => GithubAppTableDefs.columns(scope), [scope]);

    return (
        <div className="flex flex-col gap-4">
            <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
                <span className="text-orange-500">Note:</span> It is recommended to use the GitHub App if you are
                hosting your source code on GitHub, as it is the most efficient and straightforward way for LocalPaaS to
                deploy your applications.
            </div>
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        {scope.type === "project" && (
                            <ProjectSettingsImportButton
                                projectId={scope.projectId}
                                settingKind={PROJECT_SETTINGS_IMPORT_KIND.GithubApp}
                            />
                        )}
                        <SettingsScopeCreateButton
                            scope={scope}
                            onClick={() => {
                                navigate.modules(getGithubAppCreateRoute(scope));
                            }}
                        >
                            <Plus className="size-4" />
                            Create Manually
                        </SettingsScopeCreateButton>
                        <SettingsScopeCreateButton
                            scope={scope}
                            onClick={() => {
                                provisionDialog.actions.open(scope);
                            }}
                        >
                            <Plus className="size-4" />
                            Provision Github App
                        </SettingsScopeCreateButton>
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={githubApps}
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
    scope: GithubAppTableScope;
}

function getGithubAppCreateRoute(scope: GithubAppTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.githubApps.create.$route(scope.projectId);
    }

    return ROUTE.sources.githubApps.create.$route;
}

export function SettingsGithubAppTable() {
    return <GithubAppTableView scope={{ type: "settings" }} />;
}

export function ProjectGithubAppTable({ projectId }: ProjectProps) {
    return <GithubAppTableView scope={{ type: "project", projectId }} />;
}

interface ProjectProps {
    projectId: string;
}
