import { z } from "zod";
import { AppScheduledJobSchema } from "~/projects/api/services/project-apps-services/scheduled-jobs/app-scheduled-jobs.api.validator";

import { type ApiHttpResponse, BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    EnvScheduledJobs_CreateOne_Res,
    EnvScheduledJobs_FindManyPaginated_Res,
    EnvScheduledJobs_FindOneById_Res,
    EnvScheduledJobs_RunNow_Res,
} from "./project-env-scheduled-jobs.api.contracts";

const EnvScheduledJobSchema = AppScheduledJobSchema.extend({
    scope: z.enum(["project-env", "app"]).catch("app"),
    ownerApp: z
        .object({ id: z.string(), name: z.string() })
        .nullish()
        .transform(value => value ?? undefined),
});

const FindManyPaginatedSchema = z.object({
    data: z.array(EnvScheduledJobSchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: AppScheduledJobSchema,
    meta: BaseMetaApiSchema.nullable(),
});

const CreateOneSchema = z.object({
    data: z.object({ id: z.string() }),
    meta: BaseMetaApiSchema.nullable(),
});

const RunNowSchema = z.object({
    data: z.object({ task: z.object({ id: z.string() }) }),
    meta: BaseMetaApiSchema.nullable(),
});

export class EnvScheduledJobsApiValidator {
    findManyPaginated = (response: ApiHttpResponse): EnvScheduledJobs_FindManyPaginated_Res => {
        return parseApiResponse({ response, schema: FindManyPaginatedSchema });
    };

    findOneById = (response: ApiHttpResponse): EnvScheduledJobs_FindOneById_Res => {
        return parseApiResponse({ response, schema: FindOneByIdSchema });
    };

    createOne = (response: ApiHttpResponse): EnvScheduledJobs_CreateOne_Res => {
        return parseApiResponse({ response, schema: CreateOneSchema });
    };

    runNow = (response: ApiHttpResponse): EnvScheduledJobs_RunNow_Res => {
        return parseApiResponse({ response, schema: RunNowSchema });
    };
}
