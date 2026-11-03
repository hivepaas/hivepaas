import type { ESettingStatus } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface ProjectCommandPipeSettingRef {
    id: string;
    type: string;
    name: string;
    kind: string;
    status: OpenApiConstant<ESettingStatus>;
    inherited: boolean;
    inheritable: boolean;
    default: boolean;
    updateVer: number;
    createdAt: Date;
    updatedAt: Date | null;
    expireAt: Date | null;
}

export interface ProjectCommandPipe {
    id: string;
    type: string;
    name: string;
    status: OpenApiConstant<ESettingStatus>;
    inherited: boolean;
    inheritable: boolean;
    default: boolean;
    updateVer: number;
    createdAt: Date;
    updatedAt: Date | null;
    expireAt: Date | null;
    size: number;
    sourceCommand: ProjectCommandPipeSettingRef | null;
    targetCommand: ProjectCommandPipeSettingRef | null;
}
