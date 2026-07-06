import React from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    UpdateNetworkStatusDialog,
    UpdateVolumeStatusDialog,
    useUpdateNetworkStatusDialogState,
    useUpdateVolumeStatusDialogState,
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
import {
    UpdateAppHealthCheckStatusDialog,
    useUpdateAppHealthCheckStatusDialogState,
} from "~/projects/dialogs/update-app-health-check-status";
import {
    UpdateAppScheduledJobStatusDialog,
    useUpdateAppScheduledJobStatusDialogState,
} from "~/projects/dialogs/update-app-scheduled-job-status";
import {
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
    const updateAppHealthCheckStatusDialog = useUpdateAppHealthCheckStatusDialogState();
    const updateAppScheduledJobStatusDialog = useUpdateAppScheduledJobStatusDialogState();
    const runNowTaskCreatedDialog = useRunNowTaskCreatedDialogState();
    const quickInstallSslCertDialog = useQuickInstallSslCertDialogState();
    const confirmAppDangerActionDialog = useConfirmAppDangerActionDialogState();
    const confirmProjectDangerActionDialog = useConfirmProjectDangerActionDialogState();
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
    const provisionGithubAppDialog = useProvisionGithubAppDialogState();
    const updateGithubAppStatusDialog = useUpdateGithubAppStatusDialogState();
    const updateRepoWebhookStatusDialog = useUpdateRepoWebhookStatusDialogState();
    const importProjectSettingsDialog = useImportProjectSettingsDialogState();
    const updateNetworkStatusDialog = useUpdateNetworkStatusDialogState();
    const updateVolumeStatusDialog = useUpdateVolumeStatusDialogState();

    useUpdateEffect(() => {
        createProjectDialog.destroy();
        createProjectAppDialog.destroy();
        copyProjectAppDialog.destroy();
        updateAppHealthCheckStatusDialog.destroy();
        updateAppScheduledJobStatusDialog.destroy();
        runNowTaskCreatedDialog.destroy();
        quickInstallSslCertDialog.destroy();
        confirmAppDangerActionDialog.destroy();
        confirmProjectDangerActionDialog.destroy();
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
        provisionGithubAppDialog.destroy();
        updateGithubAppStatusDialog.destroy();
        updateRepoWebhookStatusDialog.destroy();
        importProjectSettingsDialog.destroy();
        updateNetworkStatusDialog.destroy();
        updateVolumeStatusDialog.destroy();
    }, [location]);

    return (
        <>
            <CreateProjectDialog />
            <CreateProjectAppDialog />
            <CopyProjectAppDialog />
            <UpdateAppHealthCheckStatusDialog />
            <UpdateAppScheduledJobStatusDialog />
            <RunNowTaskCreatedDialog />
            <QuickInstallSslCertDialog />
            <ConfirmAppDangerActionDialog />
            <ConfirmProjectDangerActionDialog />
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
            <ProvisionGithubAppDialog />
            <UpdateGithubAppStatusDialog />
            <UpdateRepoWebhookStatusDialog />
            <ImportProjectSettingsDialog />
            <UpdateNetworkStatusDialog />
            <UpdateVolumeStatusDialog />
            {/* TODO: Add other dialogs here */}
        </>
    );
}

export const ProjectsDialogsContainer = React.memo(View);
