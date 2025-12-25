import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

import {
    InviteUserDialog,
    ResetUserPasswordDialog,
    useInviteUserDialogState,
    useResetUserPasswordDialogState,
} from "../dialogs";

function View() {
    const location = useLocation();
    const inviteUserDialog = useInviteUserDialogState();
    const resetUserPasswordDialog = useResetUserPasswordDialogState();

    useUpdateEffect(() => {
        inviteUserDialog.destroy();
        resetUserPasswordDialog.destroy();
    }, [location]);

    return (
        <>
            <InviteUserDialog />
            <ResetUserPasswordDialog />
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const UserManagementDialogsContainer = React.memo(View);
