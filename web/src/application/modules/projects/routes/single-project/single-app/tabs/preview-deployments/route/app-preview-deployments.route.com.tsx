import { useMemo } from "react";

import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";
import { Plus } from "lucide-react";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { AppPreviewsCommands, AppPreviewsQueries, ProjectsQueries } from "~/projects/data";
import type { ProjectEnvEntity } from "~/projects/domain";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { Button, DataTable } from "@/components/ui";

import { AppPreviewDeploymentsTableDefs } from "../building-blocks";

const EMPTY_PROJECT_ENVS: readonly ProjectEnvEntity[] = [];

export function AppPreviewDeploymentsRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const { navigate } = useAppNavigate();
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data: { data: previews, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        AppPreviewsQueries.useFindManyPaginated({
            projectID: projectId,
            appID: appId,
            pagination,
            sorting,
            search,
        });
    const { data: projectData } = ProjectsQueries.useFindOneById({ projectID: projectId });
    const projectEnvs = projectData?.data.envs ?? EMPTY_PROJECT_ENVS;
    const columns = useMemo(
        () => AppPreviewDeploymentsTableDefs.columns(projectId, projectEnvs),
        [projectId, projectEnvs],
    );

    const { mutate: preparePreview, isPending: isPreparing } = AppPreviewsCommands.usePrepareCreate({
        onSuccess: response => {
            navigate.modules(ROUTE.projects.single.apps.single.previewDeployments.create.$route(projectId, appId), {
                state: { preparedPreview: response.data },
            });
        },
    });

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
                    showPageSizeSelector={false}
                />
            </div>
        </section>
    );
}
