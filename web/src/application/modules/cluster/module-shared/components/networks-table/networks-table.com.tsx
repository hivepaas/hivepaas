import { useMemo } from "react";

import { Button } from "@components/ui";
import { Tooltip, TooltipContent, TooltipTrigger } from "@components/ui/tooltip";
import { Plus, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { ClusterNetworksCommands } from "~/cluster/data/commands";
import { ClusterNetworksQueries } from "~/cluster/data/queries";
import { NetworksTableDefs } from "~/cluster/module-shared/definitions/tables/networks/networks-table.defs";
import type { NetworkManagementScope } from "~/cluster/module-shared/types";
import { PROJECT_SETTINGS_IMPORT_KIND } from "~/projects/data/commands";
import { ProjectNetworksQueries } from "~/projects/data/queries";
import { ProjectSettingsImportButton } from "~/settings/module-shared/components";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";
import { useConditionalModule, useConditionalProject } from "@application/shared/permissions";

import { DataTable } from "@/components/ui";

export function NetworkManagementTable({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const clusterPermission = useConditionalModule({ id: MODULE_IDS.Cluster });
    const projectPermission = useConditionalProject({
        projectId: scope.type === "project" ? scope.projectId : "",
    });
    const { navigate } = useAppNavigate();

    const clusterNetworksQuery = ClusterNetworksQueries.useFindManyPaginated(
        {
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "cluster",
        },
    );
    const projectNetworksQuery = ProjectNetworksQueries.useFindManyPaginated(
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

    const { mutate: syncFromDocker, isPending: isSyncing } = ClusterNetworksCommands.useSyncFromDocker({
        onSuccess: () => {
            toast.success("Networks synced from Docker");
        },
    });

    const { data: { data: networks } = DEFAULT_PAGINATED_DATA, isFetching } =
        scope.type === "cluster" ? clusterNetworksQuery : projectNetworksQuery;

    const columns = useMemo(() => NetworksTableDefs.columns(scope), [scope]);
    const canCreate = scope.type === "cluster" ? clusterPermission.canWrite : projectPermission.canWrite;
    const deniedMessage =
        scope.type === "cluster"
            ? "You only have view access. Create actions are disabled."
            : "You only have view access in this project. Create actions are disabled.";
    const createButton = (
        <Button
            onClick={() => {
                navigate.modules(getNetworkCreateRoute(scope));
            }}
            disabled={!canCreate}
        >
            <Plus /> New Network
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
                                settingKind={PROJECT_SETTINGS_IMPORT_KIND.ClusterNetwork}
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
                data={networks}
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
    scope: NetworkManagementScope;
}

function getNetworkCreateRoute(scope: NetworkManagementScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.clusterResources.networks.create.$route(scope.projectId);
    }

    return ROUTE.cluster.networks.create.$route;
}
