import type { AxiosResponse } from "axios";
import { z } from "zod";
import type {
    ClusterNetworks_CreateOne_Res,
    ClusterNetworks_FindManyPaginated_Res,
    ClusterNetworks_FindOneById_Res,
    ClusterNetworks_SyncFromDocker_Res,
    ClusterNetworks_UpdateOne_Res,
    ClusterNetworks_UpdateStatus_Res,
} from "~/cluster/api/services/network-services";

import { ESettingStatus } from "@application/shared/enums";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const mapSchema = z
    .record(z.string())
    .nullish()
    .default({})
    .transform(value => value ?? {});

const optionalBooleanSchema = z
    .boolean()
    .nullish()
    .transform(value => value ?? false);

export const ClusterNetworkSchema = z.object({
    id: z.string(),
    type: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    name: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    kind: z.string().optional(),
    status: z.nativeEnum(ESettingStatus),
    inherited: optionalBooleanSchema,
    availableInProjects: optionalBooleanSchema,
    default: optionalBooleanSchema,
    updateVer: z.number(),
    size: z
        .number()
        .nullish()
        .transform(value => value ?? 0),
    driver: z.string(),
    internal: z.boolean(),
    attachable: z.boolean(),
    ingress: z.boolean(),
    enableIPv4: z.boolean(),
    enableIPv6: z.boolean(),
    options: mapSchema,
    labels: mapSchema,
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullish(),
    expireAt: z.coerce.date().nullish(),
});

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

    updateOne = (response: AxiosResponse): ClusterNetworks_UpdateOne_Res => {
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

    updateStatus = (response: AxiosResponse): ClusterNetworks_UpdateStatus_Res => {
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

    syncFromDocker = (response: AxiosResponse): ClusterNetworks_SyncFromDocker_Res => {
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
