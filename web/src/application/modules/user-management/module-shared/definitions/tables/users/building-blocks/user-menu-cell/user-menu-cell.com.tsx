import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { Check, Lock, MoreVertical, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { UsersCommands } from "~/user-management/data/commands";
import type { UserBase } from "~/user-management/domain";

import { PopConfirm } from "@application/shared/components/pop-confirm";
import { EUserStatus } from "@application/shared/enums";

function View({ user }: Props) {
    const [open, setOpen] = useState(false);

    // Logic: If status = 'active' or 'pending': show 'Disable user'
    //        If status = 'disabled': show 'Activate user'
    //        If status = 'pending': do not show toggle button (Disable user)
    //        Always show: 'Remove user'
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
                        <Button
                            className="justify-start py-1.5"
                            variant="ghost"
                            onClick={() => {
                                updateOne({ user: { status: EUserStatus.Active, id: user.id } });
                            }}
                            isLoading={isUpdating}
                        >
                            <Check className="mr-2 size-4" />
                            Activate User
                        </Button>
                    )}
                    {showDisableUser && (
                        <Button
                            className="justify-start py-1.5"
                            variant="ghost"
                            onClick={() => {
                                updateOne({ user: { status: EUserStatus.Disabled, id: user.id } });
                            }}
                            isLoading={isUpdating}
                        >
                            <Lock className="mr-2 size-4" />
                            Disable User
                        </Button>
                    )}
                    <PopConfirm
                        title="Delete User"
                        variant="destructive"
                        confirmText="Delete"
                        cancelText="Cancel"
                        description="Are you sure you want to delete this user?"
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
                            Remove user
                        </Button>
                    </PopConfirm>
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    user: UserBase;
}

export const UserMenuCell = React.memo(View);
