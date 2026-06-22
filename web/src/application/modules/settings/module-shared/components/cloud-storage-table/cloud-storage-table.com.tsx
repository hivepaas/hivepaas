import { useMemo } from "react";

import { Plus } from "lucide-react";
import { PROJECT_SETTINGS_IMPORT_KIND } from "~/projects/data/commands";
import { ProjectCloudStorageQueries } from "~/projects/data/queries";
import { CloudStorageQueries } from "~/settings/data/queries";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { ProjectSettingsImportButton } from "../project-settings-import-button";
import { SettingsScopeCreateButton } from "../settings-scope-create-button";

import { CloudStorageTableDefs } from "./cloud-storage-table.defs";
import type { CloudStorageTableScope } from "./cloud-storage-table.types";

function CloudStorageTableView({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

    const settingsQuery = CloudStorageQueries.useFindManyPaginated(
        { pagination, sorting, search },
        { enabled: scope.type === "settings" },
    );

    const projectQuery = ProjectCloudStorageQueries.useFindManyPaginated(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            pagination,
            sorting,
            search,
        },
        { enabled: scope.type === "project" },
    );

    const query = scope.type === "project" ? projectQuery : settingsQuery;
    const { data: { data: cloudStorages, meta } = DEFAULT_PAGINATED_DATA, isFetching } = query;
    const columns = useMemo(() => CloudStorageTableDefs.columns(scope), [scope]);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        {scope.type === "project" && (
                            <ProjectSettingsImportButton
                                projectId={scope.projectId}
                                settingKind={PROJECT_SETTINGS_IMPORT_KIND.CloudStorage}
                            />
                        )}
                        <SettingsScopeCreateButton
                            scope={scope}
                            onClick={() => {
                                navigate.modules(getCloudStorageCreateRoute(scope));
                            }}
                        >
                            <Plus className="size-4" />
                            New Cloud Storage
                        </SettingsScopeCreateButton>
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={cloudStorages}
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

function getCloudStorageCreateRoute(scope: CloudStorageTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.cloudStorages.create.$route(scope.projectId);
    }

    return ROUTE.settings.cloudStorages.create.$route;
}

interface Props {
    scope: CloudStorageTableScope;
}

export function SettingsCloudStorageTable() {
    return <CloudStorageTableView scope={{ type: "settings" }} />;
}

export function ProjectCloudStorageTable({ projectId }: ProjectProps) {
    return <CloudStorageTableView scope={{ type: "project", projectId }} />;
}

interface ProjectProps {
    projectId: string;
}
