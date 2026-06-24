import React from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    CreateNetworkDialog,
    ViewNetworkDialog,
    useCreateNetworkDialogState,
    useViewNetworkDialogState,
} from "~/cluster/dialogs";
import {
    ConfirmAppDangerActionDialog,
    useConfirmAppDangerActionDialogState,
} from "~/projects/dialogs/confirm-app-danger-action";
import {
    ConfirmProjectDangerActionDialog,
    useConfirmProjectDangerActionDialogState,
} from "~/projects/dialogs/confirm-project-danger-action";
import { CopyProjectAppDialog, useCopyProjectAppDialogState } from "~/projects/dialogs/copy-project-app";
import {
    CreateOrEditAppConfigFileDialog,
    useCreateOrEditAppConfigFileDialogState,
} from "~/projects/dialogs/create-or-edit-app-config-file";
import {
    CreateOrEditAppHealthCheckDialog,
    useCreateOrEditAppHealthCheckDialogState,
} from "~/projects/dialogs/create-or-edit-app-health-check";
import {
    CreateOrEditAppScheduledJobDialog,
    useCreateOrEditAppScheduledJobDialogState,
} from "~/projects/dialogs/create-or-edit-app-scheduled-job";
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
import { RunNowTaskCreatedDialog, useRunNowTaskCreatedDialogState } from "~/projects/dialogs/run-now-task-created";
import { StorageMountDialog, useStorageMountDialogState } from "~/projects/dialogs/storage-mount";
import {
    UpdateAppHealthCheckStatusDialog,
    useUpdateAppHealthCheckStatusDialogState,
} from "~/projects/dialogs/update-app-health-check-status";
import {
    UpdateAppScheduledJobStatusDialog,
    useUpdateAppScheduledJobStatusDialogState,
} from "~/projects/dialogs/update-app-scheduled-job-status";
import {
    CreateOrEditGithubAppDialog,
    CreateOrEditRepoWebhookDialog,
    ProvisionGithubAppDialog,
    UpdateAccessTokenStatusDialog,
    UpdateAcmeDnsProviderStatusDialog,
    UpdateBasicAuthStatusDialog,
    UpdateCloudStorageStatusDialog,
    UpdateEmailAccountStatusDialog,
    UpdateGithubAppStatusDialog,
    UpdateImPlatformStatusDialog,
    UpdateNotificationTargetStatusDialog,
    UpdateRegistryAuthStatusDialog,
    UpdateRepoWebhookStatusDialog,
    UpdateSSHKeyStatusDialog,
    UpdateSslCertStatusDialog,
    UpdateSslProviderStatusDialog,
    useCreateOrEditGithubAppDialogState,
    useCreateOrEditRepoWebhookDialogState,
    useProvisionGithubAppDialogState,
    useUpdateAccessTokenStatusDialogState,
    useUpdateAcmeDnsProviderStatusDialogState,
    useUpdateBasicAuthStatusDialogState,
    useUpdateCloudStorageStatusDialogState,
    useUpdateEmailAccountStatusDialogState,
    useUpdateGithubAppStatusDialogState,
    useUpdateImPlatformStatusDialogState,
    useUpdateNotificationTargetStatusDialogState,
    useUpdateRegistryAuthStatusDialogState,
    useUpdateRepoWebhookStatusDialogState,
    useUpdateSSHKeyStatusDialogState,
    useUpdateSslCertStatusDialogState,
    useUpdateSslProviderStatusDialogState,
} from "~/settings/dialogs";

