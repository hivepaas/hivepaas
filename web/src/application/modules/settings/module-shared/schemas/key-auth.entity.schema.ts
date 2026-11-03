import { z } from "zod";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

export const KeyAuthSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.string(),
    kind: z.string().optional(),
    inherited: z.boolean().optional(),
    keyId: z.string(),
    secretKey: z.string(),
    secretMasked: z.boolean().optional(),
});
