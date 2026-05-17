import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    CreateOrEditAccessTokenDialog,
    CreateOrEditBasicAuthDialog,
    CreateOrEditCloudStorageDialog,
    CreateOrEditEmailAccountDialog,
    CreateOrEditImPlatformDialog,
    CreateOrEditNotificationTargetDialog,
    CreateOrEditOAuthDialog,
    CreateOrEditRegistryAuthDialog,
    CreateOrEditSSHKeyDialog,
    CreateOrEditSslCertDialog,
    UpdateAccessTokenStatusDialog,
    UpdateBasicAuthStatusDialog,
    UpdateCloudStorageStatusDialog,
    UpdateEmailAccountStatusDialog,
    UpdateImPlatformStatusDialog,
    UpdateNotificationTargetStatusDialog,
    UpdateOAuthStatusDialog,
    UpdateRegistryAuthStatusDialog,
    UpdateSSHKeyStatusDialog,
    UpdateSslCertStatusDialog,
    useCreateOrEditAccessTokenDialogState,
    useCreateOrEditBasicAuthDialogState,
    useCreateOrEditCloudStorageDialogState,
    useCreateOrEditEmailAccountDialogState,
    useCreateOrEditImPlatformDialogState,
    useCreateOrEditNotificationTargetDialogState,
    useCreateOrEditOAuthDialogState,
    useCreateOrEditRegistryAuthDialogState,
    useCreateOrEditSSHKeyDialogState,
    useCreateOrEditSslCertDialogState,
    useUpdateAccessTokenStatusDialogState,
    useUpdateBasicAuthStatusDialogState,
    useUpdateCloudStorageStatusDialogState,
    useUpdateEmailAccountStatusDialogState,
    useUpdateImPlatformStatusDialogState,
    useUpdateNotificationTargetStatusDialogState,
    useUpdateOAuthStatusDialogState,
    useUpdateRegistryAuthStatusDialogState,
    useUpdateSSHKeyStatusDialogState,
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
    const createOrEditSSHKeyDialog = useCreateOrEditSSHKeyDialogState();
    const updateSSHKeyStatusDialog = useUpdateSSHKeyStatusDialogState();
    const createOrEditAccessTokenDialog = useCreateOrEditAccessTokenDialogState();
    const updateAccessTokenStatusDialog = useUpdateAccessTokenStatusDialogState();
    const createOrEditCloudStorageDialog = useCreateOrEditCloudStorageDialogState();
    const updateCloudStorageStatusDialog = useUpdateCloudStorageStatusDialogState();
    const createOrEditOAuthDialog = useCreateOrEditOAuthDialogState();
    const updateOAuthStatusDialog = useUpdateOAuthStatusDialogState();
    const createOrEditNotificationTargetDialog = useCreateOrEditNotificationTargetDialogState();
    const updateNotificationTargetStatusDialog = useUpdateNotificationTargetStatusDialogState();

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
        createOrEditSSHKeyDialog.destroy();
        updateSSHKeyStatusDialog.destroy();
        createOrEditAccessTokenDialog.destroy();
        updateAccessTokenStatusDialog.destroy();
        createOrEditCloudStorageDialog.destroy();
        updateCloudStorageStatusDialog.destroy();
        createOrEditOAuthDialog.destroy();
        updateOAuthStatusDialog.destroy();
        createOrEditNotificationTargetDialog.destroy();
        updateNotificationTargetStatusDialog.destroy();
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
            <CreateOrEditSSHKeyDialog />
            <UpdateSSHKeyStatusDialog />
            <CreateOrEditAccessTokenDialog />
            <UpdateAccessTokenStatusDialog />
            <CreateOrEditCloudStorageDialog />
            <UpdateCloudStorageStatusDialog />
            <CreateOrEditOAuthDialog />
            <UpdateOAuthStatusDialog />
            <CreateOrEditNotificationTargetDialog />
            <UpdateNotificationTargetStatusDialog />
        </>
    );
}

export const SettingsDialogsContainer = memo(View);
