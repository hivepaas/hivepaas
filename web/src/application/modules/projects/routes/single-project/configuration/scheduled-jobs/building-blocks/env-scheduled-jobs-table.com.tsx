import { useMemo, useState } from "react";

import { Plus } from "lucide-react";
import { EnvScheduledJobsQueries, ProjectAppsQueries } from "~/projects/data/queries";
import { EAppScheduledJobType } from "~/projects/module-shared/enums";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { Button, DataTable, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui";

import { EnvScheduledJobsTableDefs } from "./env-scheduled-jobs-table.defs";

const ALL = "all";
const LIST_ALL_PAGE = { page: 1, size: 1000 };

const JOB_TYPE_OPTIONS = [
    { value: ALL, label: "All types" },
    { value: EAppScheduledJobType.JobSequence, label: "Job sequences" },
    { value: EAppScheduledJobType.ContainerCommand, label: "Container commands" },
];

/** An env's scheduled jobs: its own, and those of the apps in it. */
export function EnvScheduledJobsTable({ projectId, env }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();
    const [jobType, setJobType] = useState<string>(ALL);
    const [appId, setAppId] = useState<string>(ALL);

    const { data: appsData } = ProjectAppsQueries.useFindManyPaginated({
        projectID: projectId,
        env,
        pagination: LIST_ALL_PAGE,
    });
    const apps = appsData?.data ?? [];

    const { data: { data: jobs, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        EnvScheduledJobsQueries.useFindManyPaginated({
            projectID: projectId,
            env,
            pagination,
            sorting,
            search,
            ...(jobType === ALL ? {} : { jobTypes: [jobType as EAppScheduledJobType] }),
            ...(appId === ALL ? {} : { appId }),
        });
    const columns = useMemo(() => EnvScheduledJobsTableDefs.columns(projectId, env), [projectId, env]);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        <Select
                            value={jobType}
                            onValueChange={setJobType}
                        >
                            <SelectTrigger className="w-[190px]">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                {JOB_TYPE_OPTIONS.map(option => (
                                    <SelectItem
                                        key={option.value}
                                        value={option.value}
                                    >
                                        {option.label}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                        <Select
                            value={appId}
                            onValueChange={setAppId}
                        >
                            <SelectTrigger className="w-[190px]">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value={ALL}>All owners</SelectItem>
                                {apps.map(app => (
                                    <SelectItem
                                        key={app.id}
                                        value={app.id}
                                    >
                                        {app.name}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                        <PermissionTooltipAction
                            id={MODULE_IDS.Project}
                            action="write"
                        >
                            {({ isDenied }) => (
                                <Button
                                    disabled={isDenied}
                                    onClick={() => {
                                        navigate.modules(
                                            ROUTE.projects.single.providerConfiguration.scheduledJobs.createSequence.$route(
                                                projectId,
                                            ),
                                        );
                                    }}
                                >
                                    <Plus className="size-4" /> New Job Sequence
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={jobs}
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

interface Props {
    projectId: string;
    env: string;
}
