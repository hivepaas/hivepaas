import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { Cog, MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { AppScheduledJobsCommands } from "~/projects/data/commands";
import { useRunNowTaskCreatedDialog } from "~/projects/dialogs/run-now-task-created";
import { useUpdateAppScheduledJobStatusDialog } from "~/projects/dialogs/update-app-scheduled-job-status";
import type { AppScheduledJob } from "~/projects/domain";

import { PopConfirm } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

function View({ projectId, appId, scheduledJob }: Props) {
    const [open, setOpen] = useState(false);
    const updateStatusDialog = useUpdateAppScheduledJobStatusDialog();
    const runNowTaskCreatedDialog = useRunNowTaskCreatedDialog();
    const { canWrite, canDelete } = useConditionalModule({ id: MODULE_IDS.Project });

    const { mutate: runNow, isPending: isRunning } = AppScheduledJobsCommands.useRunNow({
        onSuccess: response => {
            setOpen(false);
            runNowTaskCreatedDialog.actions.open(response.data.task.id);
        },
    });

    const { mutate: deleteOne, isPending: isDeleting } = AppScheduledJobsCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("App scheduled job deleted successfully");
            setOpen(false);
        },
    });

    function handleRunNow() {
        if (!canWrite) {
            return;
        }

        runNow({ projectID: projectId, appID: appId, scheduledJobID: scheduledJob.id });
    }

    function handleChangeStatus() {
        updateStatusDialog.actions.open(projectId, appId, scheduledJob.id);
        setOpen(false);
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
                    {canWrite ? (
                        <Button
                            className="justify-start py-1.5"
                            variant="ghost"
                            onClick={handleRunNow}
                            disabled={isRunning}
                        >
                            <Cog className="mr-2 size-4" />
                            Run Now
                        </Button>
                    ) : (
                        <PermissionTooltipAction
                            id={MODULE_IDS.Project}
                            action="write"
                            triggerClassName="w-full"
                        >
                            {({ isDenied }) => (
                                <Button
                                    className="justify-start py-1.5 w-full"
                                    variant="ghost"
                                    disabled={isDenied}
                                >
                                    <Cog className="mr-2 size-4" />
                                    Run Now
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    )}

                    {canWrite ? (
                        <Button
                            className="justify-start py-1.5"
                            variant="ghost"
                            onClick={handleChangeStatus}
                        >
                            <SlidersHorizontal className="mr-2 size-4" />
                            Change Status
                        </Button>
                    ) : (
                        <PermissionTooltipAction
                            id={MODULE_IDS.Project}
                            action="write"
                            triggerClassName="w-full"
                        >
                            {({ isDenied }) => (
                                <Button
                                    className="justify-start py-1.5 w-full"
                                    variant="ghost"
                                    disabled={isDenied}
                                >
                                    <SlidersHorizontal className="mr-2 size-4" />
                                    Change Status
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    )}

                    {canDelete ? (
                        <PopConfirm
                            title="Remove Scheduled Job"
                            variant="destructive"
                            confirmText="Remove"
                            cancelText="Cancel"
                            description="Are you sure you want to remove this scheduled job?"
                            onConfirm={() => {
                                deleteOne({ projectID: projectId, appID: appId, scheduledJobID: scheduledJob.id });
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
                    ) : (
                        <PermissionTooltipAction
                            id={MODULE_IDS.Project}
                            action="delete"
                            triggerClassName="w-full"
                        >
                            {({ isDenied }) => (
                                <Button
                                    className="justify-start py-1.5 w-full"
                                    variant="ghost"
                                    disabled={isDenied}
                                >
                                    <Trash2Icon className="mr-2 size-4" />
                                    Remove
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    )}
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    projectId: string;
    appId: string;
    scheduledJob: AppScheduledJob;
}

export const MenuCell = memo(View);
