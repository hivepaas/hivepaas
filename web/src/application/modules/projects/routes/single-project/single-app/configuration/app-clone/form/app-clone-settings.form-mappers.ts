import type { AppCloneSettings_UpdatePayload } from "~/projects/api/services";
import type { AppCloneSettings } from "~/projects/domain";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

import type { AppCloneSettingsFormSchemaInput, AppCloneSettingsFormSchemaOutput } from "../schemas";

function normalizeTargetStatus(
    status: AppCloneSettings["targetStatus"],
): AppCloneSettingsFormSchemaInput["targetStatus"] {
    return status === EProjectAppStatus.Disabled ? EProjectAppStatus.Disabled : EProjectAppStatus.Active;
}

function buildNotificationPayload(
    notification: AppCloneSettingsFormSchemaOutput["notification"],
): AppCloneSettings_UpdatePayload["notification"] {
    return {
        successUseDefault: notification.successUseDefault,
        ...(!notification.successUseDefault && notification.success?.id
            ? { success: { id: notification.success.id } }
            : {}),
        failureUseDefault: notification.failureUseDefault,
        ...(!notification.failureUseDefault && notification.failure?.id
            ? { failure: { id: notification.failure.id } }
            : {}),
    };
}

export function mapAppCloneSettingsToFormInput(data: AppCloneSettings): AppCloneSettingsFormSchemaInput {
    return {
        targetName: data.targetName,
        targetEnv: data.targetEnv,
        targetStatus: normalizeTargetStatus(data.targetStatus),
        targetReplicas: data.targetReplicas,
        cloneDeploymentSettings: data.cloneDeploymentSettings,
        cloneHttpSettings: data.cloneHttpSettings,
        cloneHttpDomains: data.cloneHttpDomains.map(domain => ({
            sourceDomain: domain.sourceDomain,
            targetDomain: domain.targetDomain,
            sourceSslCert: domain.sourceSslCert ?? null,
            targetSslCert: domain.targetSslCert ?? null,
        })),
        cloneVolumes: data.cloneVolumes,
        cloneVolumeData: data.cloneVolumeData,
        stopSourceAppBeforeClone: !data.liveVolumeClone,
        cloneEnvVars: data.cloneEnvVars,
        cloneSecrets: data.cloneSecrets,
        cloneConfigFiles: data.cloneConfigFiles,
        clonePeriodicJobs: data.clonePeriodicJobs,
        cloneSchedJobs: data.cloneSchedJobs,
        commandPipes: data.commandPipes.map(pipe => ({
            id: pipe.id,
            name: pipe.name,
        })),
        includedVolumes: data.includedVolumes,
        excludedVolumes: data.excludedVolumes,
        notification: {
            successUseDefault: data.notification?.successUseDefault ?? true,
            success: data.notification?.success ?? null,
            failureUseDefault: data.notification?.failureUseDefault ?? true,
            failure: data.notification?.failure ?? null,
        },
        updateVer: data.updateVer,
    };
}

export function mapFormToUpdatePayload(
    values: AppCloneSettingsFormSchemaOutput,
    server: AppCloneSettings,
): AppCloneSettings_UpdatePayload {
    return {
        updateVer: server.updateVer,
        targetName: values.targetName,
        targetEnv: values.targetEnv,
        targetStatus: values.targetStatus,
        targetReplicas: values.targetReplicas,
        cloneDeploymentSettings: values.cloneDeploymentSettings,
        cloneHttpSettings: values.cloneHttpSettings,
        cloneHttpDomains: values.cloneHttpDomains
            .filter(domain => domain.targetDomain.trim() !== "")
            .map(domain => ({
                sourceDomain: domain.sourceDomain,
                targetDomain: domain.targetDomain.trim(),
                sourceSslCert: { id: domain.sourceSslCert?.id ?? "" },
                targetSslCert: { id: domain.targetSslCert?.id ?? "" },
            })),
        cloneVolumes: values.cloneVolumes,
        cloneVolumeData: values.cloneVolumeData,
        liveVolumeClone: !values.stopSourceAppBeforeClone,
        includedVolumes: server.includedVolumes,
        excludedVolumes: server.excludedVolumes,
        cloneEnvVars: values.cloneEnvVars,
        cloneSecrets: values.cloneSecrets,
        cloneConfigFiles: values.cloneConfigFiles,
        clonePeriodicJobs: values.clonePeriodicJobs,
        cloneSchedJobs: values.cloneSchedJobs,
        commandPipes: values.commandPipes.map(pipe => ({ id: pipe.id })),
        notification: buildNotificationPayload(values.notification),
    };
}
