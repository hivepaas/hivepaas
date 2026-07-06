import { useMemo } from "react";

import { Button } from "@components/ui";
import { Tooltip, TooltipContent, TooltipTrigger } from "@components/ui/tooltip";
import { Plus, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { ClusterVolumesCommands } from "~/cluster/data/commands";
import { ClusterVolumesQueries } from "~/cluster/data/queries";
import { VolumesTableDefs } from "~/cluster/module-shared/definitions/tables/volumes/volumes-table.defs";
import type { VolumeManagementScope } from "~/cluster/module-shared/types";
import { PROJECT_SETTINGS_IMPORT_KIND } from "~/projects/data/commands";
import { ProjectClusterVolumesQueries } from "~/projects/data/queries";
import { ProjectSettingsImportButton } from "~/settings/module-shared/components";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";
import { useConditionalModule, useConditionalProject } from "@application/shared/permissions";

import { DataTable } from "@/components/ui";

export function VolumeManagementTable({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const clusterPermission = useConditionalModule({ id: MODULE_IDS.Cluster });
    const projectPermission = useConditionalProject({
        projectId: scope.type === "project" ? scope.projectId : "",
    });
    const { navigate } = useAppNavigate();

    const clusterVolumesQuery = ClusterVolumesQueries.useFindManyPaginated(
        {
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "cluster",
        },
    );
    const projectVolumesQuery = ProjectClusterVolumesQueries.useFindManyPaginated(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "project",
        },
    );

    const { mutate: syncFromDocker, isPending: isSyncing } = ClusterVolumesCommands.useSyncFromDocker({
        onSuccess: () => {
            toast.success("Volumes synced from Docker");
        },
    });

    const query = scope.type === "cluster" ? clusterVolumesQuery : projectVolumesQuery;
    const { data: { data: volumes, meta } = DEFAULT_PAGINATED_DATA, isFetching } = query;
    const columns = useMemo(() => VolumesTableDefs.columns(scope), [scope]);
    const canCreate = scope.type === "cluster" ? clusterPermission.canWrite : projectPermission.canWrite;
    const deniedMessage =
        scope.type === "cluster"
            ? "You only have view access. Create actions are disabled."
            : "You only have view access in this project. Create actions are disabled.";
    const createButton = (
        <Button
            onClick={() => {
                navigate.modules(getVolumeCreateRoute(scope));
            }}
            disabled={!canCreate}
        >
            <Plus className="size-4" /> New Volume
        </Button>
    );
    const syncButton = (
        <Button
            type="button"
            variant="outline"
            disabled={!clusterPermission.canWrite || isSyncing}
            isLoading={isSyncing}
            onClick={() => {
                syncFromDocker({});
            }}
        >
            <RefreshCw className="size-4" /> Sync From Docker
        </Button>
    );

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        {scope.type === "cluster" &&
                            (clusterPermission.canWrite ? (
                                syncButton
                            ) : (
                                <Tooltip>
                                    <TooltipTrigger asChild>
                                        <span className="inline-flex">{syncButton}</span>
                                    </TooltipTrigger>
                                    <TooltipContent>{deniedMessage}</TooltipContent>
                                </Tooltip>
                            ))}
                        {scope.type === "project" && (
                            <ProjectSettingsImportButton
                                projectId={scope.projectId}
                                settingKind={PROJECT_SETTINGS_IMPORT_KIND.ClusterVolume}
                            />
                        )}
                        {canCreate ? (
                            createButton
                        ) : (
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <span className="inline-flex">{createButton}</span>
                                </TooltipTrigger>
                                <TooltipContent>{deniedMessage}</TooltipContent>
                            </Tooltip>
                        )}
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={volumes}
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

function getVolumeCreateRoute(scope: VolumeManagementScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.clusterResources.volumes.create.$route(scope.projectId);
    }

    return ROUTE.cluster.volumes.create.$route;
}

interface Props {
    scope: VolumeManagementScope;
}
