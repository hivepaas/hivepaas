import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectSslProviderCommands } from "~/projects/data/commands";
import { SslProviderCommands } from "~/settings/data/commands";
import { useUpdateSslProviderStatusDialog } from "~/settings/dialogs/update-ssl-provider-status";
import type { SettingSslProvider } from "~/settings/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { SslProviderTableScope } from "../../ssl-provider-table.types";

function View({ scope, sslProvider }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateSslProviderStatusDialog();
    const { mutate: deleteSettingSslProvider, isPending: isDeletingSetting } = SslProviderCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("SSL provider deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectSslProvider, isPending: isDeletingProject } = ProjectSslProviderCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project SSL provider deleted successfully");
            setOpen(false);
        },
    });

    const isDeleting = isDeletingSetting || isDeletingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, sslProvider.inherited);

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectSslProvider({
                projectID: scope.projectId,
                id: sslProvider.id,
            });
            return;
        }

        deleteSettingSslProvider({ id: sslProvider.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            updateStatusDialog.actions.open(scope, sslProvider.id, {
                props: {
                    readOnlyInherited: true,
                    entityTitle: SETTINGS_ENTITY_TITLES.sslProvider,
                },
            });
            setOpen(false);
            return;
        }

        updateStatusDialog.actions.open(scope, sslProvider.id);
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
                        title="Delete SSL provider"
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
    scope: SslProviderTableScope;
    sslProvider: SettingSslProvider;
}

export const SslProviderMenuCell = memo(View);
