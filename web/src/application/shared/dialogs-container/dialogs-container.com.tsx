import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

import {
    ChangePasswordDialog,
    CreateFeedbackDialog,
    CreateProfileApiKeyDialog,
    F2aSetupDialog,
    GlobalAlertDialog,
    UpdateApiKeyStatusDialog,
    useChangePasswordDialogState,
    useCreateFeedbackDialogState,
    useCreateProfileApiKeyDialogState,
    useF2aSetupDialogState,
    useGlobalAlertDialogState,
    useUpdateApiKeyStatusDialogState,
} from "@application/shared/dialogs";

function View() {
    const location = useLocation();
    const changePasswordDialog = useChangePasswordDialogState();
    const createFeedbackDialog = useCreateFeedbackDialogState();
    const f2aSetupDialog = useF2aSetupDialogState();
    const createProfileApiKeyDialog = useCreateProfileApiKeyDialogState();
    const updateApiKeyStatusDialog = useUpdateApiKeyStatusDialogState();
    const globalAlertDialog = useGlobalAlertDialogState();

    useUpdateEffect(() => {
        changePasswordDialog.destroy();
        createFeedbackDialog.destroy();
        f2aSetupDialog.destroy();
        createProfileApiKeyDialog.destroy();
        updateApiKeyStatusDialog.destroy();
        globalAlertDialog.destroy();
    }, [location]);

    return (
        <>
            <ChangePasswordDialog />
            <CreateFeedbackDialog />
            <F2aSetupDialog />
            <CreateProfileApiKeyDialog />
            <UpdateApiKeyStatusDialog />
            <GlobalAlertDialog />
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const CommonDialogsContainer = React.memo(View);
