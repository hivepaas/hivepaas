import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { ESettingType } from "@application/shared/enums";

const TraefikAppSettingsSchema = z.object({
    replicas: z.number(),
});

export const TraefikServiceSettingsEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.literal(ESettingType.TraefikService),
    appSettings: TraefikAppSettingsSchema,
});
