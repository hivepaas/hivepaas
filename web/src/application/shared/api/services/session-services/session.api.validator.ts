import { type AxiosResponse } from "axios";
import { z } from "zod";
import { AccessSchema } from "~/user-management/module-shared/schemas";

import { type Session_GetProfile_Res, type Session_Logout_Res } from "@application/shared/api/services";
import { ESecuritySettings, EUserRole, EUserStatus } from "@application/shared/enums";
import type { ModulePermission } from "@application/shared/permissions";

import { parseApiResponse } from "@infrastructure/api";

/**
 * Get account API response schema
 */
const ModuleAccessSchema = AccessSchema.extend({
    name: z.string().optional().default(""),
    access: z
        .object({
            read: z.boolean().nullable().optional(),
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
        write: moduleAccess.access?.write === true,
        delete: moduleAccess.access?.delete === true,
    },
}));

function mapModuleAccessesToModulePermissions(
    moduleAccesses: readonly z.output<typeof ModuleAccessSchema>[],
): ModulePermission[] {
    return moduleAccesses.map(moduleAccess => ({
        moduleId: moduleAccess.id,
        actions: {
            read: moduleAccess.access.read,
            write: moduleAccess.access.write,
            delete: moduleAccess.access.delete,
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
            status: z.nativeEnum(EUserStatus),
            mfaSecret: z.string().optional(),
            projectAccesses: z.array(AccessSchema).nullable().optional(),
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
                projectAccesses: user.projectAccesses ?? [],
                moduleAccesses,
                modulePermissions: mapModuleAccessesToModulePermissions(moduleAccesses),
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
