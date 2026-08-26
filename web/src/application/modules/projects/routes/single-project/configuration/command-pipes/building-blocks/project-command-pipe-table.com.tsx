import { useMemo } from "react";

import { Plus } from "lucide-react";
import { ProjectCommandPipeQueries } from "~/projects/data/queries";
import { SettingsScopeCreateButton } from "~/settings/module-shared/components";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { ProjectCommandPipeFromTemplateMenu } from "./project-command-pipe-from-template-menu.com";
import { ProjectCommandPipeTableDefs } from "./project-command-pipe-table.defs";

const PROJECT_SCOPE = { type: "project" } as const;

export function ProjectCommandPipeTable({ projectId, env }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

    const { data: { data: commandPipeItems, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        ProjectCommandPipeQueries.useFindManyPaginated({
            projectID: projectId,
            env,
            pagination,
            sorting,
            search,
        });
    const columns = useMemo(() => ProjectCommandPipeTableDefs.columns(projectId, env), [projectId, env]);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        <ProjectCommandPipeFromTemplateMenu
                            projectId={projectId}
                            env={env}
                        />
                        <SettingsScopeCreateButton
                            scope={PROJECT_SCOPE}
                            onClick={() => {
                                navigate.modules(
                                    ROUTE.projects.single.providerConfiguration.commandPipes.create.$route(projectId),
                                );
                            }}
                        >
                            <Plus className="size-4" />
                            New Command Pipe
                        </SettingsScopeCreateButton>
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={commandPipeItems}
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
    projectId: string;
    env?: string;
}
