import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectImServiceCommands } from "~/projects/data/commands";
import { ImServiceCommands } from "~/settings/data/commands";
import { useUpdateImPlatformStatusDialog } from "~/settings/dialogs/update-im-platform-status";
import type { SettingImService } from "~/settings/domain";

import { PopConfirm } from "@application/shared/components";

import type { ImPlatformTableScope } from "../../im-platform-table.types";

function View({ scope, imPlatform }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateImPlatformStatusDialog();

    const { mutate: deleteSettingImPlatform, isPending: isDeletingSetting } = ImServiceCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("IM platform deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectImPlatform, isPending: isDeletingProject } = ProjectImServiceCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project IM platform deleted successfully");
            setOpen(false);
        },
    });

    const isDeleting = isDeletingSetting || isDeletingProject;

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectImPlatform({
                projectID: scope.projectId,
                id: imPlatform.id,
            });
            return;
        }

        deleteSettingImPlatform({ id: imPlatform.id });
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
                        onClick={() => {
                            updateStatusDialog.actions.open(scope, imPlatform.id);
                            setOpen(false);
                        }}
                    >
                        <SlidersHorizontal className="mr-2 size-4" />
                        Change Status
                    </Button>
                    <PopConfirm
                        title="Delete IM platform"
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
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    scope: ImPlatformTableScope;
    imPlatform: SettingImService;
}

export const ImPlatformMenuCell = memo(View);
