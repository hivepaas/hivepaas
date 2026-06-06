import { z } from "zod";

import { EImServiceKind, ESettingType } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

const WebhookSchema = z.object({
    webhook: z.string(),
});

const TelegramSchema = z.object({
    botToken: z.string(),
    chatId: z.string(),
});

export const ImServiceSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.literal(ESettingType.IMService),
    kind: z.nativeEnum(EImServiceKind),
    slack: WebhookSchema.nullish(),
    discord: WebhookSchema.nullish(),
    telegram: TelegramSchema.nullish(),
    secretMasked: z.boolean().optional(),
    inherited: z.boolean().optional(),
});
