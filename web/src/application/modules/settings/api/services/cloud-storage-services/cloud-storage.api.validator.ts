import { type AxiosResponse } from "axios";
import { z } from "zod";
import { CloudStorageSettingEntitySchema } from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    CloudStorage_CreateOne_Res,
    CloudStorage_DeleteOne_Res,
    CloudStorage_FindManyPaginated_Res,
    CloudStorage_FindOneById_Res,
    CloudStorage_TestConn_Res,
    CloudStorage_UpdateMeta_Res,
    CloudStorage_UpdateOne_Res,
} from "./cloud-storage.api.contracts";

const FindManyPaginatedSchema = z.object({
    data: z.array(CloudStorageSettingEntitySchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: CloudStorageSettingEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const CreateOneSchema = z.object({
    data: z.object({ id: z.string() }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class CloudStorageApiValidator {
    findManyPaginated = (response: AxiosResponse): CloudStorage_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindManyPaginatedSchema });
        return { data, meta };
    };

    findOneById = (response: AxiosResponse): CloudStorage_FindOneById_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneByIdSchema });
        return { data, meta };
    };

    createOne = (response: AxiosResponse): CloudStorage_CreateOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: CreateOneSchema });
        return { data, meta };
    };

    updateOne = (response: AxiosResponse): CloudStorage_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    updateMeta = (response: AxiosResponse): CloudStorage_UpdateMeta_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    deleteOne = (response: AxiosResponse): CloudStorage_DeleteOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    testConn = (response: AxiosResponse): CloudStorage_TestConn_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
