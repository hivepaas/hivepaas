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
};

// Aliases for compatibility
export type HivePaaSHttpSettings = HivePaaSRoutingSettings;
export type HivePaaSHttpDomain = HivePaaSRoutingDomain;
export type HivePaaSHttpSettingsUpdatePayload = HivePaaSRoutingSettingsUpdatePayload;
