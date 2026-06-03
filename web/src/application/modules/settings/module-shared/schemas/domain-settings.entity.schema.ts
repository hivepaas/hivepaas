import { z } from "zod";

import { ESslCertType, ESslKeyType } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

const DomainCertSettingsSchema = z.object({
    certType: z.nativeEnum(ESslCertType),
    keyType: z.preprocess(
        value => (value === "" || value === null ? undefined : value),
        z.nativeEnum(ESslKeyType).optional(),
    ),
    validPeriod: z.string().optional(),
    email: z.string(),
    autoRenew: z.boolean().optional(),
});

export const DomainSettingsEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    rootDomain: z.string(),
    allowedDomains: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    certSettings: DomainCertSettingsSchema.nullish(),
});
