import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectRegistryAuthCommands } from "~/projects/data/commands";
import { RegistryAuthCommands } from "~/settings/data/commands";
import { useUpdateRegistryAuthStatusDialog } from "~/settings/dialogs/update-registry-auth-status";
import type { SettingRegistryAuth } from "~/settings/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { RegistryAuthTableScope } from "../../registry-auth-table.types";

function View({ scope, registryAuth }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateRegistryAuthStatusDialog();
    const { mutate: deleteSettingRegistryAuth, isPending: isDeletingSetting } = RegistryAuthCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Registry auth deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectRegistryAuth, isPending: isDeletingProject } =
        ProjectRegistryAuthCommands.useDeleteOne({
            onSuccess: () => {
                toast.success("Project registry auth deleted successfully");
                setOpen(false);
            },
        });

    const isDeleting = isDeletingSetting || isDeletingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, registryAuth.inherited);

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectRegistryAuth({
                projectID: scope.projectId,
                id: registryAuth.id,
            });
            return;
        }

        deleteSettingRegistryAuth({ id: registryAuth.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            updateStatusDialog.actions.open(scope, registryAuth.id, {
                props: {
                    readOnlyInherited: true,
                    entityTitle: SETTINGS_ENTITY_TITLES.registryAuth,
                },
            });
            setOpen(false);
            return;
        }

        updateStatusDialog.actions.open(scope, registryAuth.id);
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
                        title="Delete registry auth"
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
    scope: RegistryAuthTableScope;
    registryAuth: SettingRegistryAuth;
}

export const RegistryAuthMenuCell = memo(View);
