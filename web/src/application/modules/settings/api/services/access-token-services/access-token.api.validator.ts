import { z } from "zod";
import { AccessTokenSettingEntitySchema } from "~/settings/module-shared/schemas";

import { type ApiHttpResponse, BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    AccessToken_CreateOne_Res,
    AccessToken_DeleteOne_Res,
    AccessToken_FindManyPaginated_Res,
    AccessToken_FindOneById_Res,
    AccessToken_TestConn_Res,
    AccessToken_UpdateOne_Res,
    AccessToken_UpdateStatus_Res,
} from "./access-token.api.contracts";

const FindManyPaginatedSchema = z.object({
    data: z.array(AccessTokenSettingEntitySchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: AccessTokenSettingEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const CreateOneSchema = z.object({
    data: z.object({ id: z.string() }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class AccessTokenApiValidator {
    findManyPaginated = (response: ApiHttpResponse): AccessToken_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindManyPaginatedSchema });
        return { data, meta };
    };

    findOneById = (response: ApiHttpResponse): AccessToken_FindOneById_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneByIdSchema });
        return { data, meta };
    };

    createOne = (response: ApiHttpResponse): AccessToken_CreateOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: CreateOneSchema });
        return { data, meta };
    };

    updateOne = (response: ApiHttpResponse): AccessToken_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    updateStatus = (response: ApiHttpResponse): AccessToken_UpdateStatus_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    deleteOne = (response: ApiHttpResponse): AccessToken_DeleteOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    testConn = (response: ApiHttpResponse): AccessToken_TestConn_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
