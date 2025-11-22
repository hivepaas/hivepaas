import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

import {
    ChangePasswordDialog,
    useChangePasswordDialogState,
    F2aSetupDialog,
    useF2aSetupDialogState,
} from "@application/shared/dialogs";

function View() {
    const location = useLocation();
    const changePasswordDialog = useChangePasswordDialogState();
    const f2aSetupDialog = useF2aSetupDialogState();

    useUpdateEffect(() => {
        changePasswordDialog.destroy();
        f2aSetupDialog.destroy();
    }, [location]);

    return (
        <>
            <ChangePasswordDialog />
            <F2aSetupDialog />
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const CommonDialogsContainer = React.memo(View);
