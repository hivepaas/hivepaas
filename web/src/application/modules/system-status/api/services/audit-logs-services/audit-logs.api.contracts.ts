import type { PaginationState, SortingState } from "@infrastructure/data";
import type { AuditLog, AuditLogScope } from "~/system-status/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type AuditLogs_FindManyPaginated_Req = ApiRequestBase<{
    scope?: AuditLogScope;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
    type?: string[];
    source?: string[];
    result?: string[];
    actorID?: string[];
    resourceID?: string[];
    fromDate?: string;
    toDate?: string;
    projectID?: string;
    appID?: string;
    scopeOnly?: boolean;
}>;

export type AuditLogs_FindManyPaginated_Res = ApiResponsePaginated<AuditLog>;

export type AuditLogs_FindOneById_Req = ApiRequestBase<{
    scope?: AuditLogScope;
    itemID: string;
}>;

export type AuditLogs_FindOneById_Res = ApiResponseBase<AuditLog>;

export type AuditLogs_FindTypes_Req = ApiRequestBase<{
    scope?: AuditLogScope;
}>;

export type AuditLogs_FindTypes_Res = ApiResponseBase<string[]>;
