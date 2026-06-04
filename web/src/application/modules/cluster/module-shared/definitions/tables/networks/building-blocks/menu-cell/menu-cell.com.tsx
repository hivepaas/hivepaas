import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { Tooltip, TooltipContent, TooltipTrigger } from "@components/ui/tooltip";
import { MoreVertical, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ClusterNetworksCommands } from "~/cluster/data/commands";
import type { ClusterNetwork } from "~/cluster/domain";
import type { NetworkManagementScope } from "~/cluster/module-shared/types";
import { ProjectNetworksCommands } from "~/projects/data/commands";

import { PopConfirm } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule, useConditionalProject } from "@application/shared/permissions";

function View({ network, scope }: Props) {
    const [open, setOpen] = useState(false);
    const { canDelete: canDeleteCluster } = useConditionalModule({ id: MODULE_IDS.Cluster });
    const projectPermission = useConditionalProject({
        projectId: scope.type === "project" ? scope.projectId : "",
    });

    const { mutate: deleteOneClusterNetwork, isPending: isDeletingClusterNetwork } =
        ClusterNetworksCommands.useDeleteOne({
            onSuccess: () => {
                toast.success("Network removed");
                setOpen(false);
            },
        });

    const { mutate: deleteOneProjectNetwork, isPending: isDeletingProjectNetwork } =
        ProjectNetworksCommands.useDeleteOne({
            onSuccess: () => {
                toast.success("Network removed");
                setOpen(false);
            },
        });

    const isInheritedProjectNetwork = scope.type === "project" && network.availableInProjects;
    const canDelete = scope.type === "cluster" ? canDeleteCluster : projectPermission.canDelete;
    const isDeleting = isDeletingClusterNetwork || isDeletingProjectNetwork;

    function onDelete() {
        if (!canDelete) {
            return;
        }

        if (scope.type === "cluster") {
            deleteOneClusterNetwork({ networkID: network.id });
            return;
        }

        deleteOneProjectNetwork({
            projectID: scope.projectId,
            networkID: network.id,
        });
    }

    if (isInheritedProjectNetwork) {
        return null;
    }

    return (
        <DropdownMenu
            open={open}
            onOpenChange={setOpen}
        >
            <DropdownMenuTrigger
                asChild
                className="h-8 w-8"
            >
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                >
                    <MoreVertical className="size-4" />
                    <span className="sr-only">Actions menu</span>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <div className="flex flex-col gap-0">
                    {canDelete ? (
                        <PopConfirm
                            title="Remove network"
                            variant="destructive"
                            confirmText="Remove"
                            cancelText="Cancel"
                            description="Confirm removal of this network?"
                            onConfirm={onDelete}
                        >
                            <Button
                                className="justify-start py-1.5"
                                variant="ghost"
                                disabled={isDeleting}
                            >
                                <Trash2Icon className="mr-2 size-4" />
                                Remove
                            </Button>
                        </PopConfirm>
                    ) : (
                        <Tooltip>
                            <TooltipTrigger
                                asChild
                                className="w-full"
                            >
                                <span className="inline-flex w-full">
                                    <Button
                                        className="justify-start py-1.5 w-full"
                                        variant="ghost"
                                        disabled
                                    >
                                        <Trash2Icon className="mr-2 size-4" />
                                        Remove
                                    </Button>
                                </span>
                            </TooltipTrigger>
                            <TooltipContent side="left">
                                {scope.type === "cluster"
                                    ? "You do not have permission to delete cluster networks."
                                    : "You do not have permission to delete project networks."}
                            </TooltipContent>
                        </Tooltip>
                    )}
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    network: ClusterNetwork;
    scope: NetworkManagementScope;
}

export const MenuCell = React.memo(View);
