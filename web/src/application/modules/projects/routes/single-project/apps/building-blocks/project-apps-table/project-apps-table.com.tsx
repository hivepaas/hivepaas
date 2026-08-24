import { useEffect, useMemo } from "react";

import { CircleHelp, Plus } from "lucide-react";
import { ProjectAppsQueries, ProjectsQueries } from "~/projects/data/queries";
import { useCreateProjectAppDialog } from "~/projects/dialogs/create-project-app";
import type { ProjectEnvEntity } from "~/projects/domain";
import { ProjectEnvScopeBadge } from "~/projects/module-shared/components";
import { ProjectAppsTableDefs } from "~/projects/module-shared/definitions/tables/project-apps";
import { EProjectStatus } from "~/projects/module-shared/enums";
import {
    PROJECT_ENV_FILTER_ALL,
    getProjectEnvFilterParam,
    useSelectedProjectEnv,
} from "~/projects/module-shared/hooks";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { Button, DataTable, Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui";

const EMPTY_PROJECT_ENVS: ProjectEnvEntity[] = [];

function getScopeTooltip(selectedEnv: string): string {
    if (!selectedEnv || selectedEnv === PROJECT_ENV_FILTER_ALL) {
        return "All apps of the project. Switch environments in the top right to change scope.";
    }

    return `Env Apps belong to env "${selectedEnv}" only. Switch environments in the top right to change scope.`;
}

export function ProjectAppsTable({ projectId }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const env = getProjectEnvFilterParam(selectedEnv);
    const { actions } = useCreateProjectAppDialog({
        initialEnv: env,
        onClose: () => {
            actions.close();
        },
    });
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    useEffect(() => {
        setPagination(prev => ({ ...prev, page: 1 }));
    }, [env, setPagination]);

    const { data: { data: apps } = DEFAULT_PAGINATED_DATA, isFetching } = ProjectAppsQueries.useFindManyPaginated({
        projectID: projectId,
        pagination,
        sorting,
        search,
        env,
        getStats: true,
    });
    const { data: projectData } = ProjectsQueries.useFindOneById({ projectID: projectId });

    const project = projectData?.data;
    const projectEnvs = project?.envs ?? EMPTY_PROJECT_ENVS;
    const columns = useMemo(() => ProjectAppsTableDefs.columns(projectId, projectEnvs), [projectId, projectEnvs]);
    const isProjectActive = project?.status === EProjectStatus.Active;
    const isAddButtonDisabled = !isProjectActive || !canWrite;

    const addNewAppButton = (
        <Button
            disabled={isAddButtonDisabled}
            onClick={() => {
                if (!canWrite) {
                    return;
                }

                actions.open(projectId);
            }}
        >
            <Plus /> New App
        </Button>
    );

    const renderActions = !canWrite ? (
        <PermissionTooltipAction
            id={MODULE_IDS.Project}
            action="write"
        >
            {() => addNewAppButton}
        </PermissionTooltipAction>
    ) : isAddButtonDisabled ? (
        <Tooltip>
            <TooltipTrigger asChild>
                <span className="inline-flex">{addNewAppButton}</span>
            </TooltipTrigger>
            <TooltipContent side="top">Project is not active. Activate the project to add a new app.</TooltipContent>
        </Tooltip>
    ) : (
        addNewAppButton
    );

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
                            aria-label="App scope help"
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
                renderActions={renderActions}
            />
            <DataTable
                columns={columns}
                data={apps}
                pageSize={pagination.size}
                enablePagination
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
