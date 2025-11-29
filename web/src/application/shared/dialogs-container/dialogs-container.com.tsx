import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

import {
    ChangePasswordDialog,
    CreateProfileApiKeyDialog,
    F2aSetupDialog,
    useChangePasswordDialogState,
    useCreateProfileApiKeyDialogState,
    useF2aSetupDialogState,
} from "@application/shared/dialogs";

function View() {
    const location = useLocation();
    const changePasswordDialog = useChangePasswordDialogState();
    const f2aSetupDialog = useF2aSetupDialogState();
    const createProfileApiKeyDialog = useCreateProfileApiKeyDialogState();

    useUpdateEffect(() => {
        changePasswordDialog.destroy();
        f2aSetupDialog.destroy();
        createProfileApiKeyDialog.destroy();
    }, [location]);

    return (
        <>
            <ChangePasswordDialog />
            <F2aSetupDialog />
            <CreateProfileApiKeyDialog />
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const CommonDialogsContainer = React.memo(View);
