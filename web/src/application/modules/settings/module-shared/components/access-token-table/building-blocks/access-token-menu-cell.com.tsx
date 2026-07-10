import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectAccessTokenCommands } from "~/projects/data/commands";
import { AccessTokenCommands } from "~/settings/data/commands";
import { useUpdateAccessTokenStatusDialog } from "~/settings/dialogs/update-access-token-status";
import type { SettingAccessToken } from "~/settings/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { AccessTokenTableScope } from "../access-token-table.types";

function View({ scope, accessToken }: Props) {
    const [open, setOpen] = useState(false);
    const updateStatusDialog = useUpdateAccessTokenStatusDialog();
    const { mutate: deleteSettingAccessToken, isPending: isDeletingSetting } = AccessTokenCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Access token deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectAccessToken, isPending: isDeletingProject } = ProjectAccessTokenCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project access token deleted successfully");
            setOpen(false);
        },
    });

    const isDeleting = isDeletingSetting || isDeletingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, accessToken.inherited);

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectAccessToken({ projectID: scope.projectId, id: accessToken.id });
            return;
        }

        deleteSettingAccessToken({ id: accessToken.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            updateStatusDialog.actions.open(scope, accessToken.id, {
                props: {
                    readOnlyInherited: true,
                    entityTitle: SETTINGS_ENTITY_TITLES.accessToken,
                },
            });
            setOpen(false);
            return;
        }

        updateStatusDialog.actions.open(scope, accessToken.id);
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
                        title="Delete access token"
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
    scope: AccessTokenTableScope;
    accessToken: SettingAccessToken;
}

export const AccessTokenMenuCell = memo(View);
