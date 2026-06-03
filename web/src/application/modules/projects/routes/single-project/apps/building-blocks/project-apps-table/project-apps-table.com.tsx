import { useEffect, useMemo } from "react";

import { Plus } from "lucide-react";
import { ProjectAppsQueries, ProjectsQueries } from "~/projects/data/queries";
import { useCreateProjectAppDialog } from "~/projects/dialogs/create-project-app";
import type { ProjectEnvEntity } from "~/projects/domain";
import { ProjectAppsTableDefs } from "~/projects/module-shared/definitions/tables/project-apps";
import { EProjectStatus } from "~/projects/module-shared/enums";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { Button, DataTable, Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui";

const EMPTY_PROJECT_ENVS: readonly ProjectEnvEntity[] = [];

export function ProjectAppsTable({ projectId }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const env = getProjectEnvFilterParam(selectedEnv);
    const { actions } = useCreateProjectAppDialog({
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
