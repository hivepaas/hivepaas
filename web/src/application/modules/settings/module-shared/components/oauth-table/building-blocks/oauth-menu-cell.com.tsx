import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { OAuthCommands } from "~/settings/data/commands";
import { useUpdateOAuthStatusDialog } from "~/settings/dialogs/update-oauth-status";
import type { SettingOAuth } from "~/settings/domain";

import { PopConfirm } from "@application/shared/components";

function View({ oauth }: Props) {
    const [open, setOpen] = useState(false);
    const updateStatusDialog = useUpdateOAuthStatusDialog();

    const { mutate: deleteOAuth, isPending: isDeleting } = OAuthCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("OAuth deleted successfully");
            setOpen(false);
        },
    });

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
                            updateStatusDialog.actions.open(oauth.id);
                            setOpen(false);
                        }}
                    >
                        <SlidersHorizontal className="mr-2 size-4" />
                        Change Status
                    </Button>
                    <PopConfirm
                        title="Delete OAuth"
                        variant="destructive"
                        confirmText="Delete"
                        cancelText="Cancel"
                        description="Confirm deletion of this item?"
                        onConfirm={() => {
                            deleteOAuth({ id: oauth.id });
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
    oauth: SettingOAuth;
}

export const OAuthMenuCell = memo(View);
