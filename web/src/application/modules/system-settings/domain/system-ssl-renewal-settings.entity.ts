import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESettingStatus, ESettingType } from "@application/shared/enums";

export interface SystemSslRenewalSchedule {
    interval: string;
    cronExpr: string;
    initialTime?: Date | null;
}

export interface SystemSslRenewalNotification {
    success?: {
        id: string;
        name: string;
    };
    successUseDefault: boolean;
    failure?: {
        id: string;
        name: string;
    };
    failureUseDefault: boolean;
}

export interface SystemSslRenewalSettings extends SettingsBaseEntity {
    type: typeof ESettingType.SSLRenewal;
    status: ESettingStatus;
    schedule: SystemSslRenewalSchedule;
    notification?: SystemSslRenewalNotification | null;
    nextRuns: Date[];
}
