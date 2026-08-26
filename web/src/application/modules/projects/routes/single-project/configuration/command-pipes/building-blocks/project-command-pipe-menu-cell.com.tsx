import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectCommandPipeCommands } from "~/projects/data/commands";
import { useUpdateProjectCommandPipeStatusDialog } from "~/projects/dialogs/update-project-command-pipe-status";
import type { ProjectCommandPipe } from "~/projects/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

const PROJECT_SCOPE = { type: "project" } as const;

function View({ projectId, env, commandPipe }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateProjectCommandPipeStatusDialog();
    const { mutate: deleteProjectCommandPipe, isPending: isDeleting } = ProjectCommandPipeCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project Command Pipe deleted successfully");
            setOpen(false);
        },
    });

    const isInheritedProject = isInheritedProjectSetting(PROJECT_SCOPE, commandPipe.inherited);

    function handleDelete() {
        deleteProjectCommandPipe({
            projectID: projectId,
            env,
            id: commandPipe.id,
        });
    }

    function handleChangeStatus() {
        updateStatusDialog.actions.open(projectId, commandPipe.id, {
            env,
            props: {
                readOnlyInherited: isInheritedProject,
            },
        });
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
                        scope={PROJECT_SCOPE}
                        action="write"
                        onClick={handleChangeStatus}
                    >
                        <SlidersHorizontal className="mr-2 size-4" />
                        Change Status
                    </SettingsScopeMenuButton>
                    <SettingsScopePopConfirmButton
                        scope={PROJECT_SCOPE}
                        action="delete"
                        title="Remove Command Pipe"
                        confirmText="Remove"
                        cancelText="Cancel"
                        description="Confirm removal of this item?"
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
    projectId: string;
    env?: string;
    commandPipe: ProjectCommandPipe;
}

export const ProjectCommandPipeMenuCell = memo(View);
