import React from "react";

import { Button } from "@components/ui/button";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@components/ui/dropdown-menu";
import { Check, Lock, MoreVertical } from "lucide-react";

import { EUserStatus } from "@application/shared/enums";

function View({ status, onActivate, onDisable }: Props) {
    // Logic: If status = 'active' or 'pending': show 'Disable user'
    //        If status = 'disabled': show 'Activate user'
    //        If status = 'pending': do not show toggle button (Disable user)
    //        Always show: 'Remove user'
    const showDisableUser = status === EUserStatus.Active;
    const showActivateUser = status === EUserStatus.Disabled;

    if (!showDisableUser && !showActivateUser) {
        return null;
    }

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
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
                {showActivateUser && (
                    <DropdownMenuItem onClick={onActivate}>
                        <Check className="mr-2 size-4" />
                        Activate user
                    </DropdownMenuItem>
                )}
                {showDisableUser && (
                    <DropdownMenuItem onClick={onDisable}>
                        <Lock className="mr-2 size-4" />
                        Disable user
                    </DropdownMenuItem>
                )}
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    status: EUserStatus;
    onActivate?: () => void;
    onDisable?: () => void;
}

export const UserMenuCell = React.memo(View);
