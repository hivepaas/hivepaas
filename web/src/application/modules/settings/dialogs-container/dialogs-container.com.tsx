import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    CreateOrEditGithubAppDialog,
    ProvisionGithubAppDialog,
    UpdateAccessTokenStatusDialog,
    UpdateAcmeDnsProviderStatusDialog,
    UpdateBasicAuthStatusDialog,
    UpdateCloudStorageStatusDialog,
    UpdateEmailAccountStatusDialog,
    UpdateGithubAppStatusDialog,
    UpdateImPlatformStatusDialog,
    UpdateNotificationTargetStatusDialog,
    UpdateOAuthStatusDialog,
    UpdateRegistryAuthStatusDialog,
    UpdateSSHKeyStatusDialog,
    UpdateSslCertStatusDialog,
    UpdateSslProviderStatusDialog,
    useCreateOrEditGithubAppDialogState,
    useProvisionGithubAppDialogState,
    useUpdateAccessTokenStatusDialogState,
    useUpdateAcmeDnsProviderStatusDialogState,
    useUpdateBasicAuthStatusDialogState,
    useUpdateCloudStorageStatusDialogState,
    useUpdateEmailAccountStatusDialogState,
    useUpdateGithubAppStatusDialogState,
    useUpdateImPlatformStatusDialogState,
    useUpdateNotificationTargetStatusDialogState,
    useUpdateOAuthStatusDialogState,
    useUpdateRegistryAuthStatusDialogState,
    useUpdateSSHKeyStatusDialogState,
    useUpdateSslCertStatusDialogState,
    useUpdateSslProviderStatusDialogState,
} from "~/settings/dialogs";

function View() {
    const location = useLocation();
    const updateBasicAuthStatusDialog = useUpdateBasicAuthStatusDialogState();
    const updateAcmeDnsProviderStatusDialog = useUpdateAcmeDnsProviderStatusDialogState();
    const updateRegistryAuthStatusDialog = useUpdateRegistryAuthStatusDialogState();
    const updateSslCertStatusDialog = useUpdateSslCertStatusDialogState();
    const updateSslProviderStatusDialog = useUpdateSslProviderStatusDialogState();
    const updateImPlatformStatusDialog = useUpdateImPlatformStatusDialogState();
    const updateEmailAccountStatusDialog = useUpdateEmailAccountStatusDialogState();
    const updateSSHKeyStatusDialog = useUpdateSSHKeyStatusDialogState();
    const updateAccessTokenStatusDialog = useUpdateAccessTokenStatusDialogState();
    const updateCloudStorageStatusDialog = useUpdateCloudStorageStatusDialogState();
    const updateOAuthStatusDialog = useUpdateOAuthStatusDialogState();
    const updateNotificationTargetStatusDialog = useUpdateNotificationTargetStatusDialogState();
    const createOrEditGithubAppDialog = useCreateOrEditGithubAppDialogState();
    const provisionGithubAppDialog = useProvisionGithubAppDialogState();
    const updateGithubAppStatusDialog = useUpdateGithubAppStatusDialogState();

    useUpdateEffect(() => {
        updateBasicAuthStatusDialog.destroy();
        updateAcmeDnsProviderStatusDialog.destroy();
        updateRegistryAuthStatusDialog.destroy();
        updateSslCertStatusDialog.destroy();
        updateSslProviderStatusDialog.destroy();
        updateImPlatformStatusDialog.destroy();
        updateEmailAccountStatusDialog.destroy();
        updateSSHKeyStatusDialog.destroy();
        updateAccessTokenStatusDialog.destroy();
        updateCloudStorageStatusDialog.destroy();
        updateOAuthStatusDialog.destroy();
        updateNotificationTargetStatusDialog.destroy();
        createOrEditGithubAppDialog.destroy();
        provisionGithubAppDialog.destroy();
        updateGithubAppStatusDialog.destroy();
    }, [location]);

    return (
        <>
            <UpdateBasicAuthStatusDialog />
            <UpdateAcmeDnsProviderStatusDialog />
            <UpdateRegistryAuthStatusDialog />
            <UpdateSslCertStatusDialog />
            <UpdateSslProviderStatusDialog />
            <UpdateImPlatformStatusDialog />
            <UpdateEmailAccountStatusDialog />
            <UpdateSSHKeyStatusDialog />
            <UpdateAccessTokenStatusDialog />
            <UpdateCloudStorageStatusDialog />
            <UpdateOAuthStatusDialog />
            <UpdateNotificationTargetStatusDialog />
            <CreateOrEditGithubAppDialog />
            <ProvisionGithubAppDialog />
            <UpdateGithubAppStatusDialog />
        </>
    );
}

export const SettingsDialogsContainer = memo(View);
