import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

import {
    ChangePasswordDialog,
    CreateProfileApiKeyDialog,
    F2aSetupDialog,
    UpdateApiKeyStatusDialog,
    useChangePasswordDialogState,
    useCreateProfileApiKeyDialogState,
    useF2aSetupDialogState,
    useUpdateApiKeyStatusDialogState,
} from "@application/shared/dialogs";

function View() {
    const location = useLocation();
    const changePasswordDialog = useChangePasswordDialogState();
    const f2aSetupDialog = useF2aSetupDialogState();
    const createProfileApiKeyDialog = useCreateProfileApiKeyDialogState();
    const updateApiKeyStatusDialog = useUpdateApiKeyStatusDialogState();

    useUpdateEffect(() => {
        changePasswordDialog.destroy();
        f2aSetupDialog.destroy();
        createProfileApiKeyDialog.destroy();
        updateApiKeyStatusDialog.destroy();
    }, [location]);

    return (
        <>
            <ChangePasswordDialog />
            <F2aSetupDialog />
            <CreateProfileApiKeyDialog />
            <UpdateApiKeyStatusDialog />
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const CommonDialogsContainer = React.memo(View);
