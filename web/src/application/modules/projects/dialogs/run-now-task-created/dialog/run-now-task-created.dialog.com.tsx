import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import { Button } from "@/components/ui/button";

import { useRunNowTaskCreatedDialogState } from "../hooks";

export function RunNowTaskCreatedDialog() {
    const { state, props: dialogOptions, ...actions } = useRunNowTaskCreatedDialogState();
    const { navigate } = useAppNavigate();
    const open = state.mode === "open";

    function handleClose() {
        actions.close();
        dialogOptions?.onClose?.();
    }

    function handleViewTask() {
        if (state.mode !== "open") {
            return;
        }

        navigate.modules(
            ROUTE.projects.single.apps.single.scheduledJobTasks.details.$route(
                state.projectId,
                state.appId,
                state.scheduledJobId,
                state.taskId,
            ),
        );
        handleClose();
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
            <DialogFixedContent className="min-w-[390px] w-[680px]">
                <DialogHeader>
                    <DialogTitle>Task created</DialogTitle>
                </DialogHeader>

                <DialogBody>
                    <div className={cn(dashedBorderBox, "text-center text-base leading-7")}>
                        <p>The task has been added to the execution queue.</p>
                        <p>
                            You can view the execution details via the job&apos;s{" "}
                            <span className="text-link">View Tasks</span> link or click the button below.
                        </p>
                    </div>
                </DialogBody>

                <DialogActionFooter className="flex justify-end gap-4">
                    <Button
                        type="button"
                        onClick={handleViewTask}
                    >
                        View Task
                    </Button>
                    <Button
                        type="button"
                        onClick={handleClose}
                        className="min-w-[100px]"
                    >
                        Close
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
