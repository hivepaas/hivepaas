import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface SystemBackupSchedule {
    interval: string;
    cronExpr: string;
    initialTime?: Date | null;
}

export interface SystemBackupNotification {
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

/** How a system backup's spec holds secrets, as a spec export's modes are. */
export const SYSTEM_BACKUP_SPEC_SECRETS = {
    Encrypted: "encrypted",
    Omit: "omit",
    Plaintext: "plaintext",
} as const;

export type SystemBackupSpecSecrets = (typeof SYSTEM_BACKUP_SPEC_SECRETS)[keyof typeof SYSTEM_BACKUP_SPEC_SECRETS];

export interface SystemBackupSettings extends SettingsBaseEntity {
    status: OpenApiConstant<ESettingStatus>;
    schedule: SystemBackupSchedule;
    /** What a run takes: HivePaaS's database, the spec of the whole installation, or both. */
    includeDB: boolean;
    includeSpec: boolean;
    /** How the spec holds secrets: encrypted, omit, plaintext. */
    specSecrets: SystemBackupSpecSecrets;
    /** Masked once stored. */
    specPassphrase: string;
    /** A backup repository at the global scope; its status is "missing" when it is deleted. */
    targetRepository?: { id: string; name: string; status: string } | null;
    notification?: SystemBackupNotification | null;
    secretMasked?: boolean;
    nextRuns: Date[];
}
