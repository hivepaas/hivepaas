import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectCloudStorageCommands } from "~/projects/data/commands";
import { CloudStorageCommands } from "~/settings/data/commands";
import { useUpdateCloudStorageStatusDialog } from "~/settings/dialogs/update-cloud-storage-status";
import type { SettingCloudStorage } from "~/settings/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { CloudStorageTableScope } from "../cloud-storage-table.types";

function View({ scope, cloudStorage }: Props) {
    const [open, setOpen] = useState(false);
    const updateStatusDialog = useUpdateCloudStorageStatusDialog();
    const { mutate: deleteSettingCloudStorage, isPending: isDeletingSetting } = CloudStorageCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Cloud storage deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectCloudStorage, isPending: isDeletingProject } =
        ProjectCloudStorageCommands.useDeleteOne({
            onSuccess: () => {
                toast.success("Project cloud storage deleted successfully");
                setOpen(false);
            },
        });

    const isDeleting = isDeletingSetting || isDeletingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, cloudStorage.inherited);

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectCloudStorage({ projectID: scope.projectId, id: cloudStorage.id });
            return;
        }

        deleteSettingCloudStorage({ id: cloudStorage.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            updateStatusDialog.actions.open(scope, cloudStorage.id, {
                props: {
                    readOnlyInherited: true,
                    entityTitle: SETTINGS_ENTITY_TITLES.cloudStorage,
                },
            });
            setOpen(false);
            return;
        }

        updateStatusDialog.actions.open(scope, cloudStorage.id);
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
                        title="Delete cloud storage"
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
    scope: CloudStorageTableScope;
    cloudStorage: SettingCloudStorage;
}

export const CloudStorageMenuCell = memo(View);
