import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { DownloadIcon, MoreVertical, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { useAppConfigFilesApi } from "~/projects/api/hooks/project-apps";
import { AppConfigFilesCommands } from "~/projects/data/commands";
import type { AppConfigFile } from "~/projects/domain";

import { PopConfirm } from "@application/shared/components";

function View({ projectId, appId, configFile }: Props) {
    const [open, setOpen] = useState(false);
    const [isDownloading, setIsDownloading] = useState(false);
    const { queries, helpers } = useAppConfigFilesApi();

    const { mutate: deleteOne, isPending: isDeleting } = AppConfigFilesCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("App config file deleted successfully");
            setOpen(false);
        },
    });

    async function handleDownload() {
        try {
            setIsDownloading(true);
            const { data } = await queries.getDownloadToken({
                projectID: projectId,
                appID: appId,
                configFileID: configFile.id,
            });
            const downloadUrl = helpers.buildDownloadUrl({
                projectID: projectId,
                appID: appId,
                configFileID: configFile.id,
                token: data.token,
                viewInline: true,
            });

            window.open(downloadUrl, "_blank", "noopener,noreferrer");
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
