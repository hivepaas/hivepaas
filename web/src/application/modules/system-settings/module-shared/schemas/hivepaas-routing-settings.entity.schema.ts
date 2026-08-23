import { z } from "zod";

const SettingRefSchema = z
    .object({
        id: z.string(),
        name: z.string(),
    })
    .passthrough();

export const HivePaaSRoutingClientConfigSchema = z.object({
    enabled: z.boolean(),
    allowedIPs: z.array(z.string()).nullish(),
});

export const HivePaaSRoutingRateLimitConfigSchema = z.object({
    enabled: z.boolean(),
    average: z.number(),
    period: z.string(),
    burst: z.number(),
    maxInFlightReq: z.number(),
});

export const HivePaaSRoutingDomainSchema = z.object({
    enabled: z.boolean(),
    domain: z.string(),
    sslCert: SettingRefSchema.nullish(),
    clientConfig: HivePaaSRoutingClientConfigSchema.nullish(),
    rateLimitConfig: HivePaaSRoutingRateLimitConfigSchema.nullish(),
});

export const HivePaaSRoutingSettingsEntitySchema = z.object({
    domains: z.array(HivePaaSRoutingDomainSchema).nullish(),
    updateVer: z.number(),
});

export const HivePaaSHttpSettingsEntitySchema = HivePaaSRoutingSettingsEntitySchema;
