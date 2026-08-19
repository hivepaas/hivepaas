import { Dialog, DialogDescription, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { AppContainerFilesCommands } from "~/projects/data";

import { ImportFilesToContainerForm } from "../form";
import { useImportFilesToContainerDialogState } from "../hooks";
import type { ImportFilesToContainerFormOutput } from "../schemas";
import { mapImportCompressionToWire } from "../schemas";

export function ImportFilesToContainerDialog() {
    const { state, props: dialogOptions, ...actions } = useImportFilesToContainerDialogState();
    const open = state.mode === "open";

    const { mutateAsync: uploadOne, isPending } = AppContainerFilesCommands.useUploadOne();

    function handleClose() {
        if (isPending) {
            return;
        }

        actions.close();
        dialogOptions?.onClose?.();
    }

    async function onSubmit(values: ImportFilesToContainerFormOutput) {
        if (state.mode !== "open") {
            return;
        }

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
            });

            toast.success(response.data.message || "File uploaded successfully");
            actions.close();
            dialogOptions?.onSuccess?.();
        } catch (error) {
            const nextError = error instanceof Error ? error : new Error("Failed to upload container file");
            dialogOptions?.onError?.(nextError);
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
            <DialogFixedContent className="min-w-[390px] w-[800px]">
                <DialogHeader>
                    <DialogTitle>Import Files to Container</DialogTitle>
                    <DialogDescription>
                        Transfer files or unpack compressed archives into the running container
                    </DialogDescription>
                </DialogHeader>

                <ImportFilesToContainerForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                />
            </DialogFixedContent>
        </Dialog>
    );
}
