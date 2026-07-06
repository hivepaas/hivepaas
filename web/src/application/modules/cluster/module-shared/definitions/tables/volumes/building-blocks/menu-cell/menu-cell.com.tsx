import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { Tooltip, TooltipContent, TooltipTrigger } from "@components/ui/tooltip";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ClusterVolumesCommands } from "~/cluster/data/commands";
import { useUpdateVolumeStatusDialog } from "~/cluster/dialogs";
import type { ClusterVolume } from "~/cluster/domain";
import type { VolumeManagementScope } from "~/cluster/module-shared/types";
import { ProjectClusterVolumesCommands } from "~/projects/data/commands";

import { PopConfirm } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule, useConditionalProject } from "@application/shared/permissions";

function View({ volume, scope }: Props) {
    const [open, setOpen] = useState(false);
    const updateStatusDialog = useUpdateVolumeStatusDialog({
        onSuccess: () => {
            setOpen(false);
        },
    });
    const clusterPermission = useConditionalModule({ id: MODULE_IDS.Cluster });
    const projectPermission = useConditionalProject({
        projectId: scope.type === "project" ? scope.projectId : "",
    });
    const isInheritedProjectVolume = scope.type === "project" && volume.inherited === true;

    const { mutate: deleteClusterVolume, isPending: isDeletingClusterVolume } = ClusterVolumesCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Volume removed");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectVolume, isPending: isDeletingProjectVolume } =
        ProjectClusterVolumesCommands.useDeleteOne({
            onSuccess: () => {
                toast.success("Volume removed");
                setOpen(false);
            },
        });

    const canWrite = scope.type === "cluster" ? clusterPermission.canWrite : projectPermission.canWrite;
    const canDelete = scope.type === "cluster" ? clusterPermission.canDelete : projectPermission.canDelete;
    const isDeleting = isDeletingClusterVolume || isDeletingProjectVolume;

    function onChangeStatus() {
        updateStatusDialog.actions.open(scope, volume.id, {
            props: {
                readOnlyInherited: isInheritedProjectVolume,
                entityTitle: volume.name,
            },
        });
    }

    function onDelete() {
        if (!canDelete || isInheritedProjectVolume) {
            return;
        }

        if (scope.type === "cluster") {
            deleteClusterVolume({ volumeID: volume.id });
            return;
        }

        deleteProjectVolume({
            projectID: scope.projectId,
            volumeID: volume.id,
        });
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
                    <Button
                        className="justify-start py-1.5 w-full"
                        variant="ghost"
                        disabled={!canWrite && !isInheritedProjectVolume}
                        onClick={onChangeStatus}
                    >
                        <SlidersHorizontal className="mr-2 size-4" />
                        Change Status
                    </Button>
                    {canDelete && !isInheritedProjectVolume ? (
                        <PopConfirm
                            title="Remove volume"
                            variant="destructive"
                            confirmText="Remove"
                            cancelText="Cancel"
                            description="Confirm removal of this volume?"
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
                                {isInheritedProjectVolume
                                    ? "Inherited volumes are read-only."
                                    : scope.type === "cluster"
                                      ? "You do not have permission to delete cluster volumes."
                                      : "You do not have permission to delete project volumes."}
                            </TooltipContent>
                        </Tooltip>
                    )}
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    volume: ClusterVolume;
    scope: VolumeManagementScope;
}

export const MenuCell = React.memo(View);
