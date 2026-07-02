import { z } from "zod";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

const OptionalStringSchema = z
    .string()
    .nullish()
    .transform(value => value ?? "");

export const ProjectAppStatsSchema = z.object({
    runningTasks: z.number(),
    desiredTasks: z.number(),
    completedTasks: z.number(),
});

export const ProjectAppParentSchema = z
    .object({
        id: z.string(),
        name: OptionalStringSchema,
        key: OptionalStringSchema,
        localKey: OptionalStringSchema,
        status: z
            .union([z.nativeEnum(EProjectAppStatus), z.literal("")])
            .nullish()
            .transform(value => value ?? ""),
        env: OptionalStringSchema,
    })
    .nullish()
    .transform(value => value ?? null);

export const ProjectAppSchema = z.object({
    id: z.string(),
    name: z.string(),
    status: z.nativeEnum(EProjectAppStatus),
    env: OptionalStringSchema,
    note: z.string(),
    tags: z.array(z.string()),
    key: z.string(),
    localKey: OptionalStringSchema,
    updateVer: z.number(),
    stats: ProjectAppStatsSchema.nullable(),
    parentApp: ProjectAppParentSchema,
    accessLinks: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullable(),
});

export const ProjectAppDetailsSchema = ProjectAppSchema.extend({
    key: z.string(),
    localKey: OptionalStringSchema,
    updateVer: z.number(),
    stats: ProjectAppStatsSchema.nullable(),
});
