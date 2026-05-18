import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { DownloadIcon, MoreVertical, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { useSystemBackupFileApi } from "~/system-settings/api/hooks";
import type { SystemBackupFile } from "~/system-settings/domain";

function saveBlob(blob: Blob, filename: string) {
    const url = window.URL.createObjectURL(blob);
    const anchor = document.createElement("a");

    anchor.href = url;
    anchor.download = filename;
    anchor.click();
    window.URL.revokeObjectURL(url);
}

function View({ file }: Props) {
    const [open, setOpen] = useState(false);
    const [isDownloading, setIsDownloading] = useState(false);
    const { queries } = useSystemBackupFileApi();

    async function handleDownload() {
        try {
            setIsDownloading(true);
            const { data } = await queries.downloadOne({ fileID: file.id });
            saveBlob(data.blob, data.filename ?? file.name);
            setOpen(false);
        } catch (error) {
            toast.error(error instanceof Error ? error.message : "Failed to download backup file");
        } finally {
            setIsDownloading(false);
        }
    }

    return (
        <DropdownMenu
            open={open}
            onOpenChange={setOpen}
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
                        Download File
                    </Button>
                    <Button
                        className="justify-start py-1.5"
                        variant="ghost"
                        disabled
                    >
                        <Trash2Icon className="mr-2 size-4" />
                        Remove
                    </Button>
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    file: SystemBackupFile;
}

export const BackupFileMenuCell = memo(View);
