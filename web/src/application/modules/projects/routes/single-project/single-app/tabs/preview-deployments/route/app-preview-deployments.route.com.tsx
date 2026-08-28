import { useMemo } from "react";

import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";
import { Plus } from "lucide-react";
import { Navigate, useParams } from "react-router";
import invariant from "tiny-invariant";
import { AppPreviewsCommands, AppPreviewsQueries, ProjectAppsQueries, ProjectsQueries } from "~/projects/data";
import type { ProjectEnvEntity } from "~/projects/domain";

import { AppLink, AppLoader, TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { isFeatureDisabledException } from "@infrastructure/api";

import { Button, DataTable } from "@/components/ui";

import { AppPreviewDeploymentsTableDefs } from "../building-blocks";

const EMPTY_PROJECT_ENVS: readonly ProjectEnvEntity[] = [];

export function AppPreviewDeploymentsRoute() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    const { navigate } = useAppNavigate();
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    const { data: appData, isLoading: isLoadingApp } = ProjectAppsQueries.useFindOneById({
        projectID: projectId,
        env,
        appID: appId,
    });
    const isChildApp = Boolean(appData?.data.parentApp);
    const canLoadPreviews = Boolean(appData) && !isChildApp;
    const {
        data: { data: previews, meta } = DEFAULT_PAGINATED_DATA,
        isFetching,
        error,
    } = AppPreviewsQueries.useFindManyPaginated(
        {
            projectID: projectId,
            env,
            appID: appId,
            pagination,
            sorting,
            search,
            getStats: true,
        },
        {
            enabled: canLoadPreviews,
        },
    );
    const isFeatureDisabled = error instanceof Error && isFeatureDisabledException(error);
    const { data: projectData } = ProjectsQueries.useFindOneById({ projectID: projectId });
    const projectEnvs = projectData?.data.envs ?? EMPTY_PROJECT_ENVS;
    const columns = useMemo(
        () => AppPreviewDeploymentsTableDefs.columns(projectId, projectEnvs),
        [projectId, projectEnvs],
    );

    const { mutate: preparePreview, isPending: isPreparing } = AppPreviewsCommands.usePrepareCreate({
        onSuccess: response => {
            navigate.modules(
                ROUTE.projects.single.apps.single.previewDeployments.create.$route(projectId, env, appId),
                {
                    state: { preparedPreview: response.data },
                },
            );
        },
    });

    if (isLoadingApp) {
        return <AppLoader />;
    }

    if (isChildApp) {
        return (
            <Navigate
                to={ROUTE.projects.single.apps.single.configuration.general.$route(projectId, env, appId)}
                replace
            />
        );
    }

    if (isFeatureDisabled) {
        return (
            <section className={cn(listBox)}>
                <p className="text-base">
                    App preview feature is disabled, enable it in{" "}
                    <AppLink.Basic
                        to={ROUTE.projects.single.apps.single.configuration.featureSettings.$route(
                            projectId,
                            env,
                            appId,
                        )}
                        className="text-primary underline-offset-4 hover:underline"
                    >
                        Feature Settings
                    </AppLink.Basic>
                </p>
            </section>
        );
    }

    return (
        <section className={cn(listBox)}>
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
                                    disabled={isDenied}
                                    isLoading={isPreparing}
                                    onClick={() => {
                                        if (isDenied) {
                                            return;
                                        }

                                        preparePreview({
                                            projectID: projectId,
                                            env,
                                            appID: appId,
                                        });
                                    }}
                                >
                                    <Plus className="size-4" /> New Preview Deployment
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    }
                />

                <DataTable
                    columns={columns}
                    data={previews}
                    isLoading={isFetching}
                    pageSize={pagination.size}
                    manualPagination
                    totalCount={meta.page.total}
                    onPaginationChange={setPagination}
                    manualSorting
                    onSortingChange={setSorting}
                    enablePagination
                />
            </div>
        </section>
    );
}
