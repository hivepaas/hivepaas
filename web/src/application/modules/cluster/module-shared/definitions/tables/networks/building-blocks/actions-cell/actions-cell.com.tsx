import React from "react";

import { Button } from "@components/ui";
import { EyeIcon } from "lucide-react";
import type { ClusterNetwork } from "~/cluster/domain";
import type { NetworkManagementScope } from "~/cluster/module-shared/types";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

function View({ network, scope }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <div className="flex items-center justify-center gap-4">
            <Button
                type="button"
                variant="ghost"
                size="icon"
                className="text-link hover:opacity-50 transition-opacity duration-200"
                onClick={() => {
                    navigate.modules(getNetworkDetailsRoute(scope, network.id));
                }}
            >
                <EyeIcon className="size-5" />
                <span className="sr-only">View network</span>
            </Button>
        </div>
    );
}

interface Props {
    network: ClusterNetwork;
    scope: NetworkManagementScope;
}

export const ActionsCell = React.memo(View);

function getNetworkDetailsRoute(scope: NetworkManagementScope, networkId: string) {
    if (scope.type === "project") {
        return ROUTE.projects.single.clusterResources.networks.details.$route(scope.projectId, networkId);
    }

    return ROUTE.cluster.networks.details.$route(networkId);
}