function View() {
    const location = useLocation();
    const createProjectDialog = useCreateProjectDialogState();
    const createProjectAppDialog = useCreateProjectAppDialogState();
    const copyProjectAppDialog = useCopyProjectAppDialogState();
    const createOrEditProjectSecretDialog = useCreateOrEditProjectSecretDialogState();
    const createOrEditAppConfigFileDialog = useCreateOrEditAppConfigFileDialogState();
    const createOrEditAppHealthCheckDialog = useCreateOrEditAppHealthCheckDialogState();
    const createOrEditAppScheduledJobDialog = useCreateOrEditAppScheduledJobDialogState();
    const updateAppHealthCheckStatusDialog = useUpdateAppHealthCheckStatusDialogState();
    const updateAppScheduledJobStatusDialog = useUpdateAppScheduledJobStatusDialogState();
    const runNowTaskCreatedDialog = useRunNowTaskCreatedDialogState();
    const createOrEditAppSecretDialog = useCreateOrEditAppSecretDialogState();
    const quickInstallSslCertDialog = useQuickInstallSslCertDialogState();
    const storageMountDialog = useStorageMountDialogState();
    const confirmAppDangerActionDialog = useConfirmAppDangerActionDialogState();
    const confirmProjectDangerActionDialog = useConfirmProjectDangerActionDialogState();
    const createNetworkDialog = useCreateNetworkDialogState();
    const viewNetworkDialog = useViewNetworkDialogState();
    const updateBasicAuthStatusDialog = useUpdateBasicAuthStatusDialogState();
    const updateRegistryAuthStatusDialog = useUpdateRegistryAuthStatusDialogState();
    const updateSslCertStatusDialog = useUpdateSslCertStatusDialogState();
    const updateSslProviderStatusDialog = useUpdateSslProviderStatusDialogState();
    const updateImPlatformStatusDialog = useUpdateImPlatformStatusDialogState();
    const updateEmailAccountStatusDialog = useUpdateEmailAccountStatusDialogState();
    const updateSSHKeyStatusDialog = useUpdateSSHKeyStatusDialogState();
    const updateAccessTokenStatusDialog = useUpdateAccessTokenStatusDialogState();
    const updateAcmeDnsProviderStatusDialog = useUpdateAcmeDnsProviderStatusDialogState();
    const updateCloudStorageStatusDialog = useUpdateCloudStorageStatusDialogState();
    const updateNotificationTargetStatusDialog = useUpdateNotificationTargetStatusDialogState();
    const projectUserAccessesDialog = useProjectUserAccessesDialogState();
    const createOrEditGithubAppDialog = useCreateOrEditGithubAppDialogState();
    const provisionGithubAppDialog = useProvisionGithubAppDialogState();
    const updateGithubAppStatusDialog = useUpdateGithubAppStatusDialogState();
    const createOrEditRepoWebhookDialog = useCreateOrEditRepoWebhookDialogState();
    const updateRepoWebhookStatusDialog = useUpdateRepoWebhookStatusDialogState();
    const importProjectSettingsDialog = useImportProjectSettingsDialogState();

    useUpdateEffect(() => {
        createProjectDialog.destroy();
        createProjectAppDialog.destroy();
        copyProjectAppDialog.destroy();
        createOrEditProjectSecretDialog.destroy();
        createOrEditAppConfigFileDialog.destroy();
        createOrEditAppHealthCheckDialog.destroy();
        createOrEditAppScheduledJobDialog.destroy();
        updateAppHealthCheckStatusDialog.destroy();
        updateAppScheduledJobStatusDialog.destroy();
        runNowTaskCreatedDialog.destroy();
        createOrEditAppSecretDialog.destroy();
        quickInstallSslCertDialog.destroy();
        storageMountDialog.destroy();
        confirmAppDangerActionDialog.destroy();
        confirmProjectDangerActionDialog.destroy();
        createNetworkDialog.destroy();
        viewNetworkDialog.destroy();
        updateBasicAuthStatusDialog.destroy();
        updateRegistryAuthStatusDialog.destroy();
        updateSslCertStatusDialog.destroy();
        updateSslProviderStatusDialog.destroy();
        updateImPlatformStatusDialog.destroy();
        updateEmailAccountStatusDialog.destroy();
        updateSSHKeyStatusDialog.destroy();
        updateAccessTokenStatusDialog.destroy();
        updateAcmeDnsProviderStatusDialog.destroy();
        updateCloudStorageStatusDialog.destroy();
        updateNotificationTargetStatusDialog.destroy();
        projectUserAccessesDialog.destroy();
        createOrEditGithubAppDialog.destroy();
        provisionGithubAppDialog.destroy();
        updateGithubAppStatusDialog.destroy();
        createOrEditRepoWebhookDialog.destroy();
        updateRepoWebhookStatusDialog.destroy();
        importProjectSettingsDialog.destroy();
    }, [location]);

    return (
        <>
            <CreateProjectDialog />
            <CreateProjectAppDialog />
            <CopyProjectAppDialog />
            <CreateOrEditProjectSecretDialog />
            <CreateOrEditAppConfigFileDialog />
            <CreateOrEditAppHealthCheckDialog />
            <CreateOrEditAppScheduledJobDialog />
            <UpdateAppHealthCheckStatusDialog />
            <UpdateAppScheduledJobStatusDialog />
            <RunNowTaskCreatedDialog />
            <CreateOrEditAppSecretDialog />
            <QuickInstallSslCertDialog />
            <StorageMountDialog />
            <ConfirmAppDangerActionDialog />
            <ConfirmProjectDangerActionDialog />
            <CreateNetworkDialog />
            <ViewNetworkDialog />
            <UpdateBasicAuthStatusDialog />
            <UpdateRegistryAuthStatusDialog />
            <UpdateSslCertStatusDialog />
            <UpdateSslProviderStatusDialog />
            <UpdateImPlatformStatusDialog />
            <UpdateEmailAccountStatusDialog />
            <UpdateSSHKeyStatusDialog />
            <UpdateAccessTokenStatusDialog />
            <UpdateAcmeDnsProviderStatusDialog />
            <UpdateCloudStorageStatusDialog />
            <UpdateNotificationTargetStatusDialog />
            <ProjectUserAccessesDialog />
            <CreateOrEditGithubAppDialog />
            <ProvisionGithubAppDialog />
            <UpdateGithubAppStatusDialog />
            <CreateOrEditRepoWebhookDialog />
            <UpdateRepoWebhookStatusDialog />
            <ImportProjectSettingsDialog />
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const ProjectsDialogsContainer = React.memo(View);
