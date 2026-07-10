import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectAcmeDnsProviderCommands } from "~/projects/data/commands";
import { AcmeDnsProviderCommands } from "~/settings/data/commands";
import { useUpdateAcmeDnsProviderStatusDialog } from "~/settings/dialogs/update-acme-dns-provider-status";
import type { SettingAcmeDnsProvider } from "~/settings/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { AcmeDnsProviderTableScope } from "../../acme-dns-provider-table.types";

function View({ scope, acmeDnsProvider }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateAcmeDnsProviderStatusDialog();
    const { mutate: deleteSettingAcmeDnsProvider, isPending: isDeletingSetting } = AcmeDnsProviderCommands.useDeleteOne(
        {
            onSuccess: () => {
                toast.success("ACME DNS provider deleted successfully");
                setOpen(false);
            },
        },
    );

    const { mutate: deleteProjectAcmeDnsProvider, isPending: isDeletingProject } =
        ProjectAcmeDnsProviderCommands.useDeleteOne({
            onSuccess: () => {
                toast.success("Project ACME DNS provider deleted successfully");
                setOpen(false);
            },
        });

    const isDeleting = isDeletingSetting || isDeletingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, acmeDnsProvider.inherited);

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectAcmeDnsProvider({
                projectID: scope.projectId,
                id: acmeDnsProvider.id,
            });
            return;
        }

        deleteSettingAcmeDnsProvider({ id: acmeDnsProvider.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            updateStatusDialog.actions.open(scope, acmeDnsProvider.id, {
                props: {
                    readOnlyInherited: true,
                    entityTitle: SETTINGS_ENTITY_TITLES.acmeDnsProvider,
                },
            });
            setOpen(false);
            return;
        }

        updateStatusDialog.actions.open(scope, acmeDnsProvider.id);
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
                    {!isInheritedProject && (
                        <SettingsScopePopConfirmButton
                            scope={scope}
                            action="delete"
                            title="Delete ACME DNS provider"
                            confirmText="Delete"
                            cancelText="Cancel"
                            description="Confirm deletion of this item?"
                            onConfirm={handleDelete}
                            isLoading={isDeleting}
                        >
                            <Trash2Icon className="mr-2 size-4" />
                            Delete
                        </SettingsScopePopConfirmButton>
                    )}
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    scope: AcmeDnsProviderTableScope;
    acmeDnsProvider: SettingAcmeDnsProvider;
}

export const AcmeDnsProviderMenuCell = memo(View);
