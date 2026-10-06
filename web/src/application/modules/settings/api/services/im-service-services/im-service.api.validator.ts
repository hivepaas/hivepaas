import { z } from "zod";
import { ImServiceSettingEntitySchema } from "~/settings/module-shared/schemas";

import { type ApiHttpResponse, BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    ImService_CreateOne_Res,
    ImService_DeleteOne_Res,
    ImService_FindManyPaginated_Res,
    ImService_FindOneById_Res,
    ImService_TestSendMsg_Res,
    ImService_UpdateOne_Res,
    ImService_UpdateStatus_Res,
} from "./im-service.api.contracts";

const FindManyPaginatedSchema = z.object({
    data: z.array(ImServiceSettingEntitySchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: ImServiceSettingEntitySchema,
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

export class ImServiceApiValidator {
    findManyPaginated = (response: ApiHttpResponse): ImService_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return { data, meta };
    };

    findOneById = (response: ApiHttpResponse): ImService_FindOneById_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return { data, meta };
    };

    createOne = (response: ApiHttpResponse): ImService_CreateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return { data, meta };
    };

    updateOne = (response: ApiHttpResponse): ImService_UpdateOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    updateStatus = (response: ApiHttpResponse): ImService_UpdateStatus_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    deleteOne = (response: ApiHttpResponse): ImService_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    testSendMsg = (response: ApiHttpResponse): ImService_TestSendMsg_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
