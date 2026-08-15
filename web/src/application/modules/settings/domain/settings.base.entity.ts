import type { ESettingStatus } from "@application/shared/enums";

export interface SettingsBaseEntity {
    id: string;
    type: string;
    name: string;
    kind?: string;
    status: ESettingStatus;
    inherited?: boolean;
    availableInProjects?: boolean;
    default?: boolean;
    updateVer: number;
    createdAt: Date;
    updatedAt?: Date | null;
    expireAt?: Date | null;
}
