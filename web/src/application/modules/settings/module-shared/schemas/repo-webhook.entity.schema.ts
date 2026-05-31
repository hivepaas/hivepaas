import { z } from "zod";

import { ESettingType } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

export const RepoWebhookSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.literal(ESettingType.RepoWebhook),
    kind: z.string().default(""),
    secret: z.string().default(""),
    webhookURL: z.string().default(""),
    secretMasked: z.boolean().optional(),
    inherited: z.boolean().optional(),
});
