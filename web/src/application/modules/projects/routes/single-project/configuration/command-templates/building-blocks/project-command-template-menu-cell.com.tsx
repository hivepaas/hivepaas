import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectCommandTemplateCommands } from "~/projects/data/commands";
import { useUpdateProjectCommandTemplateStatusDialog } from "~/projects/dialogs/update-project-command-template-status";
import type { ProjectCommandTemplate } from "~/projects/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

const PROJECT_SCOPE = { type: "project" } as const;

function View({ projectId, commandTemplate }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateProjectCommandTemplateStatusDialog();
    const { mutate: deleteProjectCommandTemplate, isPending: isDeleting } = ProjectCommandTemplateCommands.useDeleteOne(
        {
            onSuccess: () => {
                toast.success("Project Command Template deleted successfully");
                setOpen(false);
            },
        },
    );

    const isInheritedProject = isInheritedProjectSetting(PROJECT_SCOPE, commandTemplate.inherited);

    function handleDelete() {
        deleteProjectCommandTemplate({
            projectID: projectId,
            id: commandTemplate.id,
        });
    }

    function handleChangeStatus() {
        updateStatusDialog.actions.open(projectId, commandTemplate.id, {
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
                        title="Delete Command Template"
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
    projectId: string;
    commandTemplate: ProjectCommandTemplate;
}

export const ProjectCommandTemplateMenuCell = memo(View);
