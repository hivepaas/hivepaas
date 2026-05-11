import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectRegistryAuthCommands } from "~/projects/data/commands";
import { RegistryAuthCommands } from "~/settings/data/commands";

import { ESettingStatus } from "@application/shared/enums";

import { UpdateRegistryAuthStatusForm } from "../form";
import { useUpdateRegistryAuthStatusDialogState } from "../hooks";
import type { UpdateRegistryAuthStatusFormOutput } from "../schemas";

export function UpdateRegistryAuthStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateRegistryAuthStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: updateSettingMeta, isPending: isUpdatingSetting } = RegistryAuthCommands.useUpdateMeta({
        onSuccess: () => {
            toast.success("Registry auth status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });

    const { mutate: updateProjectMeta, isPending: isUpdatingProject } = ProjectRegistryAuthCommands.useUpdateMeta({
        onSuccess: () => {
            toast.success("Project registry auth status updated successfully");
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

    function onSubmit(values: UpdateRegistryAuthStatusFormOutput) {
        if (state.mode !== "open") {
            return;
        }

        const payload = {
            updateVer: state.registryAuth.updateVer,
            status: values.status,
            expireAt: values.expireAt ?? null,
            availableInProjects: state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
        };

        if (state.scope.type === "project") {
            updateProjectMeta({
                projectID: state.scope.projectId,
                id: state.registryAuth.id,
                payload,
            });
            return;
        }

        updateSettingMeta({
            id: state.registryAuth.id,
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
                      state.registryAuth.status === ESettingStatus.Disabled
                          ? ESettingStatus.Disabled
                          : ESettingStatus.Active,
                  expireAt: state.registryAuth.expireAt ?? undefined,
                  availableInProjects: state.registryAuth.availableInProjects ?? false,
                  default: state.registryAuth.default ?? false,
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
                    <UpdateRegistryAuthStatusForm
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
