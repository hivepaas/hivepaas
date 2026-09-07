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

/**
 * A routing change that has been applied but not yet vouched for.
 *
 * HivePaaS applies routing changes on trial: unless a request gets back in
 * through the new configuration and confirms it, the change is undone at
 * `deadlineAt`. That is the guard for the mistakes nothing can detect up front -
 * a broken auth reference, a domain whose DNS is not ready, an allowlist that
 * excludes the person editing it.
 *
 * `confirmableFrom` is earlier than `deadlineAt` and later than `appliedAt`: the
 * proxy needs a few seconds to pick the change up, and a confirmation sent before
 * that would be vouching for the configuration being replaced.
 */
export type HivePaaSRoutingPendingChange = {
    changeId: string;
    appliedAt: Date;
    confirmableFrom: Date;
    deadlineAt: Date;
};

export type HivePaaSRoutingSettings = {
    domains: HivePaaSRoutingDomain[];
    updateVer: number;
    pendingChange: HivePaaSRoutingPendingChange | null;
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
