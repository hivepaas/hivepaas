import { type AxiosResponse } from "axios";
import { z } from "zod";
import { AccessSchema } from "~/user-management/module-shared/schemas";

import { type Session_GetProfile_Res, type Session_Logout_Res } from "@application/shared/api/services";
import { ESecuritySettings, EUserRole } from "@application/shared/enums";
import type { ModulePermission, ProjectPermission } from "@application/shared/permissions";

import { parseApiResponse } from "@infrastructure/api";

/**
 * Get account API response schema
 */
const ModuleAccessSchema = AccessSchema.extend({
    name: z.string().optional().default(""),
    access: z
        .object({
            read: z.boolean().nullable().optional(),
            execute: z.boolean().nullable().optional(),
            write: z.boolean().nullable().optional(),
            delete: z.boolean().nullable().optional(),
        })
        .nullable()
        .optional(),
}).transform(moduleAccess => ({
    id: moduleAccess.id,
    name: moduleAccess.name,
    access: {
        read: moduleAccess.access?.read === true,
        execute: moduleAccess.access?.execute === true,
        write: moduleAccess.access?.write === true,
        delete: moduleAccess.access?.delete === true,
    },
}));

const EnvAccessSchema = z.object({
    id: z.string(),
    name: z.string().optional().default(""),
    color: z.string().optional().default(""),
    access: z
        .object({
            read: z.boolean().nullable().optional(),
            execute: z.boolean().nullable().optional(),
            write: z.boolean().nullable().optional(),
            delete: z.boolean().nullable().optional(),
        })
        .nullable()
        .optional(),
});

/**
 * The API reports the account's access per project env, grouped by project. The
 * project entry itself carries no access, so the project-level actions here are
 * the union across its envs: "can the account do this somewhere in the project".
 * Anything that must be exact for one env has to read envAccesses instead.
 */
const ProjectAccessSchema = z
    .object({
        project: z.object({
            id: z.string(),
            name: z.string().optional().default(""),
        }),
        envAccesses: z.array(EnvAccessSchema).nullish(),
    })
    .transform(projectAccess => {
        const envAccesses = (projectAccess.envAccesses ?? []).map(envAccess => ({
            id: envAccess.id,
            name: envAccess.name,
            color: envAccess.color,
            access: {
                read: envAccess.access?.read === true,
                execute: envAccess.access?.execute === true,
                write: envAccess.access?.write === true,
                delete: envAccess.access?.delete === true,
            },
        }));

        return {
            id: projectAccess.project.id,
            name: projectAccess.project.name,
            envAccesses,
            access: {
                read: envAccesses.some(({ access }) => access.read),
                execute: envAccesses.some(({ access }) => access.execute),
                write: envAccesses.some(({ access }) => access.write),
                delete: envAccesses.some(({ access }) => access.delete),
            },
        };
    });

function mapModuleAccessesToModulePermissions(
    moduleAccesses: readonly z.output<typeof ModuleAccessSchema>[],
): ModulePermission[] {
    return moduleAccesses.map(moduleAccess => ({
        moduleId: moduleAccess.id,
        actions: {
            read: moduleAccess.access.read,
            execute: moduleAccess.access.execute,
            write: moduleAccess.access.write,
            delete: moduleAccess.access.delete,
        },
    }));
}

function mapProjectAccessesToProjectPermissions(
    projectAccesses: readonly z.output<typeof ProjectAccessSchema>[],
): ProjectPermission[] {
    return projectAccesses.map(projectAccess => ({
        projectId: projectAccess.id,
        actions: {
            read: projectAccess.access.read,
            execute: projectAccess.access.execute,
            write: projectAccess.access.write,
            delete: projectAccess.access.delete,
        },
    }));
}

const GetProfileSchema = z.object({
    data: z.object({
        nextStep: z.string().optional(),
        user: z.object({
            id: z.string(),
            username: z.string(),
            fullName: z.string().optional(),
            photo: z.string().transform(arg => arg.trim() || null),
            email: z.string().optional(),
            securityOption: z.nativeEnum(ESecuritySettings),
            position: z.string().nullish(),
            notes: z.string().nullish(),
            role: z.nativeEnum(EUserRole),
            accessExpireAt: z.coerce.date().nullable(),
            lastAccess: z.coerce.date().nullable(),
            createdAt: z.coerce.date(),
            status: z.string(),
            mfaSecret: z.string().optional(),
            projectAccesses: z.array(ProjectAccessSchema).nullable().optional(),
            moduleAccesses: z.array(ModuleAccessSchema).nullable().optional(),
            mfaTotpActivated: z.boolean().optional(),
        }),
    }),
});

export class SessionApiValidator {
    /**
     * Validate and transform the get account API response
     */
    getProfile = (response: AxiosResponse): Session_GetProfile_Res => {
        const {
            data: { user, nextStep },
        } = parseApiResponse({
            response,
            schema: GetProfileSchema,
        });

        const moduleAccesses = user.moduleAccesses ?? [];
        const projectAccesses = user.projectAccesses ?? [];

        return {
            data: {
                id: user.id,
                username: user.username,
                fullName: user.fullName ?? null,
                photo: user.photo,
                email: user.email ?? null,
                notes: user.notes ?? "",
                securityOption: user.securityOption,
                mfaSecret: user.mfaSecret ?? "",
                role: user.role,
                accessExpireAt: user.accessExpireAt ?? null,
                createdAt: user.createdAt,
                lastAccess: user.lastAccess ?? null,
                nextStep,
                position: user.position ?? "",
                status: user.status,
                projectAccesses,
                moduleAccesses,
                modulePermissions: mapModuleAccessesToModulePermissions(moduleAccesses),
                projectPermissions: mapProjectAccessesToProjectPermissions(projectAccesses),
                mfaTotpActivated: user.mfaTotpActivated,
            },
        };
    };

    /**
     * Validate and transform the logout API response
     */
    logout = (_: AxiosResponse): Session_Logout_Res => {
        return {
            data: {
                type: "success",
            },
        };
    };
}
