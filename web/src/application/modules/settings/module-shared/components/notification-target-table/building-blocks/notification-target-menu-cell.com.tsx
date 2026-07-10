import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectNotificationCommands } from "~/projects/data/commands";
import { NotificationCommands } from "~/settings/data/commands";
import { useUpdateNotificationTargetStatusDialog } from "~/settings/dialogs/update-notification-target-status";
import type { SettingNotification } from "~/settings/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { NotificationTargetTableScope } from "../notification-target-table.types";

function View({ scope, notificationTarget }: Props) {
    const [open, setOpen] = useState(false);
    const updateStatusDialog = useUpdateNotificationTargetStatusDialog();
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
        if (scope.type === "project") {
            deleteProjectNotification({ projectID: scope.projectId, id: notificationTarget.id });
            return;
        }

        deleteSettingNotification({ id: notificationTarget.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            updateStatusDialog.actions.open(scope, notificationTarget.id, {
                props: {
                    readOnlyInherited: true,
                    entityTitle: SETTINGS_ENTITY_TITLES.notificationTarget,
                },
            });
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
                    <SettingsScopeMenuButton
                        scope={scope}
                        action="write"
                        onClick={handleChangeStatus}
                    >
                        <SlidersHorizontal className="mr-2 size-4" />
                        Change Status
                    </SettingsScopeMenuButton>
                    <SettingsScopePopConfirmButton
                        scope={scope}
                        action="delete"
                        title="Delete notification target"
                        confirmText="Delete"
                        cancelText="Cancel"
                        description="Confirm deletion of this item?"
                        onConfirm={handleDelete}
                        isLoading={isDeleting}
                    >
                        <Trash2Icon className="mr-2 size-4" />
                        Delete
                    </SettingsScopePopConfirmButton>
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
