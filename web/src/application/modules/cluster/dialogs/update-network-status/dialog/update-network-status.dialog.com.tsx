import { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ClusterNetworksCommands } from "~/cluster/data/commands";
import { ClusterNetworksQueries } from "~/cluster/data/queries";
import { ProjectNetworksCommands } from "~/projects/data/commands";
import { ProjectNetworksQueries } from "~/projects/data/queries";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { ESettingStatus } from "@application/shared/enums";
import { useConditionalModule, useConditionalProject } from "@application/shared/permissions";

import { UpdateNetworkStatusForm } from "../form";
import { useUpdateNetworkStatusDialogState } from "../hooks";
import type { UpdateNetworkStatusFormOutput } from "../schemas";

export function UpdateNetworkStatusDialog() {
    const { state, props: dialogOptions, close: closeDialog, clear: clearDialog } = useUpdateNetworkStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const clusterPermission = useConditionalModule({ id: MODULE_IDS.Cluster });
    const projectPermission = useConditionalProject({
        projectId: state.mode === "open" && state.scope.type === "project" ? state.scope.projectId : "",
    });
    const canWrite =
        state.mode === "open" && state.scope.type === "project"
            ? projectPermission.canWrite
            : clusterPermission.canWrite;

    const { mutate: updateClusterStatus, isPending: isUpdatingCluster } = ClusterNetworksCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Network status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });

    const { mutate: updateProjectStatus, isPending: isUpdatingProject } = ProjectNetworksCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Project network status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    const detailId = state.mode === "open" ? state.id : "";
    const clusterDetailQuery = ClusterNetworksQueries.useFindOneById(
        { networkID: detailId },
        {
            enabled: state.mode === "open" && state.scope.type === "cluster",
        },
    );
    const projectDetailQuery = ProjectNetworksQueries.useFindOneById(
        {
            projectID: state.mode === "open" && state.scope.type === "project" ? state.scope.projectId : "",
            networkID: detailId,
        },
        {
            enabled: state.mode === "open" && state.scope.type === "project",
        },
    );
    const detailQuery =
        state.mode === "open" && state.scope.type === "project" ? projectDetailQuery : clusterDetailQuery;
    const network = detailQuery.data?.data;

    function onSubmit(values: UpdateNetworkStatusFormOutput) {
        if (state.mode !== "open" || !network) {
            return;
        }

        const payload = {
            updateVer: network.updateVer,
            status: values.status,
            expireAt: values.expireAt ?? null,
            default: values.default,
        };

        if (state.scope.type === "project") {
            updateProjectStatus({
                projectID: state.scope.projectId,
                networkID: network.id,
                payload,
            });
            return;
        }

        updateClusterStatus({
            networkID: network.id,
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
        resolvedDialogOptions.readOnlyInherited === true || (state.mode === "open" && network?.inherited === true);
    const dialogTitle = readOnlyInherited
        ? `${resolvedDialogOptions.entityTitle ?? "Network"} Status`
        : "Change status";
    const isPending = isUpdatingCluster || isUpdatingProject;
    const showAvailableInProjects = state.mode === "open" && state.scope.type === "cluster";
    const initialValues = network
        ? {
              status: network.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: network.expireAt ?? undefined,
              availableInProjects: network.availableInProjects,
              default: network.default,
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
                    <UpdateNetworkStatusForm
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
