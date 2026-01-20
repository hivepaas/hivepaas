import React, { useState } from "react";

import { MoreVertical, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import type { ProjectAppDetails } from "~/projects/domain";

import { PopConfirm } from "@application/shared/components";
import { ProjectAppsCommands } from "@application/modules/projects/data/commands";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";

function View({ projectId, app }: Props) {
    const [open, setOpen] = useState(false);

    const { mutate: deleteOne, isPending: isDeleting } = ProjectAppsCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project app deleted successfully");
            setOpen(false);
        },
    });

    return (
        <DropdownMenu
            open={open}
            onOpenChange={setOpen}
        >
            <DropdownMenuTrigger
                asChild
                className="h-8 w-8"
            >
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
                    <PopConfirm
                        title="Delete Project App"
                        variant="destructive"
                        confirmText="Delete"
                        cancelText="Cancel"
                        description="Are you sure you want to delete this project app?"
                        onConfirm={() => {
                            deleteOne({ projectID: projectId, appID: app.id });
                        }}
                    >
                        <Button
                            className="justify-start py-1.5"
                            variant="ghost"
                            disabled={isDeleting}
                        >
                            <Trash2Icon className="mr-2 size-4" />
                            Remove Project App
                        </Button>
                    </PopConfirm>
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    projectId: string;
    app: ProjectAppDetails;
}

export const MenuCell = React.memo(View);
