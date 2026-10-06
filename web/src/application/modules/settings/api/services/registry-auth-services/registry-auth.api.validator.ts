import { z } from "zod";
import { RegistryAuthSettingEntitySchema } from "~/settings/module-shared/schemas";

import { type ApiHttpResponse, BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    RegistryAuth_CreateOne_Res,
    RegistryAuth_DeleteOne_Res,
    RegistryAuth_FindManyPaginated_Res,
    RegistryAuth_FindOneById_Res,
    RegistryAuth_TestConn_Res,
    RegistryAuth_UpdateOne_Res,
    RegistryAuth_UpdateStatus_Res,
} from "./registry-auth.api.contracts";

const FindManyPaginatedSchema = z.object({
    data: z.array(RegistryAuthSettingEntitySchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: RegistryAuthSettingEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class RegistryAuthApiValidator {
    findManyPaginated = (response: ApiHttpResponse): RegistryAuth_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return { data, meta };
    };

    findOneById = (response: ApiHttpResponse): RegistryAuth_FindOneById_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return { data, meta };
    };

    createOne = (response: ApiHttpResponse): RegistryAuth_CreateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return { data, meta };
    };

    updateOne = (response: ApiHttpResponse): RegistryAuth_UpdateOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    updateStatus = (response: ApiHttpResponse): RegistryAuth_UpdateStatus_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    deleteOne = (response: ApiHttpResponse): RegistryAuth_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    testConn = (response: ApiHttpResponse): RegistryAuth_TestConn_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
