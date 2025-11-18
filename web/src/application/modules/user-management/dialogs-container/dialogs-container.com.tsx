import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

import { InviteUserDialog, useInviteUserDialogState } from "../dialogs";

function View() {
    const location = useLocation();
    const inviteUserDialog = useInviteUserDialogState();

    useUpdateEffect(() => {
        inviteUserDialog.destroy();
    }, [location]);

    return (
        <>
            <InviteUserDialog />
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const UserManagementDialogsContainer = React.memo(View);
