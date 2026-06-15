import { Button } from "@components/ui";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import type { ProjectImageBuildRepoCacheClearResult } from "~/projects/domain";

import { getFriendlyDataSize } from "@application/shared/utils/data-size";

interface ClearRepoCacheResultDialogProps {
    open: boolean;
    result?: ProjectImageBuildRepoCacheClearResult | null;
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
            <DialogFixedContent className="w-[640px] max-w-[calc(100vw-2rem)]">
                <DialogHeader>
                    <DialogTitle className="text-2xl">Repo cache cleared</DialogTitle>
                </DialogHeader>

                <DialogBody className="py-8">
                    <div className="grid grid-cols-[minmax(0,1fr)_auto] gap-x-8 gap-y-7 text-base sm:text-lg">
                        <span className="font-semibold">Files deleted</span>
                        <span>{result?.filesDeleted ?? 0}</span>

                        <span className="font-semibold">Space Reclaimed</span>
                        <span>{formatSpaceReclaimed(result?.spaceReclaimed)}</span>
                    </div>
                </DialogBody>

                <DialogActionFooter className="flex justify-end">
                    <Button
                        type="button"
                        className="min-w-[120px]"
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
