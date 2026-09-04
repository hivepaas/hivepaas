import { type PaginationState, type SortingState } from "@infrastructure/data";
import { type UserBase } from "~/user-management/domain";

import type { ESecuritySettings, EUserRole, EUserStatus } from "@application/shared/enums";

import { type ApiRequestBase, type ApiResponseBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many users paginated
 */
export type Users_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type Users_FindManyPaginated_Res = ApiResponsePaginated<UserBase>;

/**
 * Find one user by id
 */
export type Users_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;

export type Users_FindOneById_Res = ApiResponseBase<UserBase>;

/**
 * Delete one user
 */
export type Users_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;

export type Users_DeleteOne_Res = ApiResponseBase<{
    id: string;
}>;

/**
 * Update one user
 */
export type Users_UpdateOne_Req = ApiRequestBase<{
    user: Partial<
        Omit<UserBase, "role" | "status" | "securityOption" | "createdAt" | "updatedAt" | "lastAccess" | "photo">
    > & {
        role?: EUserRole;
        status?: EUserStatus;
        securityOption?: ESecuritySettings;
    };
}>;

export type Users_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

/**
 * Invite a user
 */
export type Users_InviteOne_Req = ApiRequestBase<{
    user: Pick<UserBase, "email" | "accessExpireAt" | "projectAccesses" | "moduleAccesses"> & {
        role: EUserRole;
        securityOption: ESecuritySettings;
    };
    sendInviteEmail?: boolean;
}>;

export type Users_InviteOne_Res = ApiResponseBase<{ inviteLink: string }>;

/**
 * Reset user password
 */
export type Users_ResetPassword_Req = ApiRequestBase<{ id: string }>;

export type Users_ResetPassword_Res = ApiResponseBase<{ resetPasswordLink: string }>;
