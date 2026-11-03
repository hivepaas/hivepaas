import { z } from "zod";

import { BACKUP_REPO_ACTION, BACKUP_REPO_STORAGE_TYPE } from "../../constants/backup-repo.constants";

const NamedObjectSchema = z
    .object({
        id: z.string(),
        name: z.string().optional(),
    })
    .nullable()
    .optional();

export const CreateOrEditBackupRepoFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    engine: z.string(),
    action: z.enum([BACKUP_REPO_ACTION.CreateNew, BACKUP_REPO_ACTION.ImportExisting]),
    description: z.string().trim().optional(),

    storageType: z.enum([BACKUP_REPO_STORAGE_TYPE.CloudStorage, BACKUP_REPO_STORAGE_TYPE.Volume]),
    cloudStorage: NamedObjectSchema,
    volume: NamedObjectSchema,
    storagePrefix: z.string().trim().optional(),

    password: z.string().optional(),
    compression: z.string(),
    packSize: z.string().trim(),

    retention: z.object({
        keepLast: z.coerce.number().int().min(0),
        keepHourly: z.coerce.number().int().min(0),
        keepDaily: z.coerce.number().int().min(0),
        keepWeekly: z.coerce.number().int().min(0),
        keepMonthly: z.coerce.number().int().min(0),
    }),

    inheritable: z.boolean(),
    default: z.boolean(),
});

export type CreateOrEditBackupRepoFormInput = z.input<typeof CreateOrEditBackupRepoFormSchema>;
export type CreateOrEditBackupRepoFormOutput = z.output<typeof CreateOrEditBackupRepoFormSchema>;
