import { useMemo } from "react";

import { Plus } from "lucide-react";
import { ProjectCommandTemplateQueries } from "~/projects/data/queries";
import { SettingsScopeCreateButton } from "~/settings/module-shared/components";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { ProjectCommandTemplateTableDefs } from "./project-command-template-table.defs";

const PROJECT_SCOPE = { type: "project" } as const;

export function ProjectCommandTemplateTable({ projectId }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

    const { data: { data: commandTemplateItems, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        ProjectCommandTemplateQueries.useFindManyPaginated({
            projectID: projectId,
            pagination,
            sorting,
            search,
        });
    const columns = useMemo(() => ProjectCommandTemplateTableDefs.columns(projectId), [projectId]);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <SettingsScopeCreateButton
                        scope={PROJECT_SCOPE}
                        onClick={() => {
                            navigate.modules(
                                ROUTE.projects.single.providerConfiguration.commandTemplates.create.$route(projectId),
                            );
                        }}
                    >
                        <Plus className="size-4" />
                        New Command Template
                    </SettingsScopeCreateButton>
                }
            />
            <DataTable
                columns={columns}
                data={commandTemplateItems}
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
}
