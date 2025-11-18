import { type PaginationState, type SortingState } from "@infrastructure/data";
import { type UserBase } from "~/user-management/domain";

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
    id: string;
    data: {
        status: UserBase["status"];
    };
}>;

export type Users_UpdateOne_Res = ApiResponseBase<UserBase>;
