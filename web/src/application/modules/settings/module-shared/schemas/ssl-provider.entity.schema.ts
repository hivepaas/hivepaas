import { z } from "zod";

import { ESslKeyType, ESslProviderKind } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

const SslProviderLetsEncryptSchema = z.object({});

const SslProviderEabSchema = z.object({
    eabKid: z.string(),
    eabHmacKey: z.string(),
});

export const SslProviderSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.string(),
    kind: z.nativeEnum(ESslProviderKind),
    inherited: z.boolean().optional(),
    email: z.string(),
    defaultKeyType: z.nativeEnum(ESslKeyType).or(z.literal("")).optional().default(""),
    letsEncrypt: SslProviderLetsEncryptSchema.nullish(),
    zeroSSL: SslProviderEabSchema.nullish(),
    googleTrust: SslProviderEabSchema.nullish(),
    secretMasked: z.boolean().optional(),
});
