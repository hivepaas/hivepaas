import { Button } from "@components/ui";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import type { ImageBuildRepoCacheClearResult } from "~/settings/domain";

import { getFriendlyDataSize } from "@application/shared/utils/data-size";

interface ClearRepoCacheResultDialogProps {
    open: boolean;
    result?: ImageBuildRepoCacheClearResult | null;
    onOpenChange: (open: boolean) => void;
}

function formatSpaceReclaimed(value?: number): string {
    return getFriendlyDataSize(value) || "0 B";
}

export function ClearRepoCacheResultDialog({ open, result, onOpenChange }: ClearRepoCacheResultDialogProps) {
    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
        >
            <DialogFixedContent className="w-[400px] max-w-[calc(100vw-2rem)]">
                <DialogHeader>
                    <DialogTitle>Repo cache cleared</DialogTitle>
                </DialogHeader>

                <DialogBody className="">
                    <div className="flex flex-col gap-2">
                        <div className="flex items-center gap-3">
                            <div className="text-sm font-medium w-[145px]">Files deleted</div>
                            <div className="flex-1">{result?.filesDeleted ?? 0}</div>
                        </div>
                        <div className="flex items-center gap-3">
                            <div className="text-sm font-medium w-[145px]">Space Reclaimed</div>
                            <div className="flex-1">{formatSpaceReclaimed(result?.spaceReclaimed)}</div>
                        </div>
                    </div>
                </DialogBody>

                <DialogActionFooter className="flex justify-end">
                    <Button
                        type="button"
                        className="min-w-[100px]"
                        onClick={() => {
                            onOpenChange(false);
                        }}
                    >
                        Close
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
