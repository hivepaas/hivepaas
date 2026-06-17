import { z } from "zod";

import { EAcmeDnsProviderKind, ESettingType } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

const stringWithDefault = z.string().optional().default("");

export const AcmeDnsProviderAcmeDNSEntitySchema = z.object({
    apiBase: z.string(),
    allowList: z.array(z.string()).optional().default([]),
    storagePath: stringWithDefault,
    storageBaseUrl: stringWithDefault,
});

export const AcmeDnsProviderAzureEntitySchema = z.object({
    clientId: z.string(),
    clientSecret: z.string(),
    subscriptionId: stringWithDefault,
    tenantId: stringWithDefault,
    resourceGroupName: stringWithDefault,
});

export const AcmeDnsProviderBaiduCloudEntitySchema = z.object({
    accessKey: z.string(),
    secretKey: z.string(),
});

export const AcmeDnsProviderCloudflareEntitySchema = z.object({
    authToken: z.string(),
});

export const AcmeDnsProviderDigitalOceanEntitySchema = z.object({
    authToken: z.string(),
});

export const AcmeDnsProviderGCloudEntitySchema = z.object({
    projectId: stringWithDefault,
    serviceAccount: z.string(),
});

export const AcmeDnsProviderGoDaddyEntitySchema = z.object({
    apiKey: z.string(),
    apiSecret: z.string(),
});

export const AcmeDnsProviderHetznerEntitySchema = z.object({
    apiToken: z.string(),
});

export const AcmeDnsProviderHuaweiCloudEntitySchema = z.object({
    accessKey: z.string(),
    secretKey: z.string(),
    region: stringWithDefault,
});

export const AcmeDnsProviderNamecheapEntitySchema = z.object({
    apiUser: z.string(),
    apiKey: z.string(),
});

export const AcmeDnsProviderRFC2136EntitySchema = z.object({
    nameserver: z.string(),
    tsigKeyName: z.string(),
    tsigSecret: z.string(),
    tsigAlgorithm: stringWithDefault,
});

export const AcmeDnsProviderRoute53EntitySchema = z.object({
    accessKeyId: z.string(),
    secretAccessKey: z.string(),
    hostedZoneId: stringWithDefault,
    region: stringWithDefault,
});

export const AcmeDnsProviderTencentCloudEntitySchema = z.object({
    secretId: z.string(),
    secretKey: z.string(),
    region: stringWithDefault,
});

export const AcmeDnsProviderSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.literal(ESettingType.AcmeDnsProvider),
    kind: z.nativeEnum(EAcmeDnsProviderKind),
    inherited: z.boolean().optional(),
    acmeDns: AcmeDnsProviderAcmeDNSEntitySchema.nullish(),
    azure: AcmeDnsProviderAzureEntitySchema.nullish(),
    baiduCloud: AcmeDnsProviderBaiduCloudEntitySchema.nullish(),
    cloudflare: AcmeDnsProviderCloudflareEntitySchema.nullish(),
    digitalOcean: AcmeDnsProviderDigitalOceanEntitySchema.nullish(),
    gCloud: AcmeDnsProviderGCloudEntitySchema.nullish(),
    goDaddy: AcmeDnsProviderGoDaddyEntitySchema.nullish(),
    hetzner: AcmeDnsProviderHetznerEntitySchema.nullish(),
    huaweiCloud: AcmeDnsProviderHuaweiCloudEntitySchema.nullish(),
    namecheap: AcmeDnsProviderNamecheapEntitySchema.nullish(),
    rfc2136: AcmeDnsProviderRFC2136EntitySchema.nullish(),
    route53: AcmeDnsProviderRoute53EntitySchema.nullish(),
    tencentCloud: AcmeDnsProviderTencentCloudEntitySchema.nullish(),
    secretMasked: z.boolean().optional(),
});
