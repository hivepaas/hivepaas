import type { SystemBackupSettings } from "~/system-settings/domain";

import { ESettingStatus } from "@application/shared/enums";

import { ESystemBackupCompressionFormat, ESystemBackupEncryptionFormat } from "../../../../module-shared/enums";
import { SystemBackupScheduleMode, type SystemBackupConfigurationFormInput } from "../schemas";

export const emptySystemBackupConfigurationFormDefaults: SystemBackupConfigurationFormInput = {
    status: ESettingStatus.Disabled,
    scheduleMode: SystemBackupScheduleMode.Interval,
    scheduleInterval: "24h",
    scheduleCronExpr: "",
    scheduleFrom: new Date(),
    compressionFormat: ESystemBackupCompressionFormat.Gzip,
    encryptionFormat: ESystemBackupEncryptionFormat.None,
    encryptionSecret: "",
    cloudStorage: undefined,
    cloudStorageDestinationDir: "",
    backupDeletedObjects: false,
    notification: {
        successUseDefault: true,
        success: undefined,
        failureUseDefault: true,
        failure: undefined,
    },
};

export function mapSystemBackupSettingsToFormInput(settings: SystemBackupSettings): SystemBackupConfigurationFormInput {
    const hasCronSchedule = settings.schedule.cronExpr.trim().length > 0;

    return {
        status: settings.status === ESettingStatus.Active ? ESettingStatus.Active : ESettingStatus.Disabled,
        scheduleMode: hasCronSchedule ? SystemBackupScheduleMode.Cron : SystemBackupScheduleMode.Interval,
        scheduleInterval: settings.schedule.interval,
        scheduleCronExpr: settings.schedule.cronExpr,
        scheduleFrom: settings.schedule.initialTime ?? null,
        compressionFormat: settings.compression.format,
        encryptionFormat: settings.encryption.format,
        encryptionSecret: settings.encryption.secret,
        cloudStorage: settings.cloudStorage
            ? {
                  id: settings.cloudStorage.id,
                  name: settings.cloudStorage.name,
              }
            : undefined,
        cloudStorageDestinationDir: settings.cloudStorage?.destinationDir ?? "",
        backupDeletedObjects: settings.dbBackupConfig.backupDeletedObjects,
        notification: {
            successUseDefault: settings.notification?.successUseDefault ?? true,
            success: settings.notification?.success,
            failureUseDefault: settings.notification?.failureUseDefault ?? true,
            failure: settings.notification?.failure,
        },
    };
}
