import { type AxiosResponse } from "axios";
import { z } from "zod";
import { RepoWebhookSettingEntitySchema } from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    RepoWebhook_CreateOne_Res,
    RepoWebhook_DeleteOne_Res,
    RepoWebhook_FindManyPaginated_Res,
    RepoWebhook_FindOneById_Res,
    RepoWebhook_UpdateOne_Res,
    RepoWebhook_UpdateStatus_Res,
} from "./repo-webhook.api.contracts";

const FindManyPaginatedSchema = z.object({
    data: z.array(RepoWebhookSettingEntitySchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: RepoWebhookSettingEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
        secret: z.string().default(""),
        webhookURL: z.string().default(""),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class RepoWebhookApiValidator {
    findManyPaginated = (response: AxiosResponse): RepoWebhook_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return { data, meta };
    };

    findOneById = (response: AxiosResponse): RepoWebhook_FindOneById_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return { data, meta };
    };

    createOne = (response: AxiosResponse): RepoWebhook_CreateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return { data, meta };
    };

    updateOne = (response: AxiosResponse): RepoWebhook_UpdateOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    updateStatus = (response: AxiosResponse): RepoWebhook_UpdateStatus_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    deleteOne = (response: AxiosResponse): RepoWebhook_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
