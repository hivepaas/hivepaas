import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectImServiceCommands } from "~/projects/data/commands";
import { ImServiceCommands } from "~/settings/data/commands";
import { useUpdateImPlatformStatusDialog } from "~/settings/dialogs/update-im-platform-status";
import type { SettingImService } from "~/settings/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { ImPlatformTableScope } from "../../im-platform-table.types";

function View({ scope, imPlatform }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateImPlatformStatusDialog();
    const { mutate: deleteSettingImPlatform, isPending: isDeletingSetting } = ImServiceCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("IM platform deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectImPlatform, isPending: isDeletingProject } = ProjectImServiceCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project IM platform deleted successfully");
            setOpen(false);
        },
    });

    const isDeleting = isDeletingSetting || isDeletingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, imPlatform.inherited);

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectImPlatform({
                projectID: scope.projectId,
                id: imPlatform.id,
            });
            return;
        }

        deleteSettingImPlatform({ id: imPlatform.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            updateStatusDialog.actions.open(scope, imPlatform.id, {
                props: {
                    readOnlyInherited: true,
                    entityTitle: SETTINGS_ENTITY_TITLES.imPlatform,
                },
            });
            setOpen(false);
            return;
        }

        updateStatusDialog.actions.open(scope, imPlatform.id);
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
                        title="Delete IM platform"
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
    scope: ImPlatformTableScope;
    imPlatform: SettingImService;
}

export const ImPlatformMenuCell = memo(View);
