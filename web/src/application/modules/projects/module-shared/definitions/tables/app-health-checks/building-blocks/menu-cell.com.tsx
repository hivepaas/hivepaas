import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { AppHealthChecksCommands } from "~/projects/data/commands";
import { useUpdateAppHealthCheckStatusDialog } from "~/projects/dialogs/update-app-health-check-status";
import type { AppHealthCheck } from "~/projects/domain";

import { PopConfirm } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

function View({ projectId, appId, healthCheck }: Props) {
    const [open, setOpen] = useState(false);
    const updateStatusDialog = useUpdateAppHealthCheckStatusDialog();
    const { canWrite, canDelete } = useConditionalModule({ id: MODULE_IDS.Project });

    const { mutate: deleteOne, isPending: isDeleting } = AppHealthChecksCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("App health check deleted successfully");
            setOpen(false);
        },
    });

    function handleChangeStatus() {
        updateStatusDialog.actions.open(projectId, appId, healthCheck.id);
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
                            title="Remove Health Check"
                            variant="destructive"
                            confirmText="Remove"
                            cancelText="Cancel"
                            description="Are you sure you want to remove this health check?"
                            onConfirm={() => {
                                deleteOne({ projectID: projectId, appID: appId, healthCheckID: healthCheck.id });
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
    healthCheck: AppHealthCheck;
}

export const MenuCell = memo(View);
