import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectKeyAuthCommands } from "~/projects/data/commands";
import { KeyAuthCommands } from "~/settings/data/commands";
import { useUpdateKeyAuthStatusDialog } from "~/settings/dialogs/update-key-auth-status";
import type { SettingKeyAuth } from "~/settings/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import { CopyIdMenuButton } from "@application/shared/components";

import type { KeyAuthTableScope } from "../../key-auth-table.types";

function View({ scope, keyAuth }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateKeyAuthStatusDialog();
    const { mutate: deleteSettingKeyAuth, isPending: isDeletingSetting } = KeyAuthCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Key auth deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectKeyAuth, isPending: isDeletingProject } = ProjectKeyAuthCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project key auth deleted successfully");
            setOpen(false);
        },
    });

    const isDeleting = isDeletingSetting || isDeletingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, keyAuth.inherited);

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectKeyAuth({
                projectID: scope.projectId,
                env: scope.env,
                id: keyAuth.id,
            });
            return;
        }

        deleteSettingKeyAuth({ id: keyAuth.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            updateStatusDialog.actions.open(scope, keyAuth.id, {
                props: {
                    readOnlyInherited: true,
                    entityTitle: SETTINGS_ENTITY_TITLES.keyAuth,
                },
            });
            setOpen(false);
            return;
        }

        updateStatusDialog.actions.open(scope, keyAuth.id);
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
                    <CopyIdMenuButton
                        id={keyAuth.id}
                        onCopied={() => {
                            setOpen(false);
                        }}
                    />
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
                        title="Delete key auth"
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
    scope: KeyAuthTableScope;
    keyAuth: SettingKeyAuth;
}

export const KeyAuthMenuCell = memo(View);
