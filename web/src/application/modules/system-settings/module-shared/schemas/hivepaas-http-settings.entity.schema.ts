import { z } from "zod";

const SettingRefSchema = z
    .object({
        id: z.string(),
        name: z.string(),
    })
    .passthrough();

export const HivePaaSHttpClientConfigSchema = z.object({
    enabled: z.boolean(),
    allowedIPs: z.array(z.string()).nullish(),
});

export const HivePaaSHttpRateLimitConfigSchema = z.object({
    enabled: z.boolean(),
    average: z.number(),
    period: z.string(),
    burst: z.number(),
    maxInFlightReq: z.number(),
});

export const HivePaaSHttpDomainSchema = z.object({
    enabled: z.boolean(),
    domain: z.string(),
    sslCert: SettingRefSchema.nullish(),
    clientConfig: HivePaaSHttpClientConfigSchema.nullish(),
    rateLimitConfig: HivePaaSHttpRateLimitConfigSchema.nullish(),
});

export const HivePaaSHttpSettingsEntitySchema = z.object({
    domains: z.array(HivePaaSHttpDomainSchema).nullish(),
    updateVer: z.number(),
});
