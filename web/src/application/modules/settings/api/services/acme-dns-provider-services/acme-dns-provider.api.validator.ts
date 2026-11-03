import { z } from "zod";
import { AcmeDnsProviderSettingEntitySchema } from "~/settings/module-shared/schemas";

import { type ApiHttpResponse, BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    AcmeDnsProvider_CreateOne_Res,
    AcmeDnsProvider_DeleteOne_Res,
    AcmeDnsProvider_FindManyPaginated_Res,
    AcmeDnsProvider_FindOneById_Res,
    AcmeDnsProvider_TestAccess_Res,
    AcmeDnsProvider_UpdateOne_Res,
    AcmeDnsProvider_UpdateStatus_Res,
} from "./acme-dns-provider.api.contracts";

const FindManyPaginatedSchema = z.object({
    data: z.array(AcmeDnsProviderSettingEntitySchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: AcmeDnsProviderSettingEntitySchema,
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

export class AcmeDnsProviderApiValidator {
    findManyPaginated = (response: ApiHttpResponse): AcmeDnsProvider_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return { data, meta };
    };

    findOneById = (response: ApiHttpResponse): AcmeDnsProvider_FindOneById_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return { data, meta };
    };

    createOne = (response: ApiHttpResponse): AcmeDnsProvider_CreateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return { data, meta };
    };

    updateOne = (response: ApiHttpResponse): AcmeDnsProvider_UpdateOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    updateStatus = (response: ApiHttpResponse): AcmeDnsProvider_UpdateStatus_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    deleteOne = (response: ApiHttpResponse): AcmeDnsProvider_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    testAccess = (response: ApiHttpResponse): AcmeDnsProvider_TestAccess_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
