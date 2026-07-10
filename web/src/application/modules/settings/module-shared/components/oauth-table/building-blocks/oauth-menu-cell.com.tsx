import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { OAuthCommands } from "~/settings/data/commands";
import { useUpdateOAuthStatusDialog } from "~/settings/dialogs/update-oauth-status";
import type { SettingOAuth } from "~/settings/domain";

import { PopConfirm } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

function View({ oauth }: Props) {
    const [open, setOpen] = useState(false);
    const updateStatusDialog = useUpdateOAuthStatusDialog();
    const { canWrite, canDelete } = useConditionalModule({ id: MODULE_IDS.Settings });

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
                    <PermissionTooltipAction
                        id={MODULE_IDS.Settings}
                        action="write"
                        triggerClassName="w-full"
                    >
                        {({ isDenied }) => (
                            <Button
                                className="justify-start py-1.5 w-full"
                                variant="ghost"
                                disabled={isDenied}
                                onClick={() => {
                                    if (!canWrite) {
                                        return;
                                    }

                                    updateStatusDialog.actions.open(oauth.id);
                                    setOpen(false);
                                }}
                            >
                                <SlidersHorizontal className="mr-2 size-4" />
                                Change Status
                            </Button>
                        )}
                    </PermissionTooltipAction>
                    {canDelete ? (
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
                                Delete
                            </Button>
                        </PopConfirm>
                    ) : (
                        <PermissionTooltipAction
                            id={MODULE_IDS.Settings}
                            action="delete"
                            triggerClassName="w-full"
                        >
                            {({ isDenied }) => (
                                <Button
                                    className="justify-start py-1.5 w-full"
                                    variant="ghost"
                                    disabled={isDenied}
                                >
                                    <Trash2Icon className="mr-2 size-4" />
                                    Delete
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    )}
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    oauth: SettingOAuth;
}

export const OAuthMenuCell = memo(View);
