import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectBackupRepoCommands } from "~/projects/data/commands";
import { BackupRepoCommands } from "~/settings/data/commands";
import { useUpdateBackupRepoStatusDialog } from "~/settings/dialogs/update-backup-repo-status";
import type { SettingBackupRepo } from "~/settings/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { BackupRepoTableScope } from "../../backup-repo-table.types";

function View({ scope, backupRepo }: Props) {
    const [open, setOpen] = useState(false);
    const updateStatusDialog = useUpdateBackupRepoStatusDialog();

    const { mutate: deleteSettingBackupRepo, isPending: isDeletingSetting } = BackupRepoCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Backup repository deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectBackupRepo, isPending: isDeletingProject } = ProjectBackupRepoCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project backup repository deleted successfully");
            setOpen(false);
        },
    });

    const isDeleting = isDeletingSetting || isDeletingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, backupRepo.inherited);

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectBackupRepo({
                projectID: scope.projectId,
                env: scope.env,
                id: backupRepo.id,
            });
            return;
        }

        deleteSettingBackupRepo({ id: backupRepo.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            updateStatusDialog.actions.open(scope, backupRepo.id, {
                props: {
                    readOnlyInherited: true,
                    entityTitle: SETTINGS_ENTITY_TITLES.backupRepo,
                },
            });
            setOpen(false);
            return;
        }

        updateStatusDialog.actions.open(scope, backupRepo.id);
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
                        title="Delete backup repository"
                        confirmText="Delete"
                        cancelText="Cancel"
                        description="Confirm deletion of this item?"
                        onConfirm={handleDelete}
                        isLoading={isDeleting}
                    >
                        <Trash2Icon className="mr-2 size-4" />
                        Remove
                    </SettingsScopePopConfirmButton>
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    scope: BackupRepoTableScope;
    backupRepo: SettingBackupRepo;
}

export const BackupRepoMenuCell = memo(View);
