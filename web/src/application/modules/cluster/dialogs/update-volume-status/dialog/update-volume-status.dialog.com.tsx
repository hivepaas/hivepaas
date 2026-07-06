import { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ClusterVolumesCommands } from "~/cluster/data/commands";
import { ClusterVolumesQueries } from "~/cluster/data/queries";
import { ProjectClusterVolumesCommands } from "~/projects/data/commands";
import { ProjectClusterVolumesQueries } from "~/projects/data/queries";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { ESettingStatus } from "@application/shared/enums";
import { useConditionalModule, useConditionalProject } from "@application/shared/permissions";

import { UpdateVolumeStatusForm } from "../form";
import { useUpdateVolumeStatusDialogState } from "../hooks";
import type { UpdateVolumeStatusFormOutput } from "../schemas";

export function UpdateVolumeStatusDialog() {
    const { state, props: dialogOptions, close: closeDialog, clear: clearDialog } = useUpdateVolumeStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const clusterPermission = useConditionalModule({ id: MODULE_IDS.Cluster });
    const projectPermission = useConditionalProject({
        projectId: state.mode === "open" && state.scope.type === "project" ? state.scope.projectId : "",
    });
    const canWrite =
        state.mode === "open" && state.scope.type === "project"
            ? projectPermission.canWrite
            : clusterPermission.canWrite;

    const { mutate: updateClusterStatus, isPending: isUpdatingCluster } = ClusterVolumesCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Volume status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });

    const { mutate: updateProjectStatus, isPending: isUpdatingProject } = ProjectClusterVolumesCommands.useUpdateStatus(
        {
            onSuccess: () => {
                toast.success("Project volume status updated successfully");
                closeDialog();
                dialogOptions?.onSuccess?.();
            },
        },
    );

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    const detailId = state.mode === "open" ? state.id : "";
    const clusterDetailQuery = ClusterVolumesQueries.useFindOneById(
        { volumeID: detailId },
        {
            enabled: state.mode === "open" && state.scope.type === "cluster",
        },
    );
    const projectDetailQuery = ProjectClusterVolumesQueries.useFindOneById(
        {
            projectID: state.mode === "open" && state.scope.type === "project" ? state.scope.projectId : "",
            volumeID: detailId,
        },
        {
            enabled: state.mode === "open" && state.scope.type === "project",
        },
    );
    const detailQuery =
        state.mode === "open" && state.scope.type === "project" ? projectDetailQuery : clusterDetailQuery;
    const volume = detailQuery.data?.data;

    function onSubmit(values: UpdateVolumeStatusFormOutput) {
        if (state.mode !== "open" || !volume) {
            return;
        }

        const payload = {
            updateVer: volume.updateVer,
            status: values.status,
            expireAt: values.expireAt ?? null,
            default: values.default,
        };

        if (state.scope.type === "project") {
            updateProjectStatus({
                projectID: state.scope.projectId,
                volumeID: volume.id,
                payload,
            });
            return;
        }

        updateClusterStatus({
            volumeID: volume.id,
            payload: {
                ...payload,
                availableInProjects: values.availableInProjects,
            },
        });
    }

    function handleClose() {
        if (isPending) {
            return;
        }

        if (
            !readOnlyInherited &&
            canWrite &&
            hasChanges &&
            !window.confirm("Are you sure you want to close without saving changes?")
        ) {
            return;
        }

        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const resolvedDialogOptions = dialogOptions ?? {};
    const readOnlyInherited =
        resolvedDialogOptions.readOnlyInherited === true || (state.mode === "open" && volume?.inherited === true);
    const dialogTitle = readOnlyInherited ? `${resolvedDialogOptions.entityTitle ?? "Volume"} Status` : "Change status";
    const isPending = isUpdatingCluster || isUpdatingProject;
    const showAvailableInProjects = state.mode === "open" && state.scope.type === "cluster";
    const initialValues = volume
        ? {
              status: volume.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: volume.expireAt ?? undefined,
              availableInProjects: volume.availableInProjects ?? false,
              default: volume.default ?? false,
          }
        : undefined;
    const isDetailLoading = state.mode === "open" && detailQuery.isFetching;

    return (
        <Dialog
            open={open}
            onOpenChange={nextOpen => {
                if (!nextOpen) {
                    handleClose();
                }
            }}
        >
            <DialogFixedContent className="sm:max-w-[560px]">
                <DialogHeader>
                    <DialogTitle>{dialogTitle}</DialogTitle>
                </DialogHeader>
                {isDetailLoading && (
                    <DialogBody>
                        <AppLoader />
                    </DialogBody>
                )}
                {state.mode === "open" && !isDetailLoading && initialValues && (
                    <UpdateVolumeStatusForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        showAvailableInProjects={showAvailableInProjects}
                        readOnlyInherited={readOnlyInherited}
                        readOnly={!canWrite}
                        onClose={handleClose}
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
