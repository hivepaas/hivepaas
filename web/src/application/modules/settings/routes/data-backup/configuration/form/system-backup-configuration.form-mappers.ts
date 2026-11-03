import { SYSTEM_BACKUP_SPEC_SECRETS, type SystemBackupSettings } from "~/system-settings/domain";
import { getScheduleModeFromCronExpr } from "~/system-settings/module-shared";

import { ESettingStatus } from "@application/shared/enums";

import { type SystemBackupConfigurationFormInput, SystemBackupScheduleMode } from "../schemas";

export const emptySystemBackupConfigurationFormDefaults: SystemBackupConfigurationFormInput = {
    status: ESettingStatus.Disabled,
    scheduleMode: SystemBackupScheduleMode.Interval,
    scheduleInterval: "24h",
    scheduleCronExpr: "",
    scheduleFrom: null,
    includeDB: true,
    includeSpec: false,
    specSecrets: SYSTEM_BACKUP_SPEC_SECRETS.Encrypted,
    specPassphrase: "",
    targetRepository: undefined,
    notification: {
        successUseDefault: true,
        success: undefined,
        failureUseDefault: true,
        failure: undefined,
    },
};

export function mapSystemBackupSettingsToFormInput(settings: SystemBackupSettings): SystemBackupConfigurationFormInput {
    return {
        status: settings.status === ESettingStatus.Active ? ESettingStatus.Active : ESettingStatus.Disabled,
        scheduleMode: getScheduleModeFromCronExpr(settings.schedule.cronExpr, SystemBackupScheduleMode),
        scheduleInterval: settings.schedule.interval,
        scheduleCronExpr: settings.schedule.cronExpr,
        scheduleFrom: settings.schedule.initialTime ?? null,
        includeDB: settings.includeDB,
        includeSpec: settings.includeSpec,
        specSecrets: settings.specSecrets,
        specPassphrase: settings.specPassphrase,
        targetRepository: settings.targetRepository
            ? {
                  id: settings.targetRepository.id,
                  name:
                      settings.targetRepository.status === "missing"
                          ? "Deleted repository"
                          : settings.targetRepository.name,
              }
            : undefined,
        notification: {
            successUseDefault: settings.notification?.successUseDefault ?? true,
            success: settings.notification?.success,
            failureUseDefault: settings.notification?.failureUseDefault ?? true,
            failure: settings.notification?.failure,
        },
    };
}
