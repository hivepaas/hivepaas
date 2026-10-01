import { useEffect, useMemo, useState } from "react";

import { CircleHelp, Plus } from "lucide-react";
import { useNavigate } from "react-router";
import { ProjectAppsQueries, ProjectsQueries } from "~/projects/data/queries";
import { useCreateFunctionDialog } from "~/projects/dialogs/create-function";
import { useCreateProjectAppDialog } from "~/projects/dialogs/create-project-app";
import {
    ALL_APP_CATEGORIES,
    APP_CATEGORY_FUNCTION,
    type ProjectAppDetails,
    type ProjectEnvEntity,
} from "~/projects/domain";
import { ProjectEnvScopeBadge } from "~/projects/module-shared/components";
import { ProjectAppsTableDefs } from "~/projects/module-shared/definitions/tables/project-apps";
import { EProjectStatus } from "~/projects/module-shared/enums";
import {
    PROJECT_ENV_FILTER_ALL,
    getProjectEnvFilterParam,
    useSelectedProjectEnv,
} from "~/projects/module-shared/hooks";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { Button, DataTable, Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

const EMPTY_PROJECT_ENVS: ProjectEnvEntity[] = [];

function getScopeTooltip(selectedEnv: string): string {
    if (!selectedEnv || selectedEnv === PROJECT_ENV_FILTER_ALL) {
        return "All apps of the project. Switch environments in the top right to change scope.";
    }

    return `Env Apps belong to env "${selectedEnv}" only. Switch environments in the top right to change scope.`;
}

const PROJECT_APPS_REFETCH_INTERVAL_MS = 5_000;

/**
 * Which apps the list shows: all of them, the functions, or the others.
 */
const APP_KIND_FILTERS = {
    all: { label: "All kinds", category: undefined },
    functions: { label: "Functions", category: [APP_CATEGORY_FUNCTION] },
    apps: { label: "Apps", category: ALL_APP_CATEGORIES },
} as const satisfies Record<string, { label: string; category: readonly string[] | undefined }>;

type AppKindFilter = keyof typeof APP_KIND_FILTERS;

export function ProjectAppsTable({ projectId }: Props) {
    const navigate = useNavigate();
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const env = getProjectEnvFilterParam(selectedEnv);
    const [kindFilter, setKindFilter] = useState<AppKindFilter>("all");
    const { actions } = useCreateProjectAppDialog({
        initialEnv: env,
        onClose: () => {
            actions.close();
        },
    });
    const { actions: createFunctionActions } = useCreateFunctionDialog({
        initialEnv: env,
        onClose: () => {
            createFunctionActions.close();
        },
    });
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    useEffect(() => {
        setPagination(prev => ({ ...prev, page: 1 }));
    }, [env, kindFilter, setPagination]);

    const {
        data: { data: rawApps, meta } = DEFAULT_PAGINATED_DATA,
        isLoading,
        isFetching,
        isPlaceholderData,
    } = ProjectAppsQueries.useFindManyPaginated(
        {
            projectID: projectId,
            pagination,
            sorting,
            search,
            env,
            getStats: true,
            getChildApps: true,
            ...(APP_KIND_FILTERS[kindFilter].category ? { category: [...APP_KIND_FILTERS[kindFilter].category] } : {}),
        },
        {
            refetchInterval: PROJECT_APPS_REFETCH_INTERVAL_MS,
        },
    );

    const apps = useMemo(() => {
        if (rawApps.length === 0) return [];

        const attachCombinedSubApps = (app: ProjectAppDetails): ProjectAppDetails => {
            const combined = [...(app.childApps ?? []), ...(app.logicalChildApps ?? [])];

            const uniqueSubApps = Array.from(
                new Map(combined.map(child => [child.id, attachCombinedSubApps(child)])).values(),
            );

            return {
                ...app,
                subApps: uniqueSubApps.length > 0 ? uniqueSubApps : undefined,
            };
        };

        return rawApps.map(attachCombinedSubApps);
    }, [rawApps]);
    const { data: projectData } = ProjectsQueries.useFindOneById({ projectID: projectId });

    const project = projectData?.data;
    const projectEnvs = project?.envs ?? EMPTY_PROJECT_ENVS;
    const columns = useMemo(
        () => ProjectAppsTableDefs.columns(projectId, projectEnvs, { isFetching }),
        [projectId, projectEnvs, isFetching],
    );
    const isProjectActive = project?.status === EProjectStatus.Active;
    const isAddButtonDisabled = !isProjectActive || !canWrite;

    const newFromTemplateButton = (
        <Button
            type="button"
            variant="outline"
            onClick={() => {
                void navigate(ROUTE.projects.single.appTemplates.$route(projectId));
            }}
        >
            <Plus /> New From Template
        </Button>
    );

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

    const addNewFunctionButton = (
        <Button
            variant="outline"
            disabled={isAddButtonDisabled}
            onClick={() => {
                if (!canWrite) {
                    return;
                }

                createFunctionActions.open(projectId);
            }}
        >
            <Plus /> New Function
        </Button>
    );

    const renderAddButton = (button: React.ReactElement) =>
        !canWrite ? (
            <PermissionTooltipAction
                id={MODULE_IDS.Project}
                action="write"
            >
                {() => button}
            </PermissionTooltipAction>
        ) : isAddButtonDisabled ? (
            <Tooltip>
                <TooltipTrigger asChild>
                    <span className="inline-flex">{button}</span>
                </TooltipTrigger>
                <TooltipContent side="top">
                    Project is not active. Activate the project to add a new app.
                </TooltipContent>
            </Tooltip>
        ) : (
            button
        );

    const renderActions = (
        <>
            {newFromTemplateButton}
            {renderAddButton(addNewFunctionButton)}
            {renderAddButton(addNewAppButton)}
        </>
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
                renderAfterSearch={
                    <Select
                        value={kindFilter}
                        onValueChange={value => {
                            setKindFilter(value as AppKindFilter);
                        }}
                    >
                        <SelectTrigger
                            className="w-[140px]"
                            aria-label="Kind of app"
                        >
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            {Object.entries(APP_KIND_FILTERS).map(([value, filter]) => (
                                <SelectItem
                                    key={value}
                                    value={value}
                                >
                                    {filter.label}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                }
                renderActions={renderActions}
            />
            <DataTable
                key={`${projectId}-${selectedEnv}`}
                columns={columns}
                data={apps}
                pageSize={pagination.size}
                enablePagination
                manualPagination
                totalCount={meta.page.total}
                manualSorting
                enableSorting
                isLoading={isLoading || isPlaceholderData}
                getSubRows={row => row.subApps}
                getRowId={(row, _, parent) => (parent ? `${parent.id}.${row.id}` : row.id)}
                initialExpanded
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
