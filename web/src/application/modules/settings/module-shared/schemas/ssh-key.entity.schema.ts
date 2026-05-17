import { z } from "zod";

import { ESettingType } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

export const SSHKeySettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    description: z.string().optional(),
    type: z.literal(ESettingType.SSHKey),
    kind: z.string().optional(),
    inherited: z.boolean().optional(),
    privateKey: z.string(),
    passphrase: z.string().optional(),
    targets: z.array(z.string()).optional(),
    secretMasked: z.boolean().optional(),
});
