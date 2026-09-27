import type { SystemSslRenewalSettings } from "~/system-settings/domain";
import { getScheduleModeFromCronExpr } from "~/system-settings/module-shared";

import { ESettingStatus } from "@application/shared/enums";

import { type SystemSslRenewalConfigurationFormInput, SystemSslRenewalScheduleMode } from "../schemas";

export const emptySystemSslRenewalConfigurationFormDefaults: SystemSslRenewalConfigurationFormInput = {
    status: ESettingStatus.Active,
    scheduleMode: SystemSslRenewalScheduleMode.Interval,
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

export function mapSystemSslRenewalSettingsToFormInput(
    settings: SystemSslRenewalSettings,
): SystemSslRenewalConfigurationFormInput {
    return {
        status: settings.status === ESettingStatus.Active ? ESettingStatus.Active : ESettingStatus.Disabled,
        scheduleMode: getScheduleModeFromCronExpr(settings.schedule.cronExpr, SystemSslRenewalScheduleMode),
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
