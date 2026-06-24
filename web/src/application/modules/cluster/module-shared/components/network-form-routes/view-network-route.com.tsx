import { Button } from "@components/ui";
import { ClusterNetworksQueries } from "~/cluster/data/queries";
import { ProjectNetworksQueries } from "~/projects/data/queries";

import { AppLoader, RouteFormHeader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import { ViewNetworkForm } from "../../../dialogs/view-network/form";
import { CLUSTER_NETWORK_FORM_ROUTE_PANEL_CLASS } from "../../constants/network-form-layout.constants";

type ViewNetworkScope = { type: "cluster" } | { type: "project"; projectId: string };

export function ViewNetworkRoute({ scope, networkId }: Props) {
    const { navigate } = useAppNavigate();

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

    return (
        <div className="flex w-full flex-col overflow-hidden">
            <RouteFormHeader title="Network info" />

            <div className={CLUSTER_NETWORK_FORM_ROUTE_PANEL_CLASS}>
                {isFetching ? (
                    <div className="flex min-h-[220px] items-center justify-center">
                        <AppLoader />
                    </div>
                ) : network ? (
                    <ViewNetworkForm
                        key={network.id}
                        network={network}
                    />
                ) : (
                    <div className="py-10 text-center text-sm text-muted-foreground">Network not found</div>
                )}

                <div className="shrink-0 px-0 mt-6 flex justify-end">
                    <Button
                        type="button"
                        className="min-w-[120px]"
                        onClick={navigateToList}
                    >
                        Close
                    </Button>
                </div>
            </div>
        </div>
    );
}

function getNetworkListRoute(scope: ViewNetworkScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.clusterResources.networks.$route(scope.projectId);
    }

    return ROUTE.cluster.networks.$route;
}

interface Props {
    scope: ViewNetworkScope;
    networkId: string;
}
