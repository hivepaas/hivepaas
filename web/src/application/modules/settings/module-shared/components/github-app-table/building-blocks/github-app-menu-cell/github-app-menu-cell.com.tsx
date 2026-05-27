import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectGithubAppCommands } from "~/projects/data/commands";
import { GithubAppCommands } from "~/settings/data/commands";
import { useUpdateGithubAppStatusDialog } from "~/settings/dialogs/update-github-app-status";
import type { SettingGithubApp } from "~/settings/domain";
import { isInheritedProjectSetting, useInheritedSettingAlert } from "~/settings/module-shared/hooks";

import { PopConfirm } from "@application/shared/components";

import type { GithubAppTableScope } from "../../github-app-table.types";

function View({ scope, githubApp }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateGithubAppStatusDialog();
    const inheritedSettingAlert = useInheritedSettingAlert();

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
        if (isInheritedProject) {
            inheritedSettingAlert.open({ entityTitle: "Github App" });
            setOpen(false);
            return;
        }

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
                            title="Delete github app"
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
    scope: GithubAppTableScope;
    githubApp: SettingGithubApp;
}

export const GithubAppMenuCell = memo(View);
