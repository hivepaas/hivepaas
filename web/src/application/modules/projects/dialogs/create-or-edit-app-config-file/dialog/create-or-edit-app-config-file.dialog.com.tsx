import React, { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { AppConfigFilesCommands } from "~/projects/data/commands";

import { CreateOrEditAppConfigFileForm } from "../form";
import { useCreateOrEditAppConfigFileDialogState } from "../hooks";
import type { CreateOrEditAppConfigFileFormOutput } from "../schemas";

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

async function getConfigFileContent(values: CreateOrEditAppConfigFileFormOutput): Promise<string | undefined> {
    if (values.valueType === "text") {
        return values.textValue.trim() ? values.textValue : undefined;
    }

    if (!values.binaryFile) {
        return undefined;
    }

    return fileToBase64(values.binaryFile);
}

function getSwarmRef(values: CreateOrEditAppConfigFileFormOutput) {
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

export function CreateOrEditAppConfigFileDialog() {
    const { state, props: dialogOptions, ...actions } = useCreateOrEditAppConfigFileDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mode } = state;
    const open = mode !== "closed";
    const isEditMode = mode === "edit";
    const projectId = mode !== "closed" ? state.projectId : null;
    const appId = mode !== "closed" ? state.appId : null;

    const { mutate: createAppConfigFile, isPending: isCreatingAppConfigFile } = AppConfigFilesCommands.useCreateOne({
        onSuccess: () => {
            toast.success("App config file created successfully");
            actions.close();
            dialogOptions?.onSuccess?.();
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    const { mutate: updateAppConfigFile, isPending: isUpdatingAppConfigFile } = AppConfigFilesCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("App config file updated successfully");
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

    async function onSubmit(values: CreateOrEditAppConfigFileFormOutput) {
        if (!projectId || !appId) {
            return;
        }

        const content = await getConfigFileContent(values);
        const base64 = values.valueType === "binary";
        const swarmRef = getSwarmRef(values);

        if (state.mode === "edit") {
            updateAppConfigFile({
                projectID: projectId,
                appID: appId,
                configFileID: state.configFile.id,
                updateVer: state.configFile.updateVer,
                name: values.name,
                content,
                base64,
                swarmRef,
            });
            return;
        }

        if (state.mode === "open" && content !== undefined) {
            createAppConfigFile({
                projectID: projectId,
                appID: appId,
                name: values.name,
                content,
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

    const isPending = isCreatingAppConfigFile || isUpdatingAppConfigFile;

    const initialValues =
        state.mode === "edit"
            ? {
                  name: state.configFile.name,
                  valueType: state.configFile.base64 ? ("binary" as const) : ("text" as const),
                  mountIntoFilesystem: Boolean(state.configFile.swarmRef?.file),
                  filePath: state.configFile.swarmRef?.file?.name,
                  fileMode: state.configFile.swarmRef?.file?.mode,
                  fileUid: state.configFile.swarmRef?.file?.uid,
                  fileGid: state.configFile.swarmRef?.file?.gid,
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
            <DialogContent className="min-w-[390px] w-[820px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Create or update a config file</DialogTitle>
                </DialogHeader>
                <CreateOrEditAppConfigFileForm
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
