import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectBasicAuthCommands } from "~/projects/data/commands";
import { BasicAuthCommands } from "~/settings/data/commands";
import { useUpdateBasicAuthStatusDialog } from "~/settings/dialogs/update-basic-auth-status";
import type { SettingBasicAuth } from "~/settings/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { BasicAuthTableScope } from "../../basic-auth-table.types";

function View({ scope, basicAuth }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateBasicAuthStatusDialog();
    const { mutate: deleteSettingBasicAuth, isPending: isDeletingSetting } = BasicAuthCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Basic auth deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectBasicAuth, isPending: isDeletingProject } = ProjectBasicAuthCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project basic auth deleted successfully");
            setOpen(false);
        },
    });

    const isDeleting = isDeletingSetting || isDeletingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, basicAuth.inherited);

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectBasicAuth({
                projectID: scope.projectId,
                id: basicAuth.id,
            });
            return;
        }

        deleteSettingBasicAuth({ id: basicAuth.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            updateStatusDialog.actions.open(scope, basicAuth.id, {
                props: {
                    readOnlyInherited: true,
                    entityTitle: SETTINGS_ENTITY_TITLES.basicAuth,
                },
            });
            setOpen(false);
            return;
        }

        updateStatusDialog.actions.open(scope, basicAuth.id);
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
                        title="Delete basic auth"
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
    scope: BasicAuthTableScope;
    basicAuth: SettingBasicAuth;
}

export const BasicAuthMenuCell = memo(View);
