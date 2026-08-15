import { z } from "zod";

import { ECloudStorageKind } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

export const CloudStorageS3EntitySchema = z.object({
    accessKeyId: z.string(),
    secretKey: z.string(),
    region: z.string(),
    bucket: z.string(),
    endpoint: z.string(),
});

export const CloudStorageSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    description: z.string().optional(),
    type: z.string(),
    kind: z.nativeEnum(ECloudStorageKind).optional(),
    inherited: z.boolean().optional(),
    s3: CloudStorageS3EntitySchema,
    secretMasked: z.boolean().optional(),
});
