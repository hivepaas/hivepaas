import { useMemo } from "react";

import { Plus } from "lucide-react";
import { PROJECT_SETTINGS_IMPORT_KIND } from "~/projects/data/commands";
import { ProjectBackupRepoQueries } from "~/projects/data/queries";
import { BackupRepoQueries } from "~/settings/data/queries";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { ProjectSettingsImportButton } from "../project-settings-import-button";
import { SettingsScopeCreateButton } from "../settings-scope-create-button";

import { BackupRepoTableDefs } from "./backup-repo-table.defs";
import type { BackupRepoTableScope } from "./backup-repo-table.types";

function BackupRepoTableView({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();

    const settingsQuery = BackupRepoQueries.useFindManyPaginated(
        {
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "settings",
        },
    );

    const projectQuery = ProjectBackupRepoQueries.useFindManyPaginated(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            env: scope.type === "project" ? scope.env : undefined,
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "project",
        },
    );

    const query = scope.type === "project" ? projectQuery : settingsQuery;
    const { data: { data: backupRepoItems, meta } = DEFAULT_PAGINATED_DATA, isFetching } = query;
    const columns = useMemo(() => BackupRepoTableDefs.columns(scope), [scope]);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        {scope.type === "project" && (
                            <ProjectSettingsImportButton
                                projectId={scope.projectId}
                                env={scope.env}
                                settingKind={PROJECT_SETTINGS_IMPORT_KIND.BackupRepo}
                            />
                        )}
                        <SettingsScopeCreateButton
                            scope={scope}
                            onClick={() => {
                                // Handler will be implemented in a dedicated task
                            }}
                        >
                            <Plus className="size-4" />
                            New Backup Repo
                        </SettingsScopeCreateButton>
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={backupRepoItems}
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
    scope: BackupRepoTableScope;
}

export function SettingsBackupRepoTable() {
    return <BackupRepoTableView scope={{ type: "settings" }} />;
}

export function ProjectBackupRepoTable({ projectId, env }: ProjectProps) {
    return <BackupRepoTableView scope={{ type: "project", projectId, env }} />;
}

interface ProjectProps {
    projectId: string;
    env?: string;
}
