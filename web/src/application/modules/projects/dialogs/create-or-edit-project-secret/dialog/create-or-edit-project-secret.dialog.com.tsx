import React, { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectAppSecretsCommands, ProjectSecretsCommands } from "~/projects/data/commands";

import { CreateOrEditProjectSecretForm } from "../form";
import { useCreateOrEditProjectSecretDialogState } from "../hooks";
import type { CreateOrEditProjectSecretFormOutput } from "../schemas";

export function CreateOrEditProjectSecretDialog() {
    const { state, props: dialogOptions, ...actions } = useCreateOrEditProjectSecretDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mode } = state;
    const open = mode !== "closed";
    const isEditMode = mode === "edit";
    const projectId = mode !== "closed" ? state.projectId : null;
    const appId = mode !== "closed" ? state.appId : null;
    const scope = mode !== "closed" ? state.scope : null;

    // Project mutations
    const { mutate: createProjectSecret, isPending: isCreatingProject } = ProjectSecretsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project secret created successfully");
            actions.close();
            dialogOptions?.onSuccess?.();
        },
    });

    const { mutate: updateProjectSecret, isPending: isUpdatingProject } = ProjectSecretsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project secret updated successfully");
            actions.close();
            dialogOptions?.onSuccess?.();
        },
    });

    // App mutations
    const { mutate: createAppSecret, isPending: isCreatingApp } = ProjectAppSecretsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("App secret created successfully");
            actions.close();
            dialogOptions?.onSuccess?.();
        },
    });

    const { mutate: updateAppSecret, isPending: isUpdatingApp } = ProjectAppSecretsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("App secret updated successfully");
            actions.close();
            dialogOptions?.onSuccess?.();
        },
    });

    // Reset hasChanges when dialog closes
    useEffect(() => {
        if (mode === "closed") {
            setHasChanges(false);
        }
    }, [mode]);

    function onSubmit(values: CreateOrEditProjectSecretFormOutput) {
        if (!projectId) {
            return;
        }

        if (scope === "project") {
            if (state.mode === "edit") {
                updateProjectSecret({
                    projectID: projectId,
                    secretID: state.secret.id,
                    updateVer: state.secret.updateVer,
                    name: values.name,
                    value: values.value,
                });
            } else if (state.mode === "open") {
                createProjectSecret({
                    projectID: projectId,
                    name: values.name,
                    value: values.value,
                });
            }
        } else if (scope === "app" && appId) {
            if (state.mode === "edit") {
                updateAppSecret({
                    projectID: projectId,
                    appID: appId,
                    secretID: state.secret.id,
                    updateVer: state.secret.updateVer,
                    name: values.name,
                    value: values.value,
                });
            } else if (state.mode === "open") {
                createAppSecret({
                    projectID: projectId,
                    appID: appId,
                    name: values.name,
                    value: values.value,
                });
            }
        }
    }

    function handleClose(): void {
        if (hasChanges) {
            const userConfirmed: boolean = window.confirm("Are you sure you want to close without saving changes?");
            if (!userConfirmed) {
                return;
            }
        }

        setHasChanges(false);
        actions.close();
    }

    if (!projectId) {
        return null;
    }

    const isPending = isCreatingProject || isUpdatingProject || isCreatingApp || isUpdatingApp;

    // Prepare initial values for edit mode
    const initialValues =
        state.mode === "edit"
            ? {
                  name: state.secret.name,
                  value: "", // Value is not returned for security reasons
              }
            : undefined;

    const titlePrefix = scope === "app" ? "App" : "Project";

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[390px] w-[650px]">
                <DialogHeader>
                    <DialogTitle>
                        {isEditMode ? "Edit" : "Create"} {titlePrefix} Secret
                    </DialogTitle>
                </DialogHeader>
                <CreateOrEditProjectSecretForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                    isEditMode={isEditMode}
                    initialValues={initialValues}
                />
            </DialogContent>
        </Dialog>
    );
}
