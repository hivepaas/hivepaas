import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";

import { Button } from "@/components/ui/button";

import { useRunNowTaskCreatedDialogState } from "../hooks";

export function RunNowTaskCreatedDialog() {
    const { state, props: dialogOptions, ...actions } = useRunNowTaskCreatedDialogState();
    const open = state.mode === "open";

    function handleClose() {
        actions.close();
        dialogOptions?.onClose?.();
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
            <DialogContent className="min-w-[390px] w-[680px]">
                <DialogHeader>
                    <DialogTitle>Task created</DialogTitle>
                </DialogHeader>

                <div className={cn(dashedBorderBox, "text-center text-base leading-7")}>
                    <p>The task has been added to the execution queue.</p>
                    <p>
                        You can view the execution details via the job&apos;s{" "}
                        <span className="text-link">View Tasks</span> link or click the button below.
                    </p>
                </div>

                <div className="flex justify-end gap-4">
                    <Button
                        type="button"
                        disabled
                    >
                        View Tasks
                    </Button>
                    <Button
                        type="button"
                        onClick={handleClose}
                    >
                        Close
                    </Button>
                </div>
            </DialogContent>
        </Dialog>
    );
}
