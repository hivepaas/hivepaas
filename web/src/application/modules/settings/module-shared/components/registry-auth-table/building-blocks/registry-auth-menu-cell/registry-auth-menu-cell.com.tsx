import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectRegistryAuthCommands } from "~/projects/data/commands";
import { RegistryAuthCommands } from "~/settings/data/commands";
import { useUpdateRegistryAuthStatusDialog } from "~/settings/dialogs/update-registry-auth-status";
import type { SettingRegistryAuth } from "~/settings/domain";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting, useInheritedSettingAlert } from "~/settings/module-shared/hooks";

import { PopConfirm } from "@application/shared/components";

import type { RegistryAuthTableScope } from "../../registry-auth-table.types";

function View({ scope, registryAuth }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateRegistryAuthStatusDialog();
    const inheritedSettingAlert = useInheritedSettingAlert();

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
        if (isInheritedProject) {
            inheritedSettingAlert.open({ entityTitle: SETTINGS_ENTITY_TITLES.registryAuth });
            setOpen(false);
            return;
        }

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
                            title="Delete registry auth"
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
    scope: RegistryAuthTableScope;
    registryAuth: SettingRegistryAuth;
}

export const RegistryAuthMenuCell = memo(View);
