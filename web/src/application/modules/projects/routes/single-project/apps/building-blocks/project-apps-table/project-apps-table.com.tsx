import { useMemo } from "react";

import { Plus } from "lucide-react";
import { ProjectAppsQueries, ProjectsQueries } from "~/projects/data/queries";
import { useCreateProjectAppDialog } from "~/projects/dialogs/create-project-app";
import { ProjectAppsTableDefs } from "~/projects/module-shared/definitions/tables/project-apps";
import { EProjectStatus } from "~/projects/module-shared/enums";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { Button, DataTable, Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui";

export function ProjectAppsTable({ projectId }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { actions } = useCreateProjectAppDialog({
        onClose: () => {
            actions.close();
        },
    });

    const { data: { data: apps } = DEFAULT_PAGINATED_DATA, isFetching } = ProjectAppsQueries.useFindManyPaginated({
        projectID: projectId,
        pagination,
        sorting,
        search,
    });
    const { data: projectData } = ProjectsQueries.useFindOneById({ projectID: projectId });

    const columns = useMemo(() => ProjectAppsTableDefs.columns(projectId), [projectId]);
    const project = projectData?.data;
    const isProjectActive = project?.status === EProjectStatus.Active;
    const isAddButtonDisabled = !isProjectActive;

    const addNewAppButton = (
        <Button
            disabled={isAddButtonDisabled}
            onClick={() => {
                actions.open(projectId);
            }}
        >
            <Plus /> New App
        </Button>
    );

    const renderActions = isAddButtonDisabled ? (
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
