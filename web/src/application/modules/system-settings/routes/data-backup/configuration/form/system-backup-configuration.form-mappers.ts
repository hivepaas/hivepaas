import type { SystemBackupSettings } from "~/system-settings/domain";

import { ESettingStatus } from "@application/shared/enums";

import { ESystemBackupCompressionFormat, ESystemBackupEncryptionFormat } from "../../../../module-shared/enums";
import type { SystemBackupConfigurationFormInput } from "../schemas";

export const emptySystemBackupConfigurationFormDefaults: SystemBackupConfigurationFormInput = {
    status: ESettingStatus.Disabled,
    scheduleInterval: "24h",
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

export function mapSystemBackupSettingsToFormInput(
    settings: SystemBackupSettings,
): SystemBackupConfigurationFormInput {
    return {
        status: settings.status === ESettingStatus.Active ? ESettingStatus.Active : ESettingStatus.Disabled,
        scheduleInterval: settings.scheduleInterval,
        scheduleFrom: settings.scheduleFrom,
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
