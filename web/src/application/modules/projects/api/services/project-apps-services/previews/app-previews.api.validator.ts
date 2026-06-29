import { type AxiosResponse } from "axios";
import { z } from "zod";
import type {
    AppPreviews_CreateOne_Res,
    AppPreviews_FindManyPaginated_Res,
    AppPreviews_PrepareCreate_Res,
} from "~/projects/api/services";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const ProjectAppStatsSchema = z.object({
    runningTasks: z.number(),
    desiredTasks: z.number(),
    completedTasks: z.number(),
});

const ProjectAppSchema = z.object({
    id: z.string(),
    name: z.string(),
    status: z.nativeEnum(EProjectAppStatus),
    env: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    note: z.string(),
    tags: z.array(z.string()),
    key: z.string(),
    localKey: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    updateVer: z.number(),
    stats: ProjectAppStatsSchema.nullable(),
    accessLinks: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullable(),
});

const FindManyPaginatedSchema = z.object({
    data: z.array(ProjectAppSchema),
    meta: PagingMetaApiSchema,
});

const PrepareCreateSchema = z.object({
    data: z.object({
        repoURL: z.string(),
        repoCredentials: z
            .object({
                id: z.string(),
            })
            .nullish()
            .transform(value => value ?? null),
        canListBranches: z.boolean(),
        canListPullRequests: z.boolean(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

export class AppPreviewsApiValidator {
    findManyPaginated = (response: AxiosResponse): AppPreviews_FindManyPaginated_Res => {
        return parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });
    };

    prepareCreate = (response: AxiosResponse): AppPreviews_PrepareCreate_Res => {
        return parseApiResponse({
            response,
            schema: PrepareCreateSchema,
        });
    };

    createOne = (response: AxiosResponse): AppPreviews_CreateOne_Res => {
        return parseApiResponse({
            response,
            schema: CreateOneSchema,
        });
    };
}
