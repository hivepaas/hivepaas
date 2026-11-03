import { z } from "zod";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

export const BasicAuthSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.string(),
    kind: z.string().optional(),
    inherited: z.boolean().optional(),
    username: z.string(),
    password: z.string(),
    secretMasked: z.boolean().optional(),
});
