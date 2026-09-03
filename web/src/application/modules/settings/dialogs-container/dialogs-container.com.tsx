import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    ProvisionGithubAppDialog,
    UpdateAccessTokenStatusDialog,
    UpdateAcmeDnsProviderStatusDialog,
    UpdateBackupRepoPasswordDialog,
    UpdateBackupRepoStatusDialog,
    UpdateBasicAuthStatusDialog,
    UpdateCloudStorageStatusDialog,
    UpdateEmailAccountStatusDialog,
    UpdateGithubAppStatusDialog,
    UpdateImPlatformStatusDialog,
    UpdateNotificationTargetStatusDialog,
    UpdateOAuthStatusDialog,
    UpdateRegistryAuthStatusDialog,
    UpdateRepoWebhookStatusDialog,
    UpdateSSHKeyStatusDialog,
    UpdateSslCertStatusDialog,
    UpdateSslProviderStatusDialog,
    useProvisionGithubAppDialogState,
    useUpdateAccessTokenStatusDialogState,
    useUpdateAcmeDnsProviderStatusDialogState,
    useUpdateBackupRepoPasswordDialogState,
    useUpdateBackupRepoStatusDialogState,
    useUpdateBasicAuthStatusDialogState,
    useUpdateCloudStorageStatusDialogState,
    useUpdateEmailAccountStatusDialogState,
    useUpdateGithubAppStatusDialogState,
    useUpdateImPlatformStatusDialogState,
    useUpdateNotificationTargetStatusDialogState,
    useUpdateOAuthStatusDialogState,
    useUpdateRegistryAuthStatusDialogState,
    useUpdateRepoWebhookStatusDialogState,
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
    const provisionGithubAppDialog = useProvisionGithubAppDialogState();
    const updateGithubAppStatusDialog = useUpdateGithubAppStatusDialogState();
    const updateRepoWebhookStatusDialog = useUpdateRepoWebhookStatusDialogState();
    const updateBackupRepoStatusDialog = useUpdateBackupRepoStatusDialogState();
    const updateBackupRepoPasswordDialog = useUpdateBackupRepoPasswordDialogState();

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
        provisionGithubAppDialog.destroy();
        updateGithubAppStatusDialog.destroy();
        updateRepoWebhookStatusDialog.destroy();
        updateBackupRepoStatusDialog.destroy();
        updateBackupRepoPasswordDialog.destroy();
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
            <ProvisionGithubAppDialog />
            <UpdateGithubAppStatusDialog />
            <UpdateRepoWebhookStatusDialog />
            <UpdateBackupRepoStatusDialog />
            <UpdateBackupRepoPasswordDialog />
        </>
    );
}

export const SettingsDialogsContainer = memo(View);
