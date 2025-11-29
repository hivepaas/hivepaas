import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, Settings2, Trash2Icon } from "lucide-react";
import { toast } from "sonner";

import { PopConfirm } from "@application/shared/components";
import { ProfileCommands } from "@application/shared/data/commands";

function View({ id }: Props) {
    const [open, setOpen] = useState(false);

    const { mutate: deleteOneApiKey, isPending: isDeleting } = ProfileCommands.useDeleteOneApiKey({
        onSuccess: () => {
            toast.success("API key deleted successfully");
            setOpen(false);
        },
    });

    const onChangeStatus = () => {
        console.log("onChangeStatus");
        setOpen(false);
    };
    const onDelete = () => {
        deleteOneApiKey({ id });
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
                    <span className="sr-only">User menu</span>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <div className="flex flex-col gap-0">
                    <Button
                        className="justify-start py-1.5"
                        variant="ghost"
                        onClick={onChangeStatus}
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
    id: string;
}

export const ActionsCell = React.memo(View);
