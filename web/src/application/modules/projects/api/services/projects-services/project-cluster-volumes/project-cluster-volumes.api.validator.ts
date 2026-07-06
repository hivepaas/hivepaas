import { type AxiosResponse } from "axios";
import { z } from "zod";
import { ClusterVolumeSchema } from "~/cluster/api/services";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    ProjectClusterVolumes_CreateOne_Res,
    ProjectClusterVolumes_DeleteOne_Res,
    ProjectClusterVolumes_FindManyPaginated_Res,
    ProjectClusterVolumes_FindOneById_Res,
    ProjectClusterVolumes_UpdateOne_Res,
    ProjectClusterVolumes_UpdateStatus_Res,
} from "./project-cluster-volumes.api.contracts";

const FindManyPaginatedSchema = z.object({
    data: z.array(ClusterVolumeSchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: ClusterVolumeSchema,
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

export class ProjectClusterVolumesApiValidator {
    findManyPaginated = (response: AxiosResponse): ProjectClusterVolumes_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return {
            data,
            meta,
        };
    };

    findOneById = (response: AxiosResponse): ProjectClusterVolumes_FindOneById_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return {
            data,
            meta,
        };
    };

    createOne = (response: AxiosResponse): ProjectClusterVolumes_CreateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return {
            data,
            meta,
        };
    };

    updateOne = (response: AxiosResponse): ProjectClusterVolumes_UpdateOne_Res => {
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

    updateStatus = (response: AxiosResponse): ProjectClusterVolumes_UpdateStatus_Res => {
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

    deleteOne = (response: AxiosResponse): ProjectClusterVolumes_DeleteOne_Res => {
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
}
