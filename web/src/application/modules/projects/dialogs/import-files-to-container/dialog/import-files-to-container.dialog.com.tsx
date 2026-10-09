import { useRef, useState } from "react";

import { Dialog, DialogDescription, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { AppContainerFilesCommands } from "~/projects/data";

import { isCancelException } from "@infrastructure/api";

import { ImportFilesToContainerForm, type UploadProgressValue } from "../form";
import { useImportFilesToContainerDialogState } from "../hooks";
import type { ImportFilesToContainerFormOutput } from "../schemas";
import { mapImportCompressionToWire } from "../schemas";

export function ImportFilesToContainerDialog() {
    const { state, props: dialogOptions, ...actions } = useImportFilesToContainerDialogState();
    const open = state.mode === "open";

    const { mutateAsync: uploadOne, isPending } = AppContainerFilesCommands.useUploadOne();
    const [progress, setProgress] = useState<UploadProgressValue | null>(null);
    // What cancels the upload going on, if one is.
    const uploading = useRef<AbortController | null>(null);

    function cancelUpload() {
        uploading.current?.abort();
    }

    // Closed while uploading, the dialog cancels the upload.
    function handleClose() {
        cancelUpload();
        actions.close();
        dialogOptions?.onClose?.();
    }

    async function onSubmit(values: ImportFilesToContainerFormOutput) {
        if (state.mode !== "open") {
            return;
        }

        setProgress(null);
        const controller = new AbortController();
        uploading.current = controller;
        try {
            const response = await uploadOne({
                projectID: state.projectId,
                env: state.env,
                appID: state.appId,
                nodeId: state.nodeId,
                containerId: state.containerId,
                path: values.path,
                file: values.file,
                extract: values.extract,
                compressionFormat: mapImportCompressionToWire(values.compression),
                overwrite: values.overwrite,
                onProgress: (sent, total) => {
                    setProgress({ sent, total });
                },
                signal: controller.signal,
            });

            toast.success(response.data.message || "File uploaded successfully");
            actions.close();
            dialogOptions?.onSuccess?.();
        } catch (error) {
            const nextError = error instanceof Error ? error : new Error("Failed to upload container file");
            if (isCancelException(nextError)) {
                toast.info("Upload cancelled");
                return;
            }
            dialogOptions?.onError?.(nextError);
        } finally {
            uploading.current = null;
            setProgress(null);
        }
    }

    return (
        <Dialog
            open={open}
            onOpenChange={isOpen => {
                if (!isOpen) {
                    handleClose();
                }
            }}
        >
            <DialogFixedContent className="sm:max-w-[680px]">
                <DialogHeader>
                    <DialogTitle>Import Files to Container</DialogTitle>
                    <DialogDescription>
                        Transfer files or unpack compressed archives into the running container
                    </DialogDescription>
                </DialogHeader>

                <ImportFilesToContainerForm
                    isPending={isPending}
                    progress={progress}
                    onSubmit={onSubmit}
                    onCancel={cancelUpload}
                />
            </DialogFixedContent>
        </Dialog>
    );
}
