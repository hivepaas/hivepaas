import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

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
    status: OpenApiConstant<ESettingStatus>;
    schedule: SystemSslRenewalSchedule;
    notification?: SystemSslRenewalNotification | null;
    nextRuns: Date[];
}
