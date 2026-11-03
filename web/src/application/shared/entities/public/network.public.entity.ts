import type { ESettingStatus } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export interface NetworkPublic {
    id: string;
    type: string;
    name: string;
    kind?: string;
    status: OpenApiConstant<ESettingStatus>;
    inherited: boolean;
    inheritable: boolean;
    default: boolean;
    updateVer: number;
    size: number;
    driver: string;
    internal: boolean;
    attachable: boolean;
    ingress: boolean;
    enableIPv4: boolean;
    enableIPv6: boolean;
    options: Record<string, string>;
    labels: Record<string, string>;
    createdAt: Date;
    updatedAt?: Date | null;
    expireAt?: Date | null;
}
