import { type AxiosResponse } from "axios";
import { z } from "zod";
import { type Users_FindManyPaginated_Res } from "~/user-management/api/services/users-services/users/users.api.contracts";

import { ESecuritySettings, EUserRole, EUserStatus } from "@application/shared/enums";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * Find many users paginated API response schema
 */
const FindManyPaginatedSchema = z.object({
    data: z.array(
        z.object({
            id: z.string(),
            email: z.string(),
            fullName: z.string(),
            photo: z.string().nullable(),
            position: z.string(),
            securityOption: z.nativeEnum(ESecuritySettings),
            status: z.nativeEnum(EUserStatus),
            role: z.nativeEnum(EUserRole),
            createdAt: z.coerce.date(),
            updatedAt: z.coerce.date().nullable(),
            accessExpireAt: z.coerce.date().nullable(),
            lastAccess: z.coerce.date().nullable(),
        }),
    ),
    meta: PagingMetaApiSchema,
});

export class UsersApiValidator {
    /**
     * Validate and transform find many users paginated API response
     */
    findManyPaginated = (response: AxiosResponse): Users_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return {
            data: data.map(user => ({
                id: user.id,
                email: user.email,
                role: user.role,
                status: user.status,
                fullName: user.fullName,
                photo: user.photo,
                position: user.position,
                securityOption: user.securityOption,
                createdAt: user.createdAt,
                updatedAt: user.updatedAt,
                accessExpireAt: user.accessExpireAt,
                lastAccess: user.lastAccess,
            })),
            meta,
        };
    };
}
