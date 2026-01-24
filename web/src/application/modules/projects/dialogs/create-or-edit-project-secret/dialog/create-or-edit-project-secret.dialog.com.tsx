import React, { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectSecretsCommands } from "~/projects/data/commands";

import { CreateOrEditProjectSecretForm } from "../form";
import { useCreateOrEditProjectSecretDialogState } from "../hooks";
import type { CreateOrEditProjectSecretFormOutput } from "../schemas";

export function CreateOrEditProjectSecretDialog() {
    const { state, props, ...actions } = useCreateOrEditProjectSecretDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const open = state.mode !== "closed";
    const isEditMode = state.mode === "edit";
    const projectId = state.mode !== "closed" ? state.projectId : null;

    const { mutate: createSecret, isPending: isCreating } = ProjectSecretsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project secret created successfully");
            actions.close();
        },
    });

    const { mutate: updateSecret, isPending: isUpdating } = ProjectSecretsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project secret updated successfully");
            actions.close();
        },
    });

    // Reset hasChanges when dialog closes
    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
        }
    }, [state.mode]);

    function onSubmit(values: CreateOrEditProjectSecretFormOutput) {
        if (!projectId) {
            return;
        }

        if (isEditMode) {
            updateSecret({
                projectID: projectId,
                secretID: state.secret.id,
                updateVer: state.secret.updateVer,
                name: values.name,
                value: values.value,
            });
        } else {
            // Generate key from name (you may want to adjust this logic)
            createSecret({
                projectID: projectId,
                name: values.name,
                value: values.value,
            });
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

    const isPending = isCreating || isUpdating;

    // Prepare initial values for edit mode
    const initialValues =
        isEditMode
            ? {
                name: state.secret.name,
                value: "", // Value is not returned for security reasons
            }
            : undefined;

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[500px] max-w-[600px]">
                <DialogHeader>
                    <DialogTitle>{isEditMode ? "Edit Secret" : "Create Secret"}</DialogTitle>
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
