import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectSSHKeyCommands } from "~/projects/data/commands";
import { SSHKeyCommands } from "~/settings/data/commands";
import { useUpdateSSHKeyStatusDialog } from "~/settings/dialogs/update-ssh-key-status";
import type { SettingSSHKey } from "~/settings/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { SSHKeyTableScope } from "../ssh-key-table.types";

function View({ scope, sshKey }: Props) {
    const [open, setOpen] = useState(false);
    const updateStatusDialog = useUpdateSSHKeyStatusDialog();
    const { mutate: deleteSettingSSHKey, isPending: isDeletingSetting } = SSHKeyCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("SSH key deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectSSHKey, isPending: isDeletingProject } = ProjectSSHKeyCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project SSH key deleted successfully");
            setOpen(false);
        },
    });

    const isDeleting = isDeletingSetting || isDeletingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, sshKey.inherited);

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectSSHKey({ projectID: scope.projectId, id: sshKey.id });
            return;
        }

        deleteSettingSSHKey({ id: sshKey.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            updateStatusDialog.actions.open(scope, sshKey.id, {
                props: {
                    readOnlyInherited: true,
                    entityTitle: SETTINGS_ENTITY_TITLES.sshKey,
                },
            });
            setOpen(false);
            return;
        }

        updateStatusDialog.actions.open(scope, sshKey.id);
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
                        title="Delete SSH key"
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
    scope: SSHKeyTableScope;
    sshKey: SettingSSHKey;
}

export const SSHKeyMenuCell = memo(View);
