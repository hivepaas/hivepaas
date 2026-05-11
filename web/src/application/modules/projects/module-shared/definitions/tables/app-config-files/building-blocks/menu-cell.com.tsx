import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { DownloadIcon, Edit2Icon, MoreVertical, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { useAppConfigFilesApi } from "~/projects/api/hooks/project-apps";
import { AppConfigFilesCommands } from "~/projects/data/commands";
import { useCreateOrEditAppConfigFileDialog } from "~/projects/dialogs/create-or-edit-app-config-file/hooks";
import type { AppConfigFile } from "~/projects/domain";

import { PopConfirm } from "@application/shared/components";

function base64ToArrayBuffer(value: string): ArrayBuffer {
    const binary = window.atob(value);
    const buffer = new ArrayBuffer(binary.length);
    const bytes = new Uint8Array(buffer);

    for (let index = 0; index < binary.length; index++) {
        bytes[index] = binary.charCodeAt(index);
    }

    return buffer;
}

function saveFile(configFile: AppConfigFile) {
    const content = configFile.base64 ? base64ToArrayBuffer(configFile.content) : configFile.content;
    const blob = new Blob([content]);
    const url = window.URL.createObjectURL(blob);
    const anchor = document.createElement("a");

    anchor.href = url;
    anchor.download = configFile.name;
    anchor.click();
    window.URL.revokeObjectURL(url);
}

function View({ projectId, appId, configFile }: Props) {
    const [open, setOpen] = useState(false);
    const [isDownloading, setIsDownloading] = useState(false);
    const { actions: configFileDialogActions } = useCreateOrEditAppConfigFileDialog();
    const { queries } = useAppConfigFilesApi();

    const { mutate: deleteOne, isPending: isDeleting } = AppConfigFilesCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("App config file deleted successfully");
            setOpen(false);
        },
    });

    async function handleDownload() {
        try {
            setIsDownloading(true);
            const { data } = await queries.findOneById({
                projectID: projectId,
                appID: appId,
                configFileID: configFile.id,
            });
            saveFile(data);
            setOpen(false);
        } catch (error) {
            toast.error(error instanceof Error ? error.message : "Failed to download config file");
        } finally {
            setIsDownloading(false);
        }
    }

    return (
        <DropdownMenu
            open={open}
            onOpenChange={setOpen}
        >
            <DropdownMenuTrigger
                asChild
                className="h-8 w-8"
            >
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                >
                    <MoreVertical className="size-4" />
                    <span className="sr-only">Actions menu</span>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <div className="flex flex-col gap-0">
                    <Button
                        className="justify-start py-1.5"
                        variant="ghost"
                        onClick={() => {
                            configFileDialogActions.openEdit(projectId, appId, configFile);
                            setOpen(false);
                        }}
                    >
                        <Edit2Icon className="mr-2 size-4" />
                        Edit
                    </Button>
                    <Button
                        className="justify-start py-1.5"
                        variant="ghost"
                        disabled={isDownloading}
                        onClick={() => {
                            void handleDownload();
                        }}
                    >
                        <DownloadIcon className="mr-2 size-4" />
                        Download File
                    </Button>
                    <PopConfirm
                        title="Delete Item"
                        variant="destructive"
                        confirmText="Delete"
                        cancelText="Cancel"
                        description="Confirm deletion of this item?"
                        onConfirm={() => {
                            deleteOne({ projectID: projectId, appID: appId, configFileID: configFile.id });
                        }}
                    >
                        <Button
                            className="justify-start py-1.5"
                            variant="ghost"
                            disabled={isDeleting}
                        >
                            <Trash2Icon className="mr-2 size-4" />
                            Remove
                        </Button>
                    </PopConfirm>
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    projectId: string;
    appId: string;
    configFile: AppConfigFile;
}

export const MenuCell = React.memo(View);
