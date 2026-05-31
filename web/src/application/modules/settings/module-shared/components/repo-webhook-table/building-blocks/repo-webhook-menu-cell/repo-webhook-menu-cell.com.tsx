import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectRepoWebhookCommands } from "~/projects/data/commands";
import { RepoWebhookCommands } from "~/settings/data/commands";
import { useUpdateRepoWebhookStatusDialog } from "~/settings/dialogs/update-repo-webhook-status";
import type { SettingRepoWebhook } from "~/settings/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { RepoWebhookTableScope } from "../../repo-webhook-table.types";

function View({ scope, repoWebhook }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateRepoWebhookStatusDialog();
    const { mutate: deleteSettingsRepoWebhook, isPending: isDeletingSettings } = RepoWebhookCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Webhook deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectRepoWebhook, isPending: isDeletingProject } = ProjectRepoWebhookCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project webhook deleted successfully");
            setOpen(false);
        },
    });

    const isDeleting = isDeletingSettings || isDeletingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, repoWebhook.inherited);

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectRepoWebhook({
                projectID: scope.projectId,
                id: repoWebhook.id,
            });
            return;
        }

        deleteSettingsRepoWebhook({ id: repoWebhook.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            updateStatusDialog.actions.open(scope, repoWebhook.id, {
                props: {
                    readOnlyInherited: true,
                    entityTitle: "Webhook",
                },
            });
            setOpen(false);
            return;
        }

        updateStatusDialog.actions.open(scope, repoWebhook.id);
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
                        title="Delete webhook"
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
    scope: RepoWebhookTableScope;
    repoWebhook: SettingRepoWebhook;
}

export const RepoWebhookMenuCell = memo(View);
