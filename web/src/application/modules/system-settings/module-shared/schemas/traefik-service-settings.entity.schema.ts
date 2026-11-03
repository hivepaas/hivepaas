import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

const TraefikAppSettingsSchema = z.object({
    replicas: z.number(),
});

export const TraefikServiceSettingsEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.string(),
    appSettings: TraefikAppSettingsSchema,
});
