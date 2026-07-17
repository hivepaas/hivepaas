export type HivePaaSHttpSslCertRef = {
    id: string;
    name: string;
};

export type HivePaaSHttpClientConfig = {
    enabled: boolean;
    allowedIPs: string[];
};

export type HivePaaSHttpRateLimitConfig = {
    enabled: boolean;
    average: number;
    period: string;
    burst: number;
    maxInFlightReq: number;
};

export type HivePaaSHttpDomain = {
    enabled: boolean;
    domain: string;
    sslCert?: HivePaaSHttpSslCertRef | null;
    clientConfig?: HivePaaSHttpClientConfig | null;
    rateLimitConfig?: HivePaaSHttpRateLimitConfig | null;
};

export type HivePaaSHttpSettings = {
    domains: HivePaaSHttpDomain[];
    updateVer: number;
};

export type HivePaaSHttpSettingsObjectIdReq = {
    id: string;
};

export type HivePaaSHttpSettingsUpdateDomain = {
    enabled: boolean;
    domain: string;
    sslCert: HivePaaSHttpSettingsObjectIdReq;
    clientConfig?: HivePaaSHttpClientConfig | null;
    rateLimitConfig?: HivePaaSHttpRateLimitConfig | null;
};

export type HivePaaSHttpSettingsUpdatePayload = {
    domains: HivePaaSHttpSettingsUpdateDomain[];
    updateVer: number;
};
