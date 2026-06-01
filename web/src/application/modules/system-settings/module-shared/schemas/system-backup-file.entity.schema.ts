import { z } from "zod";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { ESettingStatus } from "@application/shared/enums";

import { ESystemBackupFileStorageType, ESystemBackupFileType } from "../enums";

export const SystemBackupFileEntitySchema = z.object({
    id: z.string(),
    type: z.literal(ESystemBackupFileType.SystemBackup),
    status: z.nativeEnum(ESettingStatus),
    key: z.string(),
    name: z.string(),
    path: z.string(),
    bucket: z
        .string()
        .nullish()
        .transform(value => value ?? undefined),
    mimetype: z.string(),
    size: z.number(),
    sizeStr: z.string(),
    storageType: z.nativeEnum(ESystemBackupFileStorageType),
    storage: SettingsBaseEntitySchema.omit({ description: true }).nullish(),
    updateVer: z.number(),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date(),
});
