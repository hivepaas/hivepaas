import { z } from "zod";
import { ClusterNetworkSchema } from "~/cluster/api/services";
import type {
    ProjectNetworks_CreateOne_Res,
    ProjectNetworks_DeleteOne_Res,
    ProjectNetworks_FindManyPaginated_Res,
    ProjectNetworks_FindOneById_Res,
    ProjectNetworks_UpdateOne_Res,
    ProjectNetworks_UpdateStatus_Res,
} from "~/projects/api/services/projects-services";

import { type ApiHttpResponse, BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const FindManyPaginatedSchema = z.object({
    data: z.array(ClusterNetworkSchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: ClusterNetworkSchema,
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

export class ProjectNetworksApiValidator {
    findManyPaginated = (response: ApiHttpResponse): ProjectNetworks_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return {
            data,
            meta,
        };
    };

    findOneById = (response: ApiHttpResponse): ProjectNetworks_FindOneById_Res => {
        const { data } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return {
            data,
        };
    };

    createOne = (response: ApiHttpResponse): ProjectNetworks_CreateOne_Res => {
        const { data } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return {
            data,
        };
    };

    updateOne = (response: ApiHttpResponse): ProjectNetworks_UpdateOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return {
            data: {
                type: "success",
            },
        };
    };

    updateStatus = (response: ApiHttpResponse): ProjectNetworks_UpdateStatus_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return {
            data: {
                type: "success",
            },
        };
    };

    deleteOne = (response: ApiHttpResponse): ProjectNetworks_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return {
            data: {
                networkID: "",
            },
        };
    };
}
