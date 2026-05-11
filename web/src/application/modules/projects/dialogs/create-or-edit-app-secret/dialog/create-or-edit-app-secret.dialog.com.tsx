import React, { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectAppSecretsCommands } from "~/projects/data/commands";

import { CreateOrEditAppSecretForm } from "../form";
import { useCreateOrEditAppSecretDialogState } from "../hooks";
import type { CreateOrEditAppSecretFormOutput } from "../schemas";

function fileToBase64(file: File): Promise<string> {
    return file.arrayBuffer().then(buffer => {
        let binary = "";
        const bytes = new Uint8Array(buffer);
        const chunkSize = 0x8000;

        for (let index = 0; index < bytes.length; index += chunkSize) {
            const chunk = bytes.subarray(index, index + chunkSize);
            binary += String.fromCharCode(...chunk);
        }

        return window.btoa(binary);
    });
}

async function getSecretValue(values: CreateOrEditAppSecretFormOutput): Promise<string | undefined> {
    if (values.valueType === "text") {
        return values.textValue.trim() ? values.textValue : undefined;
    }

    if (!values.binaryFile) {
        return undefined;
    }

    return fileToBase64(values.binaryFile);
}

function getSwarmRef(values: CreateOrEditAppSecretFormOutput) {
    if (!values.mountIntoFilesystem) {
        return undefined;
    }

    return {
        file: {
            name: values.filePath,
            mode: values.fileMode,
            uid: values.fileUid,
            gid: values.fileGid,
        },
    };
}

export function CreateOrEditAppSecretDialog() {
    const { state, props: dialogOptions, ...actions } = useCreateOrEditAppSecretDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mode } = state;
    const open = mode !== "closed";
    const isEditMode = mode === "edit";
    const projectId = mode !== "closed" ? state.projectId : null;
    const appId = mode !== "closed" ? state.appId : null;

    const { mutate: createAppSecret, isPending: isCreatingApp } = ProjectAppSecretsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("App secret created successfully");
            actions.close();
            dialogOptions?.onSuccess?.();
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    const { mutate: updateAppSecret, isPending: isUpdatingApp } = ProjectAppSecretsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("App secret updated successfully");
            actions.close();
            dialogOptions?.onSuccess?.();
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    useEffect(() => {
        if (mode === "closed") {
            setHasChanges(false);
        }
    }, [mode]);

    async function onSubmit(values: CreateOrEditAppSecretFormOutput) {
        if (!projectId || !appId) {
            return;
        }

        const value = await getSecretValue(values);
        const base64 = values.valueType === "binary";
        const swarmRef = getSwarmRef(values);

        if (state.mode === "edit") {
            updateAppSecret({
                projectID: projectId,
                appID: appId,
                secretID: state.secret.id,
                updateVer: state.secret.updateVer,
                name: values.name,
                value,
                base64,
                swarmRef,
            });
            return;
        }

        if (state.mode === "open" && value !== undefined) {
            createAppSecret({
                projectID: projectId,
                appID: appId,
                name: values.name,
                value,
                base64,
                swarmRef,
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
        dialogOptions?.onClose?.();
    }

    if (!projectId || !appId) {
        return null;
    }

    const isPending = isCreatingApp || isUpdatingApp;

    const initialValues =
        state.mode === "edit"
            ? {
                  name: state.secret.name,
                  valueType: state.secret.base64 ? ("binary" as const) : ("text" as const),
                  mountIntoFilesystem: Boolean(state.secret.swarmRef?.file),
                  filePath: state.secret.swarmRef?.file?.name,
                  fileMode: state.secret.swarmRef?.file?.mode,
                  fileUid: state.secret.swarmRef?.file?.uid,
                  fileGid: state.secret.swarmRef?.file?.gid,
              }
            : undefined;

    return (
        <Dialog
            open={open}
            onOpenChange={isOpen => {
                if (!isOpen) {
                    handleClose();
                }
            }}
        >
            <DialogContent className="min-w-[390px] w-[760px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Create or update a secret</DialogTitle>
                </DialogHeader>
                <CreateOrEditAppSecretForm
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
