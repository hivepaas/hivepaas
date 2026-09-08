import type { SettingsPendingChange } from "./settings-probation.entity";

export type HivePaaSRoutingSslCertRef = {
    id: string;
    name: string;
};

export type HivePaaSRoutingClientConfig = {
    enabled: boolean;
    allowedIPs: string[];
};

export type HivePaaSRoutingRateLimitConfig = {
    enabled: boolean;
    average: number;
    period: string;
    burst: number;
    maxInFlightReq: number;
};

export type HivePaaSRoutingDomain = {
    enabled: boolean;
    domain: string;
    sslCert?: HivePaaSRoutingSslCertRef | null;
    clientConfig?: HivePaaSRoutingClientConfig | null;
    rateLimitConfig?: HivePaaSRoutingRateLimitConfig | null;
};

export type HivePaaSRoutingSettings = {
    domains: HivePaaSRoutingDomain[];
    updateVer: number;
    pendingChange: SettingsPendingChange | null;
};

export type HivePaaSRoutingSettingsObjectIdReq = {
    id: string;
};

export type HivePaaSRoutingSettingsUpdateDomain = {
    enabled: boolean;
    domain: string;
    sslCert: HivePaaSRoutingSettingsObjectIdReq;
    clientConfig?: HivePaaSRoutingClientConfig | null;
    rateLimitConfig?: HivePaaSRoutingRateLimitConfig | null;
};

export type HivePaaSRoutingSettingsUpdatePayload = {
    domains: HivePaaSRoutingSettingsUpdateDomain[];
    updateVer: number;

    /**
     * How long the change may stay unconfirmed before it is undone, as a Go
     * duration string ("5m", "90s").
     *
     * The server clamps it and has its own default, so omitting it is safe. There
     * is no value that turns the trial off.
     */
    confirmWindow?: string;
};

// Aliases for compatibility
export type HivePaaSHttpSettings = HivePaaSRoutingSettings;
export type HivePaaSHttpDomain = HivePaaSRoutingDomain;
export type HivePaaSHttpSettingsUpdatePayload = HivePaaSRoutingSettingsUpdatePayload;
