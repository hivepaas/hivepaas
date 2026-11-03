import { z } from "zod";
import type {
    AppPreviews_CreateOne_Res,
    AppPreviews_FindManyPaginated_Res,
    AppPreviews_PrepareCreate_Res,
} from "~/projects/api/services";

import { type ApiHttpResponse, BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import { ProjectAppSchema } from "../project-apps/project-apps.api.schemas";

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
        canCloneDbApps: z
            .boolean()
            .nullish()
            .transform(value => value ?? false),
        canSkipCloningDbApps: z
            .boolean()
            .nullish()
            .transform(value => value ?? false),
        withheldSecrets: z
            .array(
                z.object({
                    name: z.string(),
                    envVars: z
                        .array(z.string())
                        .nullish()
                        .transform(value => value ?? []),
                }),
            )
            .nullish()
            .transform(value => value ?? []),
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
    findManyPaginated = (response: ApiHttpResponse): AppPreviews_FindManyPaginated_Res => {
        return parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });
    };

    prepareCreate = (response: ApiHttpResponse): AppPreviews_PrepareCreate_Res => {
        return parseApiResponse({
            response,
            schema: PrepareCreateSchema,
        });
    };

    createOne = (response: ApiHttpResponse): AppPreviews_CreateOne_Res => {
        return parseApiResponse({
            response,
            schema: CreateOneSchema,
        });
    };
}
