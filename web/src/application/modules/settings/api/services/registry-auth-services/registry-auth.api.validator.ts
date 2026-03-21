import { type AxiosResponse } from "axios";
import { z } from "zod";

import { RegistryAuthSettingEntitySchema } from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    RegistryAuth_CreateOne_Res,
    RegistryAuth_DeleteOne_Res,
    RegistryAuth_FindManyPaginated_Res,
    RegistryAuth_FindOneById_Res,
    RegistryAuth_TestConn_Res,
    RegistryAuth_UpdateMeta_Res,
    RegistryAuth_UpdateOne_Res,
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
    findManyPaginated = (response: AxiosResponse): RegistryAuth_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return { data, meta };
    };

    findOneById = (response: AxiosResponse): RegistryAuth_FindOneById_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return { data, meta };
    };

    createOne = (response: AxiosResponse): RegistryAuth_CreateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return { data, meta };
    };

    updateOne = (response: AxiosResponse): RegistryAuth_UpdateOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    updateMeta = (response: AxiosResponse): RegistryAuth_UpdateMeta_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    deleteOne = (response: AxiosResponse): RegistryAuth_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    testConn = (response: AxiosResponse): RegistryAuth_TestConn_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
