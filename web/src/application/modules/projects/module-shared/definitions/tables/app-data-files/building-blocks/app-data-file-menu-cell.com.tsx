import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { Checkbox } from "@components/ui/checkbox";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { DownloadIcon, MoreVertical, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { useAppDataFilesApi } from "~/projects/api/hooks/project-apps";
import { AppDataFilesCommands } from "~/projects/data/commands";
import type { AppDataFile } from "~/projects/domain";

import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction } from "@application/shared/permissions";

function View({ projectId, appId, dataFile }: Props) {
    const [menuOpen, setMenuOpen] = useState(false);
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [deletePermanently, setDeletePermanently] = useState(false);
    const [isDownloading, setIsDownloading] = useState(false);
    const { queries } = useAppDataFilesApi();
    const { mutate: deleteOne, isPending: isDeleting } = AppDataFilesCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("App data file deleted successfully");
            setDeleteDialogOpen(false);
        },
    });

    const isLocalFile = dataFile.storageType === "local";

    async function handleDownload() {
        try {
            setIsDownloading(true);
            const { data } = await queries.getDownloadUrl({
                projectID: projectId,
                appID: appId,
                dataFileID: dataFile.id,
            });

            window.open(data.url, "_blank", "noopener,noreferrer");
            setMenuOpen(false);
        } catch (error) {
            toast.error(error instanceof Error ? error.message : "Failed to download app data file");
        } finally {
            setIsDownloading(false);
        }
    }

    function openDeleteDialog() {
        setDeletePermanently(false);
        setDeleteDialogOpen(true);
        setMenuOpen(false);
    }

    function closeDeleteDialog() {
        if (!isDeleting) {
            setDeleteDialogOpen(false);
        }
    }

    function handleDelete() {
        deleteOne({
            projectID: projectId,
            appID: appId,
            dataFileID: dataFile.id,
            deletePermanently: isLocalFile || deletePermanently,
        });
    }

    return (
        <>
            <DropdownMenu
                open={menuOpen}
                onOpenChange={setMenuOpen}
            >
                <DropdownMenuTrigger asChild>
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
                            Download
                        </Button>
                        <PermissionTooltipAction
                            id={MODULE_IDS.Project}
                            action="write"
                            triggerClassName="w-full"
                        >
                            {({ isDenied }) => (
                                <Button
                                    className="w-full justify-start py-1.5"
                                    variant="ghost"
                                    disabled={isDenied}
                                    onClick={openDeleteDialog}
                                >
                                    <Trash2Icon className="mr-2 size-4" />
                                    Delete
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    </div>
                </DropdownMenuContent>
            </DropdownMenu>

            <Dialog
                open={deleteDialogOpen}
                onOpenChange={isOpen => {
                    if (!isOpen) {
                        closeDeleteDialog();
                    }
                }}
            >
                <DialogFixedContent className="w-[560px] max-w-[calc(100vw-2rem)]">
                    <DialogHeader>
                        <DialogTitle>Delete file: {dataFile.name}</DialogTitle>
                    </DialogHeader>
                    <DialogBody className="flex flex-col gap-4">
                        {isLocalFile ? (
                            <p className="rounded-md border border-dashed border-primary/60 px-4 py-3 text-sm">
                                <span className="font-medium text-destructive">Warning:</span> This will delete the
                                database record and permanently erase the physical file from the storage system. Once
                                deleted, this file cannot be retrieved or restored.
                            </p>
                        ) : (
                            <>
                                <p className="rounded-md border border-dashed border-primary/60 px-4 py-3 text-sm">
                                    If you select{" "}
                                    <span className="font-medium text-destructive">Delete Physical File</span>, the file
                                    will be permanently deleted and cannot be recovered. Otherwise, only the file record
                                    in the database will be removed.
                                </p>
                                <label
                                    htmlFor={`deletePhysicalFile-${dataFile.id}`}
                                    className="flex items-center gap-3 text-sm font-medium"
                                >
                                    <Checkbox
                                        id={`deletePhysicalFile-${dataFile.id}`}
                                        checked={deletePermanently}
                                        disabled={isDeleting}
                                        onCheckedChange={checked => {
                                            setDeletePermanently(checked === true);
                                        }}
                                    />
                                    Delete Physical File
                                </label>
                            </>
                        )}
                    </DialogBody>
                    <DialogActionFooter className="gap-3">
                        <Button
                            variant="outline"
                            disabled={isDeleting}
                            onClick={closeDeleteDialog}
                        >
                            Cancel
                        </Button>
                        <Button
                            variant="destructive"
                            disabled={isDeleting}
                            onClick={handleDelete}
                        >
                            Delete
                        </Button>
                    </DialogActionFooter>
                </DialogFixedContent>
            </Dialog>
        </>
    );
}

interface Props {
    projectId: string;
    appId: string;
    dataFile: AppDataFile;
}

export const AppDataFileMenuCell = memo(View);
