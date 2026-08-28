import { useMemo } from "react";

import { Tooltip, TooltipContent, TooltipTrigger } from "@components/ui/tooltip";
import { CircleHelp, Plus } from "lucide-react";
import { ProjectsQueries } from "~/projects/data";
import { ProjectSecretsQueries } from "~/projects/data/queries";
import { ProjectEnvScopeBadge } from "~/projects/module-shared/components";
import { ProjectSecretsTableDefs } from "~/projects/module-shared/definitions/tables/project-secrets";
import {
    PROJECT_ENV_FILTER_ALL,
    getProjectEnvFilterParam,
    useSelectedProjectEnv,
} from "~/projects/module-shared/hooks";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { Button, DataTable } from "@/components/ui";

function getScopeTooltip(selectedEnv: string): string {
    if (!selectedEnv || selectedEnv === PROJECT_ENV_FILTER_ALL) {
        return "Secrets take effect in all environments. Switch environments in the top right to change scope.";
    }

    return `Secrets only take effect in env "${selectedEnv}". Switch environments in the top right to change scope.`;
}

export function ProjectSecretsTable({ projectId }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);
    const { data: projectData } = ProjectsQueries.useFindOneById({ projectID: projectId });
    const projectEnvs = projectData?.data.envs ?? [];

    const { data: { data: secrets, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        ProjectSecretsQueries.useFindManyPaginated({
            projectID: projectId,
            env: scopedEnv,
            pagination,
            sorting,
            search,
        });

    const columns = useMemo(() => ProjectSecretsTableDefs.columns(projectId, scopedEnv), [projectId, scopedEnv]);

    return (
        <div className="flex flex-col gap-4">
            <div className="flex items-center gap-2">
                <ProjectEnvScopeBadge
                    selectedEnv={selectedEnv}
                    envs={projectEnvs}
                />
                <Tooltip>
                    <TooltipTrigger asChild>
                        <button
                            type="button"
                            className="inline-flex size-5 shrink-0 items-center justify-center rounded-full text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                            aria-label="Secret scope help"
                        >
                            <CircleHelp className="size-4" />
                        </button>
                    </TooltipTrigger>
                    <TooltipContent
                        side="right"
                        className="max-w-xs"
                    >
                        {getScopeTooltip(selectedEnv)}
                    </TooltipContent>
                </Tooltip>
            </div>
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <PermissionTooltipAction
                        id={MODULE_IDS.Project}
                        action="write"
                    >
                        {({ isDenied }) => (
                            <Button
                                onClick={() => {
                                    navigate.modules(
                                        ROUTE.projects.single.providerConfiguration.secrets.create.$route(projectId),
                                    );
                                }}
                                type="button"
                                color="primary"
                                disabled={isDenied}
                            >
                                <Plus /> New Secret
                            </Button>
                        )}
                    </PermissionTooltipAction>
                }
            />
            <DataTable
                columns={columns}
                data={secrets}
                pageSize={pagination.size}
                enablePagination
                manualPagination
                totalCount={meta.page.total}
                manualSorting
                enableSorting
                isLoading={isFetching}
                onPaginationChange={value => {
                    setPagination(value);
                }}
                onSortingChange={value => {
                    setSorting(value);
                }}
            />
        </div>
    );
}

interface Props {
    projectId: string;
}
