import React from "react";

import { useLocation, useUpdateEffect } from "react-use";

import {
    ChangePasswordDialog,
    CreateFeedbackDialog,
    F2aSetupDialog,
    GlobalAlertDialog,
    MfaSetupRequiredDialog,
    UpdateApiKeyStatusDialog,
    useChangePasswordDialogState,
    useCreateFeedbackDialogState,
    useF2aSetupDialogState,
    useGlobalAlertDialogState,
    useMfaSetupRequiredDialogState,
    useUpdateApiKeyStatusDialogState,
} from "@application/shared/dialogs";

import { useAuthContext } from "@application/authentication/context";

function View() {
    const location = useLocation();
    const changePasswordDialog = useChangePasswordDialogState();
    const createFeedbackDialog = useCreateFeedbackDialogState();
    const f2aSetupDialog = useF2aSetupDialogState();
    const mfaSetupRequiredDialog = useMfaSetupRequiredDialogState();
    const updateApiKeyStatusDialog = useUpdateApiKeyStatusDialogState();
    const globalAlertDialog = useGlobalAlertDialogState();

    useUpdateEffect(() => {
        changePasswordDialog.destroy();
        createFeedbackDialog.destroy();
        updateApiKeyStatusDialog.destroy();
        globalAlertDialog.destroy();

        // Keep enforced MFA dialogs open across route changes so the user cannot escape setup.
        const authData = useAuthContext.getState().data;
        const mfaSetupRequired = "mfaSetupRequired" in authData && authData.mfaSetupRequired;
        if (!mfaSetupRequired) {
            f2aSetupDialog.destroy();
            mfaSetupRequiredDialog.destroy();
        }
    }, [location]);

    return (
        <>
            <ChangePasswordDialog />
            <CreateFeedbackDialog />
            <MfaSetupRequiredDialog />
            <F2aSetupDialog />
            <UpdateApiKeyStatusDialog />
            <GlobalAlertDialog />
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const CommonDialogsContainer = React.memo(View);
