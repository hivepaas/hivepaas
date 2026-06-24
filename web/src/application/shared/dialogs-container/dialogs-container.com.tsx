import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

import {
    ChangePasswordDialog,
    CreateFeedbackDialog,
    F2aSetupDialog,
    GlobalAlertDialog,
    UpdateApiKeyStatusDialog,
    useChangePasswordDialogState,
    useCreateFeedbackDialogState,
    useF2aSetupDialogState,
    useGlobalAlertDialogState,
    useUpdateApiKeyStatusDialogState,
} from "@application/shared/dialogs";

function View() {
    const location = useLocation();
    const changePasswordDialog = useChangePasswordDialogState();
    const createFeedbackDialog = useCreateFeedbackDialogState();
    const f2aSetupDialog = useF2aSetupDialogState();
    const updateApiKeyStatusDialog = useUpdateApiKeyStatusDialogState();
    const globalAlertDialog = useGlobalAlertDialogState();

    useUpdateEffect(() => {
        changePasswordDialog.destroy();
        createFeedbackDialog.destroy();
        f2aSetupDialog.destroy();
        updateApiKeyStatusDialog.destroy();
        globalAlertDialog.destroy();
    }, [location]);

    return (
        <>
            <ChangePasswordDialog />
            <CreateFeedbackDialog />
            <F2aSetupDialog />
            <UpdateApiKeyStatusDialog />
            <GlobalAlertDialog />
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const CommonDialogsContainer = React.memo(View);
