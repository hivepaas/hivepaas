import { z } from "zod";
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
        status: z.string().nullish().transform(value => value ?? ""),
        env: OptionalStringSchema,
    })
    .nullish()
    .transform(value => value ?? null);

export const ProjectAppSchema = z.object({
    id: z.string(),
    name: z.string(),
    photo: z.string(),
    status: z.string(),
    env: OptionalStringSchema,
    note: z.string(),
    tags: z.array(z.string()),
    key: z.string(),
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
    updateVer: z.number(),
    stats: ProjectAppStatsSchema.nullable(),
});
