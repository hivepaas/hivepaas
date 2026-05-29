import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { Check, Lock, MoreVertical, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { UsersCommands } from "~/user-management/data/commands";
import type { UserBase } from "~/user-management/domain";

import { PopConfirm } from "@application/shared/components/pop-confirm";
import { MODULE_IDS } from "@application/shared/constants";
import { EUserStatus } from "@application/shared/enums";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

function View({ user }: Props) {
    const [open, setOpen] = useState(false);
    const { canWrite, canDelete } = useConditionalModule({ id: MODULE_IDS.User });

    // Logic: If status = 'active' or 'pending': show 'Disable User'
    //        If status = 'disabled': show 'Activate User'
    //        If status = 'pending': do not show toggle button (Disable User)
    //        Always show: 'Remove User'
    const showDisableUser = user.status === EUserStatus.Active;
    const showActivateUser = user.status === EUserStatus.Disabled;

    const { mutate: deleteOne, isPending: isDeleting } = UsersCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("User removed successfully");
            setOpen(false);
        },
    });

    const { mutate: updateOne, isPending: isUpdating } = UsersCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("User status updated successfully");
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
                    <span className="sr-only">User menu</span>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <div className="flex flex-col gap-0">
                    {showActivateUser && (
                        <PermissionTooltipAction
                            id={MODULE_IDS.User}
                            action="write"
                            triggerClassName="w-full"
                        >
                            {({ isDenied }) => (
                                <Button
                                    className="justify-start py-1.5 w-full"
                                    variant="ghost"
                                    onClick={() => {
                                        if (!canWrite) {
                                            return;
                                        }

                                        updateOne({ user: { status: EUserStatus.Active, id: user.id } });
                                    }}
                                    disabled={isDenied}
                                    isLoading={isUpdating}
                                >
                                    <Check className="mr-2 size-4" />
                                    Activate User
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    )}
                    {showDisableUser && canWrite && (
                        <PopConfirm
                            title="Disable User"
                            variant="destructive"
                            confirmText="Disable"
                            cancelText="Cancel"
                            description="Confirm disabling of this user?"
                            onConfirm={() => {
                                updateOne({ user: { status: EUserStatus.Disabled, id: user.id } });
                            }}
                        >
                            <Button
                                className="justify-start py-1.5"
                                variant="ghost"
                                isLoading={isUpdating}
                            >
                                <Lock className="mr-2 size-4" />
                                Disable User
                            </Button>
                        </PopConfirm>
                    )}
                    {showDisableUser && !canWrite && (
                        <PermissionTooltipAction
                            id={MODULE_IDS.User}
                            action="write"
                            triggerClassName="w-full"
                        >
                            {({ isDenied }) => (
                                <Button
                                    className="justify-start py-1.5 w-full"
                                    variant="ghost"
                                    disabled={isDenied}
                                >
                                    <Lock className="mr-2 size-4" />
                                    Disable User
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    )}
                    {canDelete ? (
                        <PopConfirm
                            title="Delete Item"
                            variant="destructive"
                            confirmText="Delete"
                            cancelText="Cancel"
                            description="Confirm deletion of this item?"
                            onConfirm={() => {
                                deleteOne({ id: user.id });
                            }}
                        >
                            <Button
                                className="justify-start py-1.5"
                                variant="ghost"
                                isLoading={isDeleting}
                            >
                                <Trash2Icon className="mr-2 size-4" />
                                Remove User
                            </Button>
                        </PopConfirm>
                    ) : (
                        <PermissionTooltipAction
                            id={MODULE_IDS.User}
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
                                    Remove User
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
    user: UserBase;
}

export const UserMenuCell = React.memo(View);
