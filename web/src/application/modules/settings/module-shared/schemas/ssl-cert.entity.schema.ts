import { z } from "zod";

import { ESettingType, ESslCertType, ESslKeyType } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

const NamedObjectSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const SslCertEventNotificationSchema = z.object({
    success: NamedObjectSchema.nullish(),
    successUseDefault: z.boolean(),
    failure: NamedObjectSchema.nullish(),
    failureUseDefault: z.boolean(),
});

const SslCertTypeSchema = z
    .union([z.nativeEnum(ESslCertType), z.literal("googlets")])
    .transform(value => (value === "googlets" ? ESslCertType.GoogleTrust : value));

/**
 * SSL cert setting from API (aligned with BE `SSLCertResp` + `BaseSettingResp`).
 */
export const SslCertSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.literal(ESettingType.SSLCert),
    kind: z.string().optional(),
    inherited: z.boolean().optional(),
    certType: SslCertTypeSchema,
    provider: NamedObjectSchema.nullish(),
    domain: z.string(),
    certificate: z.string(),
    privateKey: z.string(),
    keyType: z.nativeEnum(ESslKeyType),
    validPeriod: z.string(),
    email: z.string(),
    autoRenew: z.boolean(),
    renewableFrom: z.coerce.date().nullish(),
    notifyFrom: z.coerce.date().nullish(),
    notification: SslCertEventNotificationSchema.nullish(),
    secretMasked: z.boolean().optional(),
});
