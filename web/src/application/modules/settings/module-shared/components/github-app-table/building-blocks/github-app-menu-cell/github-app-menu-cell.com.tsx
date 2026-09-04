import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@components/ui/dropdown-menu";
import { AlertTriangle, ExternalLink, MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectGithubAppCommands } from "~/projects/data/commands";
import { GithubAppCommands } from "~/settings/data/commands";
import { useUpdateGithubAppStatusDialog } from "~/settings/dialogs/update-github-app-status";
import type { SettingGithubApp } from "~/settings/domain";
import { SettingsScopeMenuButton } from "~/settings/module-shared/components";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import { Separator } from "@/components/ui";

import type { GithubAppTableScope } from "../../github-app-table.types";

function View({ scope, githubApp }: Props) {
    const [open, setOpen] = useState(false);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

    const updateStatusDialog = useUpdateGithubAppStatusDialog();
    const { mutate: deleteSettingsGithubApp, isPending: isDeletingSettings } = GithubAppCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Github app deleted successfully");
            setOpen(false);
            setIsDeleteDialogOpen(false);
        },
    });

    const { mutate: deleteProjectGithubApp, isPending: isDeletingProject } = ProjectGithubAppCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project github app deleted successfully");
            setOpen(false);
            setIsDeleteDialogOpen(false);
        },
    });

    const isDeleting = isDeletingSettings || isDeletingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, githubApp.inherited);

    const settingsUrl =
        githubApp.settingsURL ?? (githubApp.appId ? `https://github.com/settings/apps/${githubApp.appId}` : "");
    const installationsUrl = settingsUrl ? `${settingsUrl}/installations` : "";
    const advancedSettingsUrl = settingsUrl ? `${settingsUrl}/advanced` : "";

    function handleOpenSettings() {
        if (!settingsUrl) return;
        window.open(settingsUrl, "_blank", "noopener,noreferrer");
        setOpen(false);
    }

    function handleOpenInstallations() {
        if (!installationsUrl) return;
        window.open(installationsUrl, "_blank", "noopener,noreferrer");
        setOpen(false);
    }

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
        <>
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
                            action="read"
                            disabled={!settingsUrl}
                            onClick={handleOpenSettings}
                        >
                            <ExternalLink className="mr-2 size-4" />
                            App Settings
                        </SettingsScopeMenuButton>
                        <SettingsScopeMenuButton
                            scope={scope}
                            action="read"
                            disabled={!installationsUrl}
                            onClick={handleOpenInstallations}
                        >
                            <ExternalLink className="mr-2 size-4" />
                            App Installations
                        </SettingsScopeMenuButton>
                        <DropdownMenuSeparator />
                        <SettingsScopeMenuButton
                            scope={scope}
                            action="write"
                            onClick={handleChangeStatus}
                        >
                            <SlidersHorizontal className="mr-2 size-4" />
                            Change Status
                        </SettingsScopeMenuButton>
                        <SettingsScopeMenuButton
                            scope={scope}
                            action="delete"
                            onClick={() => {
                                setIsDeleteDialogOpen(true);
                                setOpen(false);
                            }}
                        >
                            <Trash2Icon className="mr-2 size-4" />
                            Delete
                        </SettingsScopeMenuButton>
                    </div>
                </DropdownMenuContent>
            </DropdownMenu>

            <Dialog
                open={isDeleteDialogOpen}
                onOpenChange={nextOpen => {
                    if (!isDeleting) {
                        setIsDeleteDialogOpen(nextOpen);
                    }
                }}
            >
                <DialogFixedContent className="sm:max-w-[560px]">
                    <DialogHeader>
                        <DialogTitle>Delete Github App</DialogTitle>
                    </DialogHeader>
                    <div className="px-4">
                        <Separator className="opacity-50" />
                    </div>
                    <DialogBody className="flex flex-col gap-4">
                        <p className="text-sm leading-6 text-foreground">
                            Deleting this GitHub App from HivePaaS will not delete it on GitHub. You need to manually
                            delete the GitHub App on GitHub{" "}
                            {advancedSettingsUrl ? (
                                <a
                                    href={advancedSettingsUrl}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="text-primary underline underline-offset-4 hover:opacity-80 font-medium"
                                >
                                    here
                                </a>
                            ) : (
                                "here"
                            )}
                            .
                        </p>
                        <div className="flex items-center gap-2.5 rounded-md border border-destructive bg-destructive/10 px-3.5 py-2.5 text-sm text-destructive font-medium">
                            <AlertTriangle className="size-4 shrink-0 text-destructive" />
                            <span>Warning: This action is permanent and cannot be undone.</span>
                        </div>
                    </DialogBody>
                    <DialogActionFooter className="flex justify-end">
                        <Button
                            type="button"
                            variant="destructive"
                            onClick={handleDelete}
                            isLoading={isDeleting}
                            disabled={isDeleting}
                            className="min-w-[120px]"
                        >
                            Delete this Github App
                        </Button>
                    </DialogActionFooter>
                </DialogFixedContent>
            </Dialog>
        </>
    );
}

interface Props {
    scope: GithubAppTableScope;
    githubApp: SettingGithubApp;
}

export const GithubAppMenuCell = memo(View);
