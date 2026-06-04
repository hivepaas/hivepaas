import { Button } from "@components/ui";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ClusterNetworksCommands } from "~/cluster/data/commands";
import { EClusterNetworkDriver } from "~/cluster/module-shared/enums";
import { ProjectNetworksCommands } from "~/projects/data/commands";

import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule, useConditionalProject } from "@application/shared/permissions";

import { CreateNetworkForm } from "../form";
import { useCreateNetworkDialogState } from "../hooks";
import type { CreateNetworkFormOutput } from "../schemas";

function toRecord(items: { key: string; value: string }[]) {
    return items.reduce<Record<string, string>>((acc, item) => {
        const key = item.key.trim();
        if (!key) {
            return acc;
        }

        return {
            ...acc,
            [key]: item.value.trim(),
        };
    }, {});
}

export function CreateNetworkDialog() {
    const { state, props: dialogOptions, ...actions } = useCreateNetworkDialogState();
    const open = state.mode !== "closed";
    const scope = state.mode === "open" ? state.scope : null;
    const clusterPermission = useConditionalModule({ id: MODULE_IDS.Cluster });
    const projectPermission = useConditionalProject({
        projectId: scope?.type === "project" ? scope.projectId : "",
    });
    const canWrite = scope?.type === "cluster" ? clusterPermission.canWrite : projectPermission.canWrite;

    const { mutate: createClusterNetwork, isPending: isCreatingClusterNetwork } = ClusterNetworksCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Network created");
            actions.close();
            dialogOptions?.onSuccess?.();
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });
    const { mutate: createProjectNetwork, isPending: isCreatingProjectNetwork } = ProjectNetworksCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Network created");
            actions.close();
            dialogOptions?.onSuccess?.();
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    const isPending = isCreatingClusterNetwork || isCreatingProjectNetwork;

    function onSubmit(values: CreateNetworkFormOutput) {
        if (!scope || !canWrite) {
            return;
        }

        const payload = {
            name: values.name,
            driver: EClusterNetworkDriver.Overlay,
            enableIPv4: values.enableIPv4,
            enableIPv6: values.enableIPv6,
            internal: values.internal,
            attachable: values.attachable,
            ingress: values.ingress,
            labels: toRecord(values.labels),
            options: toRecord(values.options),
        };

        if (scope.type === "cluster") {
            createClusterNetwork({
                payload: {
                    ...payload,
                    availableInProjects: values.availableInProjects,
                },
            });
            return;
        }

        createProjectNetwork({
            projectID: scope.projectId,
            payload,
        });
    }

    function handleClose(): void {
        if (isPending) {
            return;
        }

        actions.close();
        dialogOptions?.onClose?.();
    }

    return (
        <Dialog
            open={open}
            onOpenChange={nextOpen => {
                if (!nextOpen) {
                    handleClose();
                }
            }}
        >
            <DialogContent className="sm:max-w-[820px]">
                <DialogHeader>
                    <DialogTitle className="text-2xl">Create a network</DialogTitle>
                </DialogHeader>
                <CreateNetworkForm
                    readOnly={!canWrite}
                    isPending={isPending}
                    showAvailableInProjects={scope?.type !== "project"}
                    onSubmit={onSubmit}
                >
                    <div className="flex justify-end pt-4">
                        <Button
                            type="submit"
                            disabled={!canWrite || isPending}
                            className="min-w-[120px]"
                        >
                            Save
                        </Button>
                    </div>
                </CreateNetworkForm>
            </DialogContent>
        </Dialog>
    );
}
