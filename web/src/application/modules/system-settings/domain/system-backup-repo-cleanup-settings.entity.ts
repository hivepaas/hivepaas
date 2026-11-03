import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface SystemBackupRepoCleanupSchedule {
    interval: string;
    cronExpr: string;
    initialTime?: Date | null;
}

export interface SystemBackupRepoCleanupNotification {
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

export interface SystemBackupRepoCleanupSettings extends SettingsBaseEntity {
    status: OpenApiConstant<ESettingStatus>;
    schedule: SystemBackupRepoCleanupSchedule;
    notification?: SystemBackupRepoCleanupNotification | null;
    nextRuns: Date[];
}
