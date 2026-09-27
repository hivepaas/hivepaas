import type { PaginationState, SortingState } from "@infrastructure/data";
import type { AppScheduledJobs_Upsert_Payload } from "~/projects/api/services/project-apps-services";
import type { AppScheduledJob, EnvScheduledJob } from "~/projects/domain";
import type { EAppScheduledJobType } from "~/projects/module-shared/enums";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

interface EnvScope {
    projectID: string;
    env: string;
}

export type EnvScheduledJobs_FindManyPaginated_Req = ApiRequestBase<
    EnvScope & {
        pagination?: PaginationState;
        sorting?: SortingState;
        search?: string;
        /** Only jobs of these types. */
        jobTypes?: EAppScheduledJobType[];
        /** Only the jobs of this app. */
        appId?: string;
    }
>;

/** The env's own jobs, and those of every app in the env. */
export type EnvScheduledJobs_FindManyPaginated_Res = ApiResponsePaginated<EnvScheduledJob>;

export type EnvScheduledJobs_FindOneById_Req = ApiRequestBase<EnvScope & { scheduledJobID: string }>;

export type EnvScheduledJobs_FindOneById_Res = ApiResponseBase<AppScheduledJob>;

export type EnvScheduledJobs_CreateOne_Req = ApiRequestBase<EnvScope & { payload: AppScheduledJobs_Upsert_Payload }>;

export type EnvScheduledJobs_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type EnvScheduledJobs_UpdateOne_Req = ApiRequestBase<
    EnvScope & {
        scheduledJobID: string;
        payload: AppScheduledJobs_Upsert_Payload & { updateVer: number };
    }
>;

export type EnvScheduledJobs_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type EnvScheduledJobs_UpdateStatus_Req = ApiRequestBase<
    EnvScope & {
        scheduledJobID: string;
        payload: {
            updateVer: number;
            status: ESettingStatus;
            expireAt: Date | null;
            inheritable: boolean;
            default: boolean;
        };
    }
>;

export type EnvScheduledJobs_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type EnvScheduledJobs_DeleteOne_Req = ApiRequestBase<EnvScope & { scheduledJobID: string }>;

export type EnvScheduledJobs_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;

export type EnvScheduledJobs_RunNow_Req = ApiRequestBase<EnvScope & { scheduledJobID: string }>;

export type EnvScheduledJobs_RunNow_Res = ApiResponseBase<{ task: { id: string } }>;
