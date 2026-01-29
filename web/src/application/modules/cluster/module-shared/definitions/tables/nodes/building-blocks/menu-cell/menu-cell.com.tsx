import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { AlertTriangle, MoreVertical, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { NodesCommands } from "~/cluster/data/commands";
import type { NodeDetails } from "~/cluster/domain";

import { PopConfirm } from "@application/shared/components";

function View({ node }: Props) {
    const [open, setOpen] = useState(false);

    const { mutate: deleteOneNode, isPending: isDeleting } = NodesCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Node deleted successfully");
            setOpen(false);
        },
    });

    const onDelete = () => {
        deleteOneNode({ id: node.id, force: false });
    };

    const onForceDelete = () => {
        deleteOneNode({ id: node.id, force: true });
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
                    <PopConfirm
                        title="Delete Item"
                        variant="destructive"
                        confirmText="Delete"
                        cancelText="Cancel"
                        description="Confirm deletion of this item?"
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
                            Remove
                        </Button>
                    </PopConfirm>
                    <PopConfirm
                        title="Force Delete Item"
                        variant="destructive"
                        confirmText="Force Delete"
                        cancelText="Cancel"
                        description="Confirm deletion of this item?"
                        onConfirm={() => {
                            onForceDelete();
                        }}
                    >
                        <Button
                            className="justify-start py-1.5"
                            variant="ghost"
                            disabled={isDeleting}
                        >
                            <AlertTriangle className="mr-2 size-4" />
                            Force Remove
                        </Button>
                    </PopConfirm>
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    node: NodeDetails;
}

export const MenuCell = React.memo(View);
