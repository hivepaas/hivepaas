import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectSSHKeyCommands } from "~/projects/data/commands";
import { SSHKeyCommands } from "~/settings/data/commands";
import { useUpdateSSHKeyStatusDialog } from "~/settings/dialogs/update-ssh-key-status";
import type { SettingSSHKey } from "~/settings/domain";

import { PopConfirm } from "@application/shared/components";

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

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectSSHKey({ projectID: scope.projectId, id: sshKey.id });
            return;
        }

        deleteSettingSSHKey({ id: sshKey.id });
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
                        onClick={() => {
                            updateStatusDialog.actions.open(scope, sshKey.id);
                            setOpen(false);
                        }}
                    >
                        <SlidersHorizontal className="mr-2 size-4" />
                        Change Status
                    </Button>
                    <PopConfirm
                        title="Delete SSH key"
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
