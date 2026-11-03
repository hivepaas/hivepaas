import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

import type { SystemSslRenewalNotification } from "./system-ssl-renewal-settings.entity";

/** An interval only, from 1 to 10 hours: never a cron. */
export interface SystemRegistryAuthRenewalSchedule {
    interval: string;
    initialTime?: Date | null;
}

export interface SystemRegistryAuthRenewalSettings extends SettingsBaseEntity {
    status: OpenApiConstant<ESettingStatus>;
    schedule: SystemRegistryAuthRenewalSchedule;
    notification?: SystemSslRenewalNotification | null;
    nextRuns: Date[];
}
