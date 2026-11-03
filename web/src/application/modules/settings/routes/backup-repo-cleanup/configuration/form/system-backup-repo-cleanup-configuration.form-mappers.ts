import type { SystemBackupRepoCleanupSettings } from "~/system-settings/domain";
import { getScheduleModeFromCronExpr } from "~/system-settings/module-shared";

import { ESettingStatus } from "@application/shared/enums";

import { type SystemBackupRepoCleanupConfigurationFormInput, SystemBackupRepoCleanupScheduleMode } from "../schemas";

export const emptySystemBackupRepoCleanupConfigurationFormDefaults: SystemBackupRepoCleanupConfigurationFormInput = {
    status: ESettingStatus.Active,
    scheduleMode: SystemBackupRepoCleanupScheduleMode.Interval,
    scheduleInterval: "24h",
    scheduleCronExpr: "",
    scheduleFrom: null,
    notification: {
        successUseDefault: true,
        success: undefined,
        failureUseDefault: true,
        failure: undefined,
    },
};

export function mapSystemBackupRepoCleanupSettingsToFormInput(
    settings: SystemBackupRepoCleanupSettings,
): SystemBackupRepoCleanupConfigurationFormInput {
    return {
        status: settings.status === ESettingStatus.Active ? ESettingStatus.Active : ESettingStatus.Disabled,
        scheduleMode: getScheduleModeFromCronExpr(settings.schedule.cronExpr, SystemBackupRepoCleanupScheduleMode),
        scheduleInterval: settings.schedule.interval,
        scheduleCronExpr: settings.schedule.cronExpr,
        scheduleFrom: settings.schedule.initialTime ?? null,
        notification: {
            successUseDefault: settings.notification?.successUseDefault ?? true,
            success: settings.notification?.success,
            failureUseDefault: settings.notification?.failureUseDefault ?? true,
            failure: settings.notification?.failure,
        },
    };
}
