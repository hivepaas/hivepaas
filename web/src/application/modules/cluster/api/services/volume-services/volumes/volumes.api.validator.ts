import type { AxiosResponse } from "axios";
import { z } from "zod";
import { EClusterVolumePropagation } from "~/cluster/module-shared/enums";
import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    ClusterVolumes_CreateOne_Res,
    ClusterVolumes_FindManyPaginated_Res,
    ClusterVolumes_FindOneById_Res,
    ClusterVolumes_List_Res,
    ClusterVolumes_SyncFromDocker_Res,
    ClusterVolumes_UpdateOne_Res,
    ClusterVolumes_UpdateStatus_Res,
} from "./volumes.api.contracts";

const mapSchema = z
    .record(z.string())
    .nullish()
    .default({})
    .transform(value => value ?? {});

const optionalStringSchema = z
    .string()
    .nullish()
    .transform(value => value ?? undefined);

const BindOptionsSchema = z
    .object({
        directory: optionalStringSchema,
        propagation: z
            .nativeEnum(EClusterVolumePropagation)
            .nullish()
            .transform(value => value ?? undefined),
        readonly: z
            .boolean()
            .nullish()
            .transform(value => value ?? undefined),
        extraOptions: optionalStringSchema,
    })
    .nullish()
    .transform(value => value ?? null);

const NfsOptionsSchema = z
    .object({
        addr: optionalStringSchema,
        device: optionalStringSchema,
        version: optionalStringSchema,
        readonly: z
            .boolean()
            .nullish()
            .transform(value => value ?? undefined),
        extraOptions: optionalStringSchema,
    })
    .nullish()
    .transform(value => value ?? null);

const TmpfsOptionsSchema = z
    .object({
        device: optionalStringSchema,
        size: z
            .union([z.string(), z.number()])
            .nullish()
            .transform(value => (value === null || value === undefined ? undefined : String(value))),
        uid: z
            .union([z.string(), z.number()])
            .nullish()
            .transform(value => (value === null || value === undefined ? undefined : String(value))),
        gid: z
            .union([z.string(), z.number()])
            .nullish()
            .transform(value => (value === null || value === undefined ? undefined : String(value))),
        mode: optionalStringSchema,
        extraOptions: optionalStringSchema,
    })
    .nullish()
    .transform(value => value ?? null);

const BtrfsOptionsSchema = z
    .object({
        device: optionalStringSchema,
        subvol: optionalStringSchema,
        readonly: z
            .boolean()
            .nullish()
            .transform(value => value ?? undefined),
        extraOptions: optionalStringSchema,
    })
    .passthrough()
    .nullish()
    .transform(value => value ?? null);

export const ClusterVolumeSchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    driver: z.string(),
    mountpoint: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    labels: mapSchema,
    options: mapSchema,
    scope: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    refCount: z
        .number()
        .nullish()
        .transform(value => value ?? 0),
    size: z
        .number()
        .nullish()
        .transform(value => value ?? 0),
    clusterVolumeSpec: z.unknown().nullish().default(null),
    bindOptions: BindOptionsSchema,
    nfsOptions: NfsOptionsSchema,
    tmpfsOptions: TmpfsOptionsSchema,
    btrfsOptions: BtrfsOptionsSchema,
});

const FindManyPaginatedSchema = z.object({
    data: z.array(ClusterVolumeSchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: ClusterVolumeSchema,
});

const IdSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class ClusterVolumesApiValidator {
    findManyPaginated = (response: AxiosResponse): ClusterVolumes_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return {
            data,
            meta,
        };
    };

    list = (response: AxiosResponse): ClusterVolumes_List_Res => this.findManyPaginated(response);

    findOneById = (response: AxiosResponse): ClusterVolumes_FindOneById_Res => {
        const { data } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return {
            data,
        };
    };

    createOne = (response: AxiosResponse): ClusterVolumes_CreateOne_Res => {
        const { data } = parseApiResponse({
            response,
            schema: IdSchema,
        });

        return {
            data,
        };
    };

    updateOne = (response: AxiosResponse): ClusterVolumes_UpdateOne_Res => {
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

    updateStatus = (response: AxiosResponse): ClusterVolumes_UpdateStatus_Res => {
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

    syncFromDocker = (response: AxiosResponse): ClusterVolumes_SyncFromDocker_Res => {
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
