import React from "react";

import { Button } from "@components/ui";
import { EyeIcon } from "lucide-react";
import type { ClusterVolume } from "~/cluster/domain";
import type { VolumeManagementScope } from "~/cluster/module-shared/types";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

function View({ volume, scope }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <div className="flex items-center justify-center gap-4">
            <Button
                type="button"
                variant="ghost"
                size="icon"
                className="text-link hover:opacity-50 transition-opacity duration-200"
                onClick={() => {
                    navigate.modules(getVolumeEditRoute(scope, volume.id));
                }}
            >
                <EyeIcon className="size-5" />
                <span className="sr-only">View volume</span>
            </Button>
        </div>
    );
}

interface Props {
    volume: ClusterVolume;
    scope: VolumeManagementScope;
}

export const ActionsCell = React.memo(View);

function getVolumeEditRoute(scope: VolumeManagementScope, volumeId: string) {
    if (scope.type === "project") {
        return ROUTE.projects.single.clusterResources.volumes.edit.$route(scope.projectId, volumeId);
    }

    return ROUTE.cluster.volumes.edit.$route(volumeId);
}
