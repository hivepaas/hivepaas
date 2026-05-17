import { type AxiosResponse } from "axios";
import { z } from "zod";
import { OAuthSettingEntitySchema } from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    OAuth_CreateOne_Res,
    OAuth_DeleteOne_Res,
    OAuth_FindManyPaginated_Res,
    OAuth_FindOneById_Res,
    OAuth_UpdateMeta_Res,
    OAuth_UpdateOne_Res,
} from "./oauth.api.contracts";

const FindManyPaginatedSchema = z.object({
    data: z.array(OAuthSettingEntitySchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: OAuthSettingEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
        callbackURL: z.string().optional(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class OAuthApiValidator {
    findManyPaginated = (response: AxiosResponse): OAuth_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindManyPaginatedSchema });
        return { data, meta };
    };

    findOneById = (response: AxiosResponse): OAuth_FindOneById_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneByIdSchema });
        return { data, meta };
    };

    createOne = (response: AxiosResponse): OAuth_CreateOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: CreateOneSchema });
        return { data, meta };
    };

    updateOne = (response: AxiosResponse): OAuth_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    updateMeta = (response: AxiosResponse): OAuth_UpdateMeta_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    deleteOne = (response: AxiosResponse): OAuth_DeleteOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
