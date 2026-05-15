import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    CreateOrEditBasicAuthDialog,
    CreateOrEditEmailAccountDialog,
    CreateOrEditImPlatformDialog,
    CreateOrEditRegistryAuthDialog,
    CreateOrEditSslCertDialog,
    UpdateBasicAuthStatusDialog,
    UpdateEmailAccountStatusDialog,
    UpdateImPlatformStatusDialog,
    UpdateRegistryAuthStatusDialog,
    UpdateSslCertStatusDialog,
    useCreateOrEditBasicAuthDialogState,
    useCreateOrEditEmailAccountDialogState,
    useCreateOrEditImPlatformDialogState,
    useCreateOrEditRegistryAuthDialogState,
    useCreateOrEditSslCertDialogState,
    useUpdateBasicAuthStatusDialogState,
    useUpdateEmailAccountStatusDialogState,
    useUpdateImPlatformStatusDialogState,
    useUpdateRegistryAuthStatusDialogState,
    useUpdateSslCertStatusDialogState,
} from "~/settings/dialogs";

function View() {
    const location = useLocation();
    const createOrEditBasicAuthDialog = useCreateOrEditBasicAuthDialogState();
    const updateBasicAuthStatusDialog = useUpdateBasicAuthStatusDialogState();
    const createOrEditRegistryAuthDialog = useCreateOrEditRegistryAuthDialogState();
    const updateRegistryAuthStatusDialog = useUpdateRegistryAuthStatusDialogState();
    const createOrEditSslCertDialog = useCreateOrEditSslCertDialogState();
    const updateSslCertStatusDialog = useUpdateSslCertStatusDialogState();
    const createOrEditImPlatformDialog = useCreateOrEditImPlatformDialogState();
    const updateImPlatformStatusDialog = useUpdateImPlatformStatusDialogState();
    const createOrEditEmailAccountDialog = useCreateOrEditEmailAccountDialogState();
    const updateEmailAccountStatusDialog = useUpdateEmailAccountStatusDialogState();

    useUpdateEffect(() => {
        createOrEditBasicAuthDialog.destroy();
        updateBasicAuthStatusDialog.destroy();
        createOrEditRegistryAuthDialog.destroy();
        updateRegistryAuthStatusDialog.destroy();
        createOrEditSslCertDialog.destroy();
        updateSslCertStatusDialog.destroy();
        createOrEditImPlatformDialog.destroy();
        updateImPlatformStatusDialog.destroy();
        createOrEditEmailAccountDialog.destroy();
        updateEmailAccountStatusDialog.destroy();
    }, [location]);

    return (
        <>
            <CreateOrEditBasicAuthDialog />
            <UpdateBasicAuthStatusDialog />
            <CreateOrEditRegistryAuthDialog />
            <UpdateRegistryAuthStatusDialog />
            <CreateOrEditSslCertDialog />
            <UpdateSslCertStatusDialog />
            <CreateOrEditImPlatformDialog />
            <UpdateImPlatformStatusDialog />
            <CreateOrEditEmailAccountDialog />
            <UpdateEmailAccountStatusDialog />
        </>
    );
}

export const SettingsDialogsContainer = memo(View);
