import { type AxiosResponse } from "axios";
import { z } from "zod";
import {
    type Users_FindManyPaginated_Res,
    type Users_FindOneById_Res,
    type Users_InviteOne_Res,
} from "~/user-management/api/services/users-services/users/users.api.contracts";
import { AccessSchema } from "~/user-management/module-shared/schemas";

import { ESecuritySettings, EUserRole, EUserStatus } from "@application/shared/enums";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * User schema
 */
const UserSchema = z.object({
    id: z.string(),
    email: z.string(),
    username: z.string().optional(),
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
});

/**
 * Find many users paginated API response schema
 */
const FindManyPaginatedSchema = z.object({
    data: z.array(UserSchema),
    meta: PagingMetaApiSchema,
});

/**
 * Find one user by id API response schema
 */
const FindOneByIdSchema = z.object({
    data: UserSchema.extend({
        projectAccesses: z.array(AccessSchema).nullable(),
        moduleAccesses: z.array(AccessSchema).nullable(),
    }),
});

/**
 * Invite one user API response schema
 */
const InviteOneSchema = z.object({
    data: z.object({
        inviteLink: z.string(),
    }),
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
                username: user.username ?? "",
                projectAccesses: [],
                moduleAccesses: [],
            })),
            meta,
        };
    };

    /**
     * Validate and transform find one user by id API response
     */
    findOneById = (response: AxiosResponse): Users_FindOneById_Res => {
        const { data } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return {
            data: {
                id: data.id,
                email: data.email,
                role: data.role,
                status: data.status,
                fullName: data.fullName,
                photo: data.photo,
                position: data.position,
                securityOption: data.securityOption,
                createdAt: data.createdAt,
                updatedAt: data.updatedAt,
                accessExpireAt: data.accessExpireAt,
                lastAccess: data.lastAccess,
                username: data.username ?? "",
                projectAccesses: data.projectAccesses ?? [],
                moduleAccesses: data.moduleAccesses ?? [],
            },
        };
    };

    /**
     * Validate and transform invite one user API response
     */
    inviteOne = (response: AxiosResponse): Users_InviteOne_Res => {
        const { data } = parseApiResponse({
            response,
            schema: InviteOneSchema,
        });

        return {
            data: {
                inviteLink: data.inviteLink,
            },
        };
    };
}
