import React, { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectSecretsCommands } from "~/projects/data/commands";

import { CreateOrEditProjectSecretForm } from "../form";
import { useCreateOrEditProjectSecretDialogState } from "../hooks";
import type { CreateOrEditProjectSecretFormOutput } from "../schemas";

async function fileToBase64(file: File): Promise<string> {
    const buffer = await file.arrayBuffer();
    const bytes = new Uint8Array(buffer);
    const chunkSize = 0x8000;
    let binary = "";

    for (let index = 0; index < bytes.length; index += chunkSize) {
        binary += String.fromCharCode(...bytes.subarray(index, index + chunkSize));
    }

    return window.btoa(binary);
}

async function getSecretValue(values: CreateOrEditProjectSecretFormOutput): Promise<string | undefined> {
    if (values.valueType === "text") {
        return values.textValue.trim() ? values.textValue : undefined;
    }

    if (!values.binaryFile) {
        return undefined;
    }

    return fileToBase64(values.binaryFile);
}

export function CreateOrEditProjectSecretDialog() {
    const { state, props: dialogOptions, ...actions } = useCreateOrEditProjectSecretDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mode } = state;
    const open = mode !== "closed";
    const isEditMode = mode === "edit";
    const projectId = mode !== "closed" ? state.projectId : null;
    const scope = mode !== "closed" ? state.scope : null;

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

    useEffect(() => {
        if (mode === "closed") {
            setHasChanges(false);
        }
    }, [mode]);

    async function onSubmit(values: CreateOrEditProjectSecretFormOutput) {
        if (!projectId || scope !== "project") {
            return;
        }

        const value = await getSecretValue(values);
        const base64 = values.valueType === "binary";

        if (state.mode === "edit") {
            updateProjectSecret({
                projectID: projectId,
                secretID: state.secret.id,
                updateVer: state.secret.updateVer,
                name: values.name,
                value,
                base64,
            });
        } else if (state.mode === "open" && value !== undefined) {
            createProjectSecret({
                projectID: projectId,
                name: values.name,
                value,
                base64,
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

    if (!projectId || scope !== "project") {
        return null;
    }

    const isPending = isCreatingProject || isUpdatingProject;

    const initialValues =
        state.mode === "edit"
            ? {
                  name: state.secret.name,
                  valueType: state.secret.base64 ? ("binary" as const) : ("text" as const),
              }
            : undefined;

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[390px] w-[760px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Create or update a secret</DialogTitle>
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
