import { Dialog, DialogDescription, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { AppContainerFilesCommands } from "~/projects/data";

import { ExportContainerFilesForm } from "../form";
import { useExportContainerFilesDialogState } from "../hooks";
import type { ExportContainerFilesFormOutput } from "../schemas";
import { mapExportCompressionToWire } from "../schemas";

function saveBlob(blob: Blob, filename: string) {
    const url = window.URL.createObjectURL(blob);
    const anchor = document.createElement("a");

    anchor.href = url;
    anchor.download = filename;
    document.body.append(anchor);
    anchor.click();
    anchor.remove();
    window.URL.revokeObjectURL(url);
}

export function ExportContainerFilesDialog() {
    const { state, props: dialogOptions, ...actions } = useExportContainerFilesDialogState();
    const open = state.mode === "open";

    const { mutateAsync: downloadOne, isPending } = AppContainerFilesCommands.useDownloadOne({
        onSuccess: response => {
            saveBlob(response.data.blob, response.data.filename ?? "download");
            actions.close();
            dialogOptions?.onSuccess?.();
        },
    });

    function handleClose() {
        if (isPending) {
            return;
        }

        actions.close();
        dialogOptions?.onClose?.();
    }

    async function onSubmit(values: ExportContainerFilesFormOutput) {
        if (state.mode !== "open") {
            return;
        }

        await downloadOne({
            projectID: state.projectId,
            env: state.env,
            appID: state.appId,
            nodeId: state.nodeId,
            containerId: state.containerId,
            path: values.path,
            isDir: values.isDir,
            compressionFormat: mapExportCompressionToWire(values.compression),
        });
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
                    <DialogTitle>Export Container Files</DialogTitle>
                    <DialogDescription>
                        Retrieve files or compressed folder archives from the running container
                    </DialogDescription>
                </DialogHeader>

                <ExportContainerFilesForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                />
            </DialogFixedContent>
        </Dialog>
    );
}
