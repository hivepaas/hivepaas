import { z } from "zod";

import { ESettingType } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

export const BasicAuthSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.literal(ESettingType.BasicAuth),
    kind: z.string().optional(),
    inherited: z.boolean().optional(),
    username: z.string(),
    password: z.string(),
    secretMasked: z.boolean().optional(),
});
