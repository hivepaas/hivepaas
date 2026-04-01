export const EEndpointResolutionMode = {
    VIP: "vip",
    DNSRR: "dnsrr",
} as const;

export type EEndpointResolutionMode = (typeof EEndpointResolutionMode)[keyof typeof EEndpointResolutionMode];
