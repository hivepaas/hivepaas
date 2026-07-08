import { Button } from "@components/ui";
import { toast } from "sonner";
import { ClusterNetworksCommands } from "~/cluster/data/commands";
import { ClusterNetworksQueries } from "~/cluster/data/queries";
import type { NetworkManagementScope } from "~/cluster/module-shared/types";
import { ProjectNetworksCommands } from "~/projects/data/commands";
import { ProjectNetworksQueries } from "~/projects/data/queries";

import { AppLoader, FormActionBar, RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule, useConditionalProject } from "@application/shared/permissions";

import { ViewNetworkForm, type ViewNetworkFormOutput } from "../../../dialogs/view-network/form";

export function ViewNetworkRoute({ scope, networkId }: Props) {
    const { navigate } = useAppNavigate();
    const clusterPermission = useConditionalModule({ id: MODULE_IDS.Cluster });
    const projectPermission = useConditionalProject({
        projectId: scope.type === "project" ? scope.projectId : "",
    });
    const canWrite = scope.type === "cluster" ? clusterPermission.canWrite : projectPermission.canWrite;

    function navigateToList() {
        navigate.modules(getNetworkListRoute(scope), { ignorePrevPath: true });
    }

    const clusterNetworkQuery = ClusterNetworksQueries.useFindOneById(
        { networkID: networkId },
        {
            enabled: scope.type === "cluster" && Boolean(networkId),
        },
    );
    const projectNetworkQuery = ProjectNetworksQueries.useFindOneById(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            networkID: networkId,
        },
        {
            enabled: scope.type === "project" && Boolean(networkId),
        },
    );
    const network = scope.type === "cluster" ? clusterNetworkQuery.data?.data : projectNetworkQuery.data?.data;
    const isFetching = scope.type === "cluster" ? clusterNetworkQuery.isFetching : projectNetworkQuery.isFetching;
    const isInherited = scope.type === "project" && network?.inherited === true;
    const canSubmit = canWrite && !isInherited;

    const { mutate: updateClusterNetwork, isPending: isUpdatingClusterNetwork } = ClusterNetworksCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Network updated");
            navigateToList();
        },
    });
    const { mutate: updateProjectNetwork, isPending: isUpdatingProjectNetwork } = ProjectNetworksCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Network updated");
            navigateToList();
        },
    });

    const isPending = isUpdatingClusterNetwork || isUpdatingProjectNetwork;

    function onSubmit(values: ViewNetworkFormOutput) {
        if (!canSubmit || !network) {
            return;
        }

        if (scope.type === "cluster") {
            updateClusterNetwork({
                networkID: network.id,
                payload: {
                    updateVer: network.updateVer,
                    availableInProjects: values.availableInProjects,
                    default: values.default,
                },
            });
            return;
        }

        updateProjectNetwork({
            projectID: scope.projectId,
            networkID: network.id,
            payload: {
                updateVer: network.updateVer,
                default: values.default,
            },
        });
    }

    return (
        <div className="flex w-full flex-col">
            <RouteFormHeader title="Update network" />

            {isFetching ? (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            ) : network ? (
                <ViewNetworkForm
                    key={network.id}
                    network={network}
                    readOnlyAvailableInProjects={scope.type === "project" || !canSubmit}
                    readOnlyDefault={!canSubmit}
                    readOnlyInherited={isInherited}
                    readOnlyPermission={!canWrite}
                    isPending={isPending}
                    showAvailableInProjects={scope.type === "cluster"}
                    onSubmit={onSubmit}
                >
                    {!canSubmit ? (
                        <FormActionBar>
                            <Button
                                type="button"
                                onClick={navigateToList}
                            >
                                Close
                            </Button>
                        </FormActionBar>
                    ) : (
                        <FormActionBar>
                            <Button
                                type="button"
                                variant="outline"
                                className="min-w-[100px]"
                                disabled={isPending}
                                onClick={navigateToList}
                            >
                                Cancel
                            </Button>
                            <Button
                                type="submit"
                                disabled={isPending}
                                className="min-w-[120px]"
                            >
                                Save
                            </Button>
                        </FormActionBar>
                    )}
                </ViewNetworkForm>
            ) : (
                <div className="py-10 text-center text-sm text-muted-foreground">Network not found</div>
            )}
        </div>
    );
}

function getNetworkListRoute(scope: NetworkManagementScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.clusterResources.networks.$route(scope.projectId);
    }

    return ROUTE.cluster.networks.$route;
}

interface Props {
    scope: NetworkManagementScope;
    networkId: string;
}
