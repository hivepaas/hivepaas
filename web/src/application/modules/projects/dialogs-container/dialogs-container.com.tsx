import React from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    CreateOrEditAppConfigFileDialog,
    useCreateOrEditAppConfigFileDialogState,
} from "~/projects/dialogs/create-or-edit-app-config-file";
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
    QuickInstallSslCertDialog,
    useQuickInstallSslCertDialogState,
} from "~/projects/dialogs/quick-install-ssl-cert";
import { StorageMountDialog, useStorageMountDialogState } from "~/projects/dialogs/storage-mount";
import {
    CreateOrEditAccessTokenDialog,
    CreateOrEditBasicAuthDialog,
    CreateOrEditCloudStorageDialog,
    CreateOrEditEmailAccountDialog,
    CreateOrEditImPlatformDialog,
    CreateOrEditRegistryAuthDialog,
    CreateOrEditSSHKeyDialog,
    CreateOrEditSslCertDialog,
    UpdateAccessTokenStatusDialog,
    UpdateBasicAuthStatusDialog,
    UpdateCloudStorageStatusDialog,
    UpdateEmailAccountStatusDialog,
    UpdateImPlatformStatusDialog,
    UpdateRegistryAuthStatusDialog,
    UpdateSSHKeyStatusDialog,
    UpdateSslCertStatusDialog,
    useCreateOrEditAccessTokenDialogState,
    useCreateOrEditBasicAuthDialogState,
    useCreateOrEditCloudStorageDialogState,
    useCreateOrEditEmailAccountDialogState,
    useCreateOrEditImPlatformDialogState,
    useCreateOrEditRegistryAuthDialogState,
    useCreateOrEditSSHKeyDialogState,
    useCreateOrEditSslCertDialogState,
    useUpdateAccessTokenStatusDialogState,
    useUpdateBasicAuthStatusDialogState,
    useUpdateCloudStorageStatusDialogState,
    useUpdateEmailAccountStatusDialogState,
    useUpdateImPlatformStatusDialogState,
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

    useUpdateEffect(() => {
        createProjectDialog.destroy();
        createProjectAppDialog.destroy();
        createOrEditProjectSecretDialog.destroy();
        createOrEditAppConfigFileDialog.destroy();
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
    }, [location]);

    return (
        <>
            <CreateProjectDialog />
            <CreateProjectAppDialog />
            <CreateOrEditProjectSecretDialog />
            <CreateOrEditAppConfigFileDialog />
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
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const ProjectsDialogsContainer = React.memo(View);
