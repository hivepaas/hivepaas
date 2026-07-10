import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectGithubAppCommands } from "~/projects/data/commands";
import { GithubAppCommands } from "~/settings/data/commands";
import { useUpdateGithubAppStatusDialog } from "~/settings/dialogs/update-github-app-status";
import type { SettingGithubApp } from "~/settings/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { GithubAppTableScope } from "../../github-app-table.types";

function View({ scope, githubApp }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateGithubAppStatusDialog();
    const { mutate: deleteSettingsGithubApp, isPending: isDeletingSettings } = GithubAppCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Github app deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectGithubApp, isPending: isDeletingProject } = ProjectGithubAppCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project github app deleted successfully");
            setOpen(false);
        },
    });

    const isDeleting = isDeletingSettings || isDeletingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, githubApp.inherited);

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectGithubApp({
                projectID: scope.projectId,
                id: githubApp.id,
            });
            return;
        }

        deleteSettingsGithubApp({ id: githubApp.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            updateStatusDialog.actions.open(scope, githubApp.id, {
                props: {
                    readOnlyInherited: true,
                    entityTitle: "Github App",
                },
            });
            setOpen(false);
            return;
        }

        updateStatusDialog.actions.open(scope, githubApp.id);
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
                        title="Delete github app"
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
    scope: GithubAppTableScope;
    githubApp: SettingGithubApp;
}

export const GithubAppMenuCell = memo(View);
