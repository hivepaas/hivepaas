import { z } from "zod";

import { ESslCertType, ESslKeyType } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

const DomainCertSettingsSchema = z.object({
    certType: z.nativeEnum(ESslCertType),
    keyType: z.nativeEnum(ESslKeyType),
    validPeriod: z.string(),
    email: z.string(),
    autoRenew: z.boolean(),
});

export const DomainSettingsEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    rootDomain: z.string(),
    certSettings: DomainCertSettingsSchema.nullish(),
});
