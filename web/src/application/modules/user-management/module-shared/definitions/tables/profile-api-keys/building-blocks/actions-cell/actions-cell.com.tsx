import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, Settings2, Trash2Icon } from "lucide-react";
import { toast } from "sonner";

import { PopConfirm } from "@application/shared/components";
import { ProfileCommands } from "@application/shared/data/commands";
import { useUpdateApiKeyStatusDialog } from "@application/shared/dialogs";
import type { ProfileApiKey } from "@application/shared/entities";

function View({ apiKey }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateApiKeyStatusDialog({
        onClose: () => {
            updateStatusDialog.actions.close();
        },
    });

    const { mutate: deleteOneApiKey, isPending: isDeleting } = ProfileCommands.useDeleteOneApiKey({
        onSuccess: () => {
            toast.success("API key deleted successfully");
            setOpen(false);
        },
    });

    const onDelete = () => {
        deleteOneApiKey({ id: apiKey.id });
    };
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
                    <Button
                        className="justify-start py-1.5"
                        variant="ghost"
                        onClick={() => {
                            updateStatusDialog.actions.open(apiKey);
                        }}
                    >
                        <Settings2 className="mr-2 size-4" />
                        Change status
                    </Button>
                    <PopConfirm
                        title="Delete API Key"
                        variant="destructive"
                        confirmText="Delete"
                        cancelText="Cancel"
                        description="Are you sure you want to delete this API key?"
                        onConfirm={() => {
                            onDelete();
                        }}
                    >
                        <Button
                            className="justify-start py-1.5"
                            variant="ghost"
                            disabled={isDeleting}
                        >
                            <Trash2Icon className="mr-2 size-4" />
                            Delete
                        </Button>
                    </PopConfirm>
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    apiKey: ProfileApiKey;
}

export const ActionsCell = React.memo(View);
