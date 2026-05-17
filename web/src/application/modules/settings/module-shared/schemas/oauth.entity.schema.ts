import { z } from "zod";

import { ESettingType } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

export const OAuthSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    description: z.string().optional(),
    type: z.literal(ESettingType.OAuth),
    kind: z.string().optional(),
    organization: z.string(),
    clientId: z.string(),
    clientSecret: z.string(),
    callbackURL: z.string().optional(),
    authURL: z.string().optional(),
    tokenURL: z.string().optional(),
    profileURL: z.string().optional(),
    autoDiscoveryURL: z.string().optional(),
    scopes: z.array(z.string()).optional(),
    secretMasked: z.boolean().optional(),
});
