import type { PaginationState } from "@infrastructure/data";
import type { AppScheduledJobs_Command_Payload } from "~/projects/api/services";
import type {
    BackupRestoreMode,
    BackupSnapshot,
    BackupSnapshotEntry,
    BackupSnapshotRepoRef,
    BackupSnapshotScope,
} from "~/settings/domain";

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

export type BackupSnapshot_FindEntries_Req = ApiRequestBase<{
    scope: BackupSnapshotScope;
    id: string;
    /** A directory inside the snapshot; "" for its root. */
    path: string;
}>;

export type BackupSnapshot_FindEntries_Res = ApiResponseBase<BackupSnapshotEntry[]>;

/** A command snapshot's restore has `command`; a volume snapshot's the rest. */
export type BackupSnapshot_Restore_Payload = {
    targetApp: { id: string };
    command?: AppScheduledJobs_Command_Payload;
    volume?: { id: string };
    subpath?: string;
    snapshotPath?: string;
    stopApp?: boolean;
    mode?: BackupRestoreMode;
};

export type BackupSnapshot_Restore_Req = ApiRequestBase<{
    scope: BackupSnapshotScope;
    id: string;
    payload: BackupSnapshot_Restore_Payload;
}>;

export type BackupSnapshot_Restore_Res = ApiResponseBase<{ taskId: string }>;

export type BackupSnapshot_DownloadFile_Req = ApiRequestBase<{
    scope: BackupSnapshotScope;
    id: string;
    /** A file inside the snapshot. */
    path: string;
}>;

export type BackupSnapshot_DownloadFile_Res = ApiResponseBase<{ blob: Blob; filename: string }>;
