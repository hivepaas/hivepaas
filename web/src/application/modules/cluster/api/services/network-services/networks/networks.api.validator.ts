import type { AxiosResponse } from "axios";
import { z } from "zod";
import type {
    ClusterNetworks_CreateOne_Res,
    ClusterNetworks_FindManyPaginated_Res,
    ClusterNetworks_FindOneById_Res,
} from "~/cluster/api/services/network-services";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const mapSchema = z
    .record(z.string())
    .nullish()
    .default({})
    .transform(value => value ?? {});

const NetworkSchema = z.object({
    id: z.string(),
    name: z.string(),
    availableInProjects: z.boolean(),
    driver: z.string(),
    internal: z.boolean(),
    attachable: z.boolean(),
    ingress: z.boolean(),
    enableIPv4: z.boolean(),
    enableIPv6: z.boolean(),
    options: mapSchema,
    labels: mapSchema,
    createdAt: z.coerce.date(),
});

const FindManyPaginatedSchema = z.object({
    data: z.array(NetworkSchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: NetworkSchema,
});

const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
});

export class ClusterNetworksApiValidator {
    findManyPaginated = (response: AxiosResponse): ClusterNetworks_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return {
            data,
            meta,
        };
    };

    findOneById = (response: AxiosResponse): ClusterNetworks_FindOneById_Res => {
        const { data } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return {
            data,
        };
    };

    createOne = (response: AxiosResponse): ClusterNetworks_CreateOne_Res => {
        const { data } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return {
            data,
        };
    };
}
