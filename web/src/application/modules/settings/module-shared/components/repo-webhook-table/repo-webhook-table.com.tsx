import { useMemo } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Plus } from "lucide-react";
import { PROJECT_SETTINGS_IMPORT_KIND } from "~/projects/data/commands";
import { ProjectRepoWebhookQueries } from "~/projects/data/queries";
import { RepoWebhookQueries } from "~/settings/data/queries";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { ProjectSettingsImportButton } from "../project-settings-import-button";
import { SettingsScopeCreateButton } from "../settings-scope-create-button";

import { RepoWebhookTableDefs } from "./repo-webhook-table.defs";
import type { RepoWebhookTableScope } from "./repo-webhook-table.types";

function RepoWebhookTableView({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

    const settingsQuery = RepoWebhookQueries.useFindManyPaginated(
        {
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "settings",
        },
    );

    const projectQuery = ProjectRepoWebhookQueries.useFindManyPaginated(
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
    const { data: { data: repoWebhooks, meta } = DEFAULT_PAGINATED_DATA, isFetching } = query;
    const columns = useMemo(() => RepoWebhookTableDefs.columns(scope), [scope]);

    return (
        <div className="flex flex-col gap-4">
            <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
                <span className="text-orange-500">Note:</span> If you are using a GitHub App, webhook events have been
                automatically configured, and you don&apos;t need to set them up here.
            </div>
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        {scope.type === "project" && (
                            <ProjectSettingsImportButton
                                projectId={scope.projectId}
                                settingKind={PROJECT_SETTINGS_IMPORT_KIND.RepoWebhook}
                            />
                        )}
                        <SettingsScopeCreateButton
                            scope={scope}
                            onClick={() => {
                                navigate.modules(getRepoWebhookCreateRoute(scope));
                            }}
                        >
                            <Plus className="size-4" />
                            New Webhook
                        </SettingsScopeCreateButton>
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={repoWebhooks}
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
    scope: RepoWebhookTableScope;
}

function getRepoWebhookCreateRoute(scope: RepoWebhookTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.webhooks.create.$route(scope.projectId);
    }

    return ROUTE.sources.webhooks.create.$route;
}

export function SettingsRepoWebhookTable() {
    return <RepoWebhookTableView scope={{ type: "settings" }} />;
}

export function ProjectRepoWebhookTable({ projectId }: ProjectProps) {
    return <RepoWebhookTableView scope={{ type: "project", projectId }} />;
}

interface ProjectProps {
    projectId: string;
}
