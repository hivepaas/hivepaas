import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectBasicAuthCommands } from "~/projects/data/commands";
import { BasicAuthCommands } from "~/settings/data/commands";

import { ESettingStatus } from "@application/shared/enums";

import { UpdateBasicAuthStatusForm } from "../form";
import { useUpdateBasicAuthStatusDialogState } from "../hooks";
import type { UpdateBasicAuthStatusFormOutput } from "../schemas";

export function UpdateBasicAuthStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateBasicAuthStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: updateSettingStatus, isPending: isUpdatingSetting } = BasicAuthCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Basic auth status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });

    const { mutate: updateProjectStatus, isPending: isUpdatingProject } = ProjectBasicAuthCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Project basic auth status updated successfully");
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

    function onSubmit(values: UpdateBasicAuthStatusFormOutput) {
        if (state.mode !== "open") {
            return;
        }

        const payload = {
            updateVer: state.basicAuth.updateVer,
            status: values.status,
            expireAt: values.expireAt ?? null,
            availableInProjects: state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
        };

        if (state.scope.type === "project") {
            updateProjectStatus({
                projectID: state.scope.projectId,
                id: state.basicAuth.id,
                payload,
            });
            return;
        }

        updateSettingStatus({
            id: state.basicAuth.id,
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
                      state.basicAuth.status === ESettingStatus.Disabled
                          ? ESettingStatus.Disabled
                          : ESettingStatus.Active,
                  expireAt: state.basicAuth.expireAt ?? undefined,
                  availableInProjects: state.basicAuth.availableInProjects ?? false,
                  default: state.basicAuth.default ?? false,
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
                    <UpdateBasicAuthStatusForm
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
