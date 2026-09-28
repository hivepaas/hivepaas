import type { PaginationState } from "@infrastructure/data";
import type { BackupSnapshot, BackupSnapshotRepoRef, BackupSnapshotScope } from "~/settings/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

/** What a view is narrowed to. The API takes several repositories and apps. */
export type BackupSnapshot_Filters = {
    repo?: string[];
    app?: string[];
    /** key:value, all of which must match. */
    tag?: string[];
    fromDate?: string;
    toDate?: string;
    search?: string;
};

export type BackupSnapshot_FindManyPaginated_Req = ApiRequestBase<
    {
        scope: BackupSnapshotScope;
        pagination?: PaginationState;
    } & BackupSnapshot_Filters
>;

export type BackupSnapshot_FindManyPaginated_Res = ApiResponsePaginated<BackupSnapshot> & {
    /** The repositories the view reaches. */
    repos: BackupSnapshotRepoRef[];
};

export type BackupSnapshot_FindOneById_Req = ApiRequestBase<{
    scope: BackupSnapshotScope;
    id: string;
}>;

export type BackupSnapshot_FindOneById_Res = ApiResponseBase<BackupSnapshot>;

export type BackupSnapshot_DeleteOne_Req = ApiRequestBase<{
    scope: BackupSnapshotScope;
    id: string;
}>;

export type BackupSnapshot_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
