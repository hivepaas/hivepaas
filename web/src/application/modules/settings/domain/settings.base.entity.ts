import type { ESettingStatus } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface SettingsBaseEntity {
    id: string;
    type: string;
    name: string;
    kind?: string;
    status: OpenApiConstant<ESettingStatus>;
    inherited?: boolean;
    inheritable?: boolean;
    default?: boolean;
    updateVer: number;
    createdAt: Date;
    updatedAt?: Date | null;
    expireAt?: Date | null;
}
