import { z } from "zod";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

export const AccessTokenSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    description: z.string().optional(),
    type: z.string(),
    kind: z.string().optional(),
    inherited: z.boolean().optional(),
    user: z.string(),
    token: z.string(),
    baseURL: z.string(),
    secretMasked: z.boolean().optional(),
});
