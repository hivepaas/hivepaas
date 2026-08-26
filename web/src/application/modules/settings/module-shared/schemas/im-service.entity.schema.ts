import { z } from "zod";

import { EImServiceKind } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

const WebhookSchema = z.object({
    webhook: z.string(),
});

const TelegramSchema = z.object({
    botToken: z.string(),
    chatId: z.string(),
});

const LarkSchema = z.object({
    webhook: z.string(),
    secret: z.string().nullish(),
});

export const ImServiceSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.string(),
    kind: z.nativeEnum(EImServiceKind),
    slack: WebhookSchema.nullish(),
    discord: WebhookSchema.nullish(),
    telegram: TelegramSchema.nullish(),
    lark: LarkSchema.nullish(),
    secretMasked: z.boolean().optional(),
    inherited: z.boolean().optional(),
});
