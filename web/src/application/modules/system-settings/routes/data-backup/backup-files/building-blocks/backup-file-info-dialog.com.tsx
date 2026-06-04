import { format } from "date-fns";
import { SystemBackupFileQueries } from "~/system-settings/data";
import type { SystemBackupFile } from "~/system-settings/domain";
import { getBackupFileStorageLabel } from "~/system-settings/module-shared/definitions/tables/system-backup-files";
import { ESystemBackupFileStorageType } from "~/system-settings/module-shared/enums";

import { AppLoader } from "@application/shared/components";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";

function InfoRow({ label, value }: InfoRowProps) {
    return (
        <div className="grid grid-cols-[120px_1fr] gap-6 text-sm">
            <div className="font-semibold text-foreground">{label}</div>
            <div className="text-foreground break-words">{value ?? "-"}</div>
        </div>
    );
}

function BackupFileInfoContent({ file }: { file: SystemBackupFile }) {
    const isCloudFile = file.storageType === ESystemBackupFileStorageType.Cloud;

    return (
        <div className="flex flex-col gap-8 pt-2">
            <InfoRow
                label="Name"
                value={file.name}
            />
            <InfoRow
                label="Mimetype"
                value={file.mimetype}
            />
            <InfoRow
                label="Size"
                value={file.sizeStr}
            />
            <InfoRow
                label="Storage"
                value={getBackupFileStorageLabel(file)}
            />
            {isCloudFile && (
                <>
                    <InfoRow
                        label="Bucket"
                        value={file.bucket}
                    />
                    <InfoRow
                        label="Directory Path"
                        value={file.path}
                    />
                </>
            )}
            <InfoRow
                label="Created At"
                value={format(file.createdAt, "yyyy-MM-dd HH:mm:ss")}
            />
        </div>
    );
}

export function BackupFileInfoDialog({ fileID, open, onOpenChange }: Props) {
    const { data, isLoading } = SystemBackupFileQueries.useFindOneById(
        { fileID: fileID ?? "" },
        { enabled: open && Boolean(fileID) },
    );

    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
        >
            <DialogContent className="w-[640px] max-w-[calc(100vw-2rem)]">
                <DialogHeader>
                    <DialogTitle>Backup file info</DialogTitle>
                </DialogHeader>

                {isLoading ? <AppLoader /> : data?.data ? <BackupFileInfoContent file={data.data} /> : null}
            </DialogContent>
        </Dialog>
    );
}

interface InfoRowProps {
    label: string;
    value?: string;
}

interface Props {
    fileID?: string;
    open: boolean;
    onOpenChange: (open: boolean) => void;
}
