import { type AxiosResponse } from "axios";
import { z } from "zod";
import { AccessSchema } from "~/user-management/module-shared/schemas";

import { type Session_GetProfile_Res, type Session_Logout_Res } from "@application/shared/api/services";
import { ESecuritySettings, EUserRole, EUserStatus } from "@application/shared/enums";

import { parseApiResponse } from "@infrastructure/api";

/**
 * Get account API response schema
 */
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
            projectAccesses: z.array(AccessSchema).nullable(),
            moduleAccesses: z.array(AccessSchema).nullable(),
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
                moduleAccesses: user.moduleAccesses ?? [],
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
