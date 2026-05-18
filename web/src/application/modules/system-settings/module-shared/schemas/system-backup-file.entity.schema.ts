import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { ESettingType } from "@application/shared/enums";

import { ESystemBackupFileStorageType } from "../enums";

export const SystemBackupFileEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    type: z.literal(ESettingType.File),
    storageType: z.nativeEnum(ESystemBackupFileStorageType),
    storage: SettingsBaseEntitySchema.omit({ description: true }).nullish(),
    bucket: z
        .string()
        .nullish()
        .transform(value => value ?? undefined),
    mimetype: z.string(),
    size: z.number(),
    path: z.string(),
});
