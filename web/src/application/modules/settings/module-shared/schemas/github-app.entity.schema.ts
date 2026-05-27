import { z } from "zod";

import { ESettingType } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

export const GithubAppSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.literal(ESettingType.GithubApp),
    clientId: z.string().default(""),
    clientSecret: z.string().default(""),
    organization: z.string().default(""),
    callbackURL: z.string().optional().default(""),
    webhookURL: z.string().optional().default(""),
    webhookSecret: z.string().optional().default(""),
    appId: z.coerce.number().default(0),
    installationId: z.coerce.number().default(0),
    privateKey: z.string().default(""),
    ssoEnabled: z.boolean().default(false),
    secretMasked: z.boolean().optional(),
    inherited: z.boolean().optional(),
});
