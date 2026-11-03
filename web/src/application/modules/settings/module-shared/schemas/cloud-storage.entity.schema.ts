import { z } from "zod";

import { ECloudStorageKind } from "@application/shared/enums";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

export const CloudStorageS3EntitySchema = z.object({
    keyAuth: z
        .object({ id: z.string(), name: z.string(), status: z.string().optional() })
        .nullish()
        .transform(value => value ?? null),
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
});
