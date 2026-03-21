import { z } from "zod";

import { ESettingType } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

/**
 * Registry-auth setting row from API (aligned with BE RegistryAuthResp + BaseSettingResp).
 */
export const RegistryAuthSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    description: z.string().optional(),
    type: z.literal(ESettingType.RegistryAuth),
    kind: z.string().optional(),
    inherited: z.boolean().optional(),
    address: z.string(),
    username: z.string(),
    password: z.string(),
    readonly: z.boolean(),
    secretMasked: z.boolean().optional(),
});
