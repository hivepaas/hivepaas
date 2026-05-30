import React from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    CreateOrEditAppConfigFileDialog,
    useCreateOrEditAppConfigFileDialogState,
} from "~/projects/dialogs/create-or-edit-app-config-file";
import {
    CreateOrEditAppHealthCheckDialog,
    useCreateOrEditAppHealthCheckDialogState,
} from "~/projects/dialogs/create-or-edit-app-health-check";
import {
    CreateOrEditAppSecretDialog,
    useCreateOrEditAppSecretDialogState,
} from "~/projects/dialogs/create-or-edit-app-secret";
import {
    CreateOrEditProjectSecretDialog,
    useCreateOrEditProjectSecretDialogState,
} from "~/projects/dialogs/create-or-edit-project-secret";
import { CreateProjectDialog, useCreateProjectDialogState } from "~/projects/dialogs/create-project";
import { CreateProjectAppDialog, useCreateProjectAppDialogState } from "~/projects/dialogs/create-project-app";
import {
    ImportProjectSettingsDialog,
    useImportProjectSettingsDialogState,
} from "~/projects/dialogs/import-project-settings";
import { ProjectUserAccessesDialog, useProjectUserAccessesDialogState } from "~/projects/dialogs/project-user-accesses";
import {
    QuickInstallSslCertDialog,
    useQuickInstallSslCertDialogState,
} from "~/projects/dialogs/quick-install-ssl-cert";
import { StorageMountDialog, useStorageMountDialogState } from "~/projects/dialogs/storage-mount";
import {
    UpdateAppHealthCheckStatusDialog,
    useUpdateAppHealthCheckStatusDialogState,
} from "~/projects/dialogs/update-app-health-check-status";
import {
    CreateOrEditAccessTokenDialog,
    CreateOrEditBasicAuthDialog,
    CreateOrEditCloudStorageDialog,
    CreateOrEditEmailAccountDialog,
    CreateOrEditGithubAppDialog,
    CreateOrEditImPlatformDialog,
    CreateOrEditNotificationTargetDialog,
    CreateOrEditRegistryAuthDialog,
    CreateOrEditSSHKeyDialog,
    CreateOrEditSslCertDialog,
    ProvisionGithubAppDialog,
    UpdateAccessTokenStatusDialog,
    UpdateBasicAuthStatusDialog,
    UpdateCloudStorageStatusDialog,
    UpdateEmailAccountStatusDialog,
    UpdateGithubAppStatusDialog,
    UpdateImPlatformStatusDialog,
    UpdateNotificationTargetStatusDialog,
    UpdateRegistryAuthStatusDialog,
    UpdateSSHKeyStatusDialog,
    UpdateSslCertStatusDialog,
    useCreateOrEditAccessTokenDialogState,
    useCreateOrEditBasicAuthDialogState,
    useCreateOrEditCloudStorageDialogState,
    useCreateOrEditEmailAccountDialogState,
    useCreateOrEditGithubAppDialogState,
    useCreateOrEditImPlatformDialogState,
    useCreateOrEditNotificationTargetDialogState,
    useCreateOrEditRegistryAuthDialogState,
    useCreateOrEditSSHKeyDialogState,
    useCreateOrEditSslCertDialogState,
    useProvisionGithubAppDialogState,
    useUpdateAccessTokenStatusDialogState,
    useUpdateBasicAuthStatusDialogState,
    useUpdateCloudStorageStatusDialogState,
    useUpdateEmailAccountStatusDialogState,
    useUpdateGithubAppStatusDialogState,
    useUpdateImPlatformStatusDialogState,
    useUpdateNotificationTargetStatusDialogState,
    useUpdateRegistryAuthStatusDialogState,
    useUpdateSSHKeyStatusDialogState,
    useUpdateSslCertStatusDialogState,
} from "~/settings/dialogs";

function View() {
    const location = useLocation();
    const createProjectDialog = useCreateProjectDialogState();
    const createProjectAppDialog = useCreateProjectAppDialogState();
    const createOrEditProjectSecretDialog = useCreateOrEditProjectSecretDialogState();
    const createOrEditAppConfigFileDialog = useCreateOrEditAppConfigFileDialogState();
    const createOrEditAppHealthCheckDialog = useCreateOrEditAppHealthCheckDialogState();
    const updateAppHealthCheckStatusDialog = useUpdateAppHealthCheckStatusDialogState();
    const createOrEditAppSecretDialog = useCreateOrEditAppSecretDialogState();
    const quickInstallSslCertDialog = useQuickInstallSslCertDialogState();
    const storageMountDialog = useStorageMountDialogState();
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
    const createOrEditNotificationTargetDialog = useCreateOrEditNotificationTargetDialogState();
    const updateNotificationTargetStatusDialog = useUpdateNotificationTargetStatusDialogState();
    const projectUserAccessesDialog = useProjectUserAccessesDialogState();
    const createOrEditGithubAppDialog = useCreateOrEditGithubAppDialogState();
    const provisionGithubAppDialog = useProvisionGithubAppDialogState();
    const updateGithubAppStatusDialog = useUpdateGithubAppStatusDialogState();
    const importProjectSettingsDialog = useImportProjectSettingsDialogState();

    useUpdateEffect(() => {
        createProjectDialog.destroy();
        createProjectAppDialog.destroy();
        createOrEditProjectSecretDialog.destroy();
        createOrEditAppConfigFileDialog.destroy();
        createOrEditAppHealthCheckDialog.destroy();
        updateAppHealthCheckStatusDialog.destroy();
        createOrEditAppSecretDialog.destroy();
        quickInstallSslCertDialog.destroy();
        storageMountDialog.destroy();
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
        createOrEditNotificationTargetDialog.destroy();
        updateNotificationTargetStatusDialog.destroy();
        projectUserAccessesDialog.destroy();
        createOrEditGithubAppDialog.destroy();
        provisionGithubAppDialog.destroy();
        updateGithubAppStatusDialog.destroy();
        importProjectSettingsDialog.destroy();
    }, [location]);

    return (
        <>
            <CreateProjectDialog />
            <CreateProjectAppDialog />
            <CreateOrEditProjectSecretDialog />
            <CreateOrEditAppConfigFileDialog />
            <CreateOrEditAppHealthCheckDialog />
            <UpdateAppHealthCheckStatusDialog />
            <CreateOrEditAppSecretDialog />
            <QuickInstallSslCertDialog />
            <StorageMountDialog />
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
            <CreateOrEditNotificationTargetDialog />
            <UpdateNotificationTargetStatusDialog />
            <ProjectUserAccessesDialog />
            <CreateOrEditGithubAppDialog />
            <ProvisionGithubAppDialog />
            <UpdateGithubAppStatusDialog />
            <ImportProjectSettingsDialog />
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const ProjectsDialogsContainer = React.memo(View);
