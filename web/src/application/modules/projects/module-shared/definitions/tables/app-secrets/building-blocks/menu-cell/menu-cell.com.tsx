import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectAppSecretsCommands } from "~/projects/data/commands";
import type { AppSecret } from "~/projects/domain";

import { PopConfirm } from "@application/shared/components";

function View({ projectId, appId, secret }: Props) {
    const [open, setOpen] = useState(false);

    const { mutate: deleteOne, isPending: isDeleting } = ProjectAppSecretsCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("App secret deleted successfully");
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
                            deleteOne({ projectID: projectId, appID: appId, secretID: secret.id });
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
    projectId: string;
    appId: string;
    secret: AppSecret;
}

export const MenuCell = React.memo(View);
