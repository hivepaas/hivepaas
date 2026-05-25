import { z } from "zod";

import { ESSHKeyType, ESettingType } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

export const SSHKeySettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    description: z.string().optional(),
    type: z.literal(ESettingType.SSHKey),
    kind: z.string().optional(),
    inherited: z.boolean().optional(),
    keyType: z.union([z.nativeEnum(ESSHKeyType), z.literal("")]).optional(),
    publicKey: z.string().optional(),
    privateKey: z.string(),
    passphrase: z.string().optional(),
    secretMasked: z.boolean().optional(),
});
