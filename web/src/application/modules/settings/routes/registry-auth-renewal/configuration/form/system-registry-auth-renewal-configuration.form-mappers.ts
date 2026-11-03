import type { SystemRegistryAuthRenewalSettings } from "~/system-settings/domain";

import { ESettingStatus } from "@application/shared/enums";

import type { SystemRegistryAuthRenewalConfigurationFormInput } from "../schemas";

export const emptySystemRegistryAuthRenewalConfigurationFormDefaults: SystemRegistryAuthRenewalConfigurationFormInput =
    {
        status: ESettingStatus.Active,
        scheduleInterval: "6h",
        scheduleFrom: null,
        notification: {
            successUseDefault: false,
            success: undefined,
            failureUseDefault: true,
            failure: undefined,
        },
    };

export function mapSystemRegistryAuthRenewalSettingsToFormInput(
    settings: SystemRegistryAuthRenewalSettings,
): SystemRegistryAuthRenewalConfigurationFormInput {
    return {
        status: settings.status === ESettingStatus.Active ? ESettingStatus.Active : ESettingStatus.Disabled,
        scheduleInterval: settings.schedule.interval || "6h",
        scheduleFrom: settings.schedule.initialTime ?? null,
        notification: {
            successUseDefault: settings.notification?.successUseDefault ?? false,
            success: settings.notification?.success,
            failureUseDefault: settings.notification?.failureUseDefault ?? true,
            failure: settings.notification?.failure,
        },
    };
}
