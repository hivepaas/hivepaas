import type { ESettingStatus, ESettingType } from "../../enums";

export interface SettingsBaseEntity {
    id: string;
    type: ESettingType;
    name: string;
    kind?: string;
    status: ESettingStatus;
    inherited?: boolean;
    availableInProjects?: boolean;
    default?: boolean;
    updateVer: number;
    createdAt: string;
    updatedAt: string;
    expireAt?: string;
}
