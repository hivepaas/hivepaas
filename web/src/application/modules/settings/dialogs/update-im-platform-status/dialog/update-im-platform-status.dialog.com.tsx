import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectImServiceCommands } from "~/projects/data/commands";
import { ImServiceCommands } from "~/settings/data/commands";

import { ESettingStatus } from "@application/shared/enums";

import { UpdateImPlatformStatusForm } from "../form";
import { useUpdateImPlatformStatusDialogState } from "../hooks";
import type { UpdateImPlatformStatusFormOutput } from "../schemas";

export function UpdateImPlatformStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateImPlatformStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: updateSettingStatus, isPending: isUpdatingSetting } = ImServiceCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("IM platform status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });

    const { mutate: updateProjectStatus, isPending: isUpdatingProject } = ProjectImServiceCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Project IM platform status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    function onSubmit(values: UpdateImPlatformStatusFormOutput) {
        if (state.mode !== "open") {
            return;
        }

        const payload = {
            updateVer: state.imPlatform.updateVer,
            status: values.status,
            expireAt: values.expireAt ?? null,
            availableInProjects: state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
        };

        if (state.scope.type === "project") {
            updateProjectStatus({
                projectID: state.scope.projectId,
                id: state.imPlatform.id,
                payload,
            });
            return;
        }

        updateSettingStatus({
            id: state.imPlatform.id,
            payload,
        });
    }

    function handleClose() {
        if (isPending) {
            return;
        }

        if (hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) {
            return;
        }

        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const isPending = isUpdatingSetting || isUpdatingProject;
    const showAvailableInProjects = state.mode === "open" && state.scope.type === "settings";
    const initialValues =
        state.mode === "open"
            ? {
                  status:
                      state.imPlatform.status === ESettingStatus.Disabled
                          ? ESettingStatus.Disabled
                          : ESettingStatus.Active,
                  expireAt: state.imPlatform.expireAt ?? undefined,
                  availableInProjects: state.imPlatform.availableInProjects ?? false,
                  default: state.imPlatform.default ?? false,
              }
            : undefined;

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="sm:max-w-[560px]">
                <DialogHeader>
                    <DialogTitle>Change status</DialogTitle>
                </DialogHeader>
                {state.mode === "open" && (
                    <UpdateImPlatformStatusForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        showAvailableInProjects={showAvailableInProjects}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
