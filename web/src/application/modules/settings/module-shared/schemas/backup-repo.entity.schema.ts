import { z } from "zod";

import { SettingsBaseEntitySchema } from "./settings-base.schema";

const BackupRepoRetentionSchema = z.object({
    keepLast: z.number().optional(),
    keepHourly: z.number().optional(),
    keepDaily: z.number().optional(),
    keepWeekly: z.number().optional(),
    keepMonthly: z.number().optional(),
});

export const BackupRepoSettingEntitySchema = SettingsBaseEntitySchema.extend({
    engine: z.string().optional(),
    description: z.string().optional(),
    cloudStorage: z.object({ id: z.string(), name: z.string().optional() }).nullish(),
    volume: z.object({ id: z.string(), name: z.string().optional() }).nullish(),
    storagePrefix: z.string().optional(),
    password: z.string().optional(),
    compression: z.string().optional(),
    packSize: z.string().optional(),
    retention: BackupRepoRetentionSchema.nullish(),
    secretMasked: z.boolean().optional(),
});
