import type { SystemCleanupSettings } from "~/system-settings/domain";
import { getScheduleModeFromCronExpr } from "~/system-settings/module-shared";

import { ESettingStatus } from "@application/shared/enums";

import { type SystemCleanupConfigurationFormInput, SystemCleanupScheduleMode } from "../schemas";

export const emptySystemCleanupConfigurationFormDefaults: SystemCleanupConfigurationFormInput = {
    status: ESettingStatus.Active,
    scheduleMode: SystemCleanupScheduleMode.Interval,
    scheduleInterval: "24h",
    scheduleCronExpr: "",
    scheduleFrom: null,
    dbObjectRetention: {
        enabled: true,
        tasks: "180d",
        deployments: "180d",
        sysErrors: "180d",
        deletedObjects: "180d",
    },
    clusterCleanup: {
        enabled: true,
        pruneImages: true,
        pruneVolumes: true,
        pruneNetworks: true,
        pruneContainers: true,
    },
    backupCleanup: {
        enabled: true,
        localBackupRetention: "30d",
        cloudBackupRetention: "30d",
    },
    notification: {
        successUseDefault: true,
        success: undefined,
        failureUseDefault: true,
        failure: undefined,
    },
};

export function mapSystemCleanupSettingsToFormInput(
    settings: SystemCleanupSettings,
): SystemCleanupConfigurationFormInput {
    return {
        status: settings.status === ESettingStatus.Active ? ESettingStatus.Active : ESettingStatus.Disabled,
        scheduleMode: getScheduleModeFromCronExpr(settings.schedule.cronExpr, SystemCleanupScheduleMode),
        scheduleInterval: settings.schedule.interval,
        scheduleCronExpr: settings.schedule.cronExpr,
        scheduleFrom: settings.schedule.initialTime ?? null,
        dbObjectRetention: {
            enabled: settings.dbObjectRetention.enabled,
            tasks: settings.dbObjectRetention.tasks,
            deployments: settings.dbObjectRetention.deployments,
            sysErrors: settings.dbObjectRetention.sysErrors,
            deletedObjects: settings.dbObjectRetention.deletedObjects,
        },
        clusterCleanup: {
            enabled: settings.clusterCleanup.enabled,
            pruneImages: settings.clusterCleanup.pruneImages,
            pruneVolumes: settings.clusterCleanup.pruneVolumes,
            pruneNetworks: settings.clusterCleanup.pruneNetworks,
            pruneContainers: settings.clusterCleanup.pruneContainers,
        },
        backupCleanup: {
            enabled: settings.backupCleanup.enabled,
            localBackupRetention: settings.backupCleanup.localBackupRetention,
            cloudBackupRetention: settings.backupCleanup.cloudBackupRetention,
        },
        notification: {
            successUseDefault: settings.notification?.successUseDefault ?? true,
            success: settings.notification?.success,
            failureUseDefault: settings.notification?.failureUseDefault ?? true,
            failure: settings.notification?.failure,
        },
    };
}
