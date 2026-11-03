import { z } from "zod";

const AccessActionsSchema = z.object({
    read: z.boolean(),
    execute: z.boolean(),
    write: z.boolean(),
    delete: z.boolean(),
});

export const AccessSchema = z.object({
    id: z.string(),
    name: z.string(),
    access: AccessActionsSchema,
});

/**
 * Access to a single project env. `id` is the project env id ("<projectId>:<envKey>"),
 * which is what the API expects when granting permissions.
 */
export const ProjectEnvAccessSchema = z.object({
    id: z.string(),
    name: z.string(),
    color: z.string(),
    access: AccessActionsSchema,
});

/**
 * Access to a project, expressed per env. The project itself carries no access:
 * permissions are granted per env only.
 */
export const ProjectAccessSchema = z.object({
    id: z.string(),
    name: z.string(),
    envAccesses: z.array(ProjectEnvAccessSchema),
});

export type AccessActions = z.infer<typeof AccessActionsSchema>;
export type ProjectEnvAccess = z.infer<typeof ProjectEnvAccessSchema>;
export type ProjectAccess = z.infer<typeof ProjectAccessSchema>;
