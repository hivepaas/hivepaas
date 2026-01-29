import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import type { ProjectBaseEntity } from "~/projects/domain";

import { PopConfirm } from "@application/shared/components";

import { ProjectsCommands } from "@application/modules/projects/data/commands";

function View({ project }: Props) {
    void project;
    const [open, setOpen] = useState(false);

    const { mutate: deleteOne, isPending: isDeleting } = ProjectsCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project deleted successfully");
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
                        title="Delete Item"
                        variant="destructive"
                        confirmText="Delete"
                        cancelText="Cancel"
                        description="Confirm deletion of this item?"
                        onConfirm={() => {
                            deleteOne({ projectID: project.id });
                        }}
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
    project: ProjectBaseEntity;
}

export const MenuCell = React.memo(View);
