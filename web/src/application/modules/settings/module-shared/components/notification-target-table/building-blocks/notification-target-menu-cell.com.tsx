import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectNotificationCommands } from "~/projects/data/commands";
import { NotificationCommands } from "~/settings/data/commands";
import { useUpdateNotificationTargetStatusDialog } from "~/settings/dialogs/update-notification-target-status";
import type { SettingNotification } from "~/settings/domain";
import { isInheritedProjectSetting, useInheritedSettingAlert } from "~/settings/module-shared/hooks";

import { PopConfirm } from "@application/shared/components";

import type { NotificationTargetTableScope } from "../notification-target-table.types";

function View({ scope, notificationTarget }: Props) {
    const [open, setOpen] = useState(false);
    const updateStatusDialog = useUpdateNotificationTargetStatusDialog();
    const inheritedSettingAlert = useInheritedSettingAlert();

    const { mutate: deleteSettingNotification, isPending: isDeletingSetting } = NotificationCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Notification target deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectNotification, isPending: isDeletingProject } =
        ProjectNotificationCommands.useDeleteOne({
            onSuccess: () => {
                toast.success("Project notification target deleted successfully");
                setOpen(false);
            },
        });

    const isDeleting = isDeletingSetting || isDeletingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, notificationTarget.inherited);

    function handleDelete() {
        if (isInheritedProject) {
            inheritedSettingAlert.open();
            setOpen(false);
            return;
        }

        if (scope.type === "project") {
            deleteProjectNotification({ projectID: scope.projectId, id: notificationTarget.id });
            return;
        }

        deleteSettingNotification({ id: notificationTarget.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            inheritedSettingAlert.open();
            setOpen(false);
            return;
        }

        updateStatusDialog.actions.open(scope, notificationTarget.id);
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
                    <Button
                        className="justify-start py-1.5"
                        variant="ghost"
                        onClick={handleChangeStatus}
                    >
                        <SlidersHorizontal className="mr-2 size-4" />
                        Change Status
                    </Button>
                    {isInheritedProject ? (
                        <Button
                            className="justify-start py-1.5"
                            variant="ghost"
                            onClick={handleDelete}
                        >
                            <Trash2Icon className="mr-2 size-4" />
                            Remove
                        </Button>
                    ) : (
                        <PopConfirm
                            title="Delete notification target"
                            variant="destructive"
                            confirmText="Delete"
                            cancelText="Cancel"
                            description="Confirm deletion of this item?"
                            onConfirm={handleDelete}
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
                    )}
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    scope: NotificationTargetTableScope;
    notificationTarget: SettingNotification;
}

export const NotificationTargetMenuCell = memo(View);
