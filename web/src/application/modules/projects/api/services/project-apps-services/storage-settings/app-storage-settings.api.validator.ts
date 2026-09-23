import { type AxiosResponse } from "axios";
import { z } from "zod";
import { EMountConsistency, EMountPropagation, EMountType } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import {
    type AppStorageSettings_FindOne_Res,
    type AppStorageSettings_Preflight_Res,
} from "./app-storage-settings.api.contracts";

const VolumeDriverSchema = z.object({
    name: z.string().optional(),
    options: z
        .record(z.string())
        .nullish()
        .transform(rec => rec ?? ({} as Record<string, string>)),
});

const BindOptionsSchema = z.object({
    propagation: z.nativeEnum(EMountPropagation).optional(),
    nonRecursive: z.boolean().optional(),
    createMountpoint: z.boolean().optional(),
    readOnlyNonRecursive: z.boolean().optional(),
    readOnlyForceRecursive: z.boolean().optional(),
});

const VolumeOptionsSchema = z.object({
    subpath: z.string().optional(),
    noCopy: z.boolean().optional(),
    labels: z
        .record(z.string())
        .nullish()
        .transform(rec => rec ?? ({} as Record<string, string>)),
    driverConfig: VolumeDriverSchema.nullish(),
});

const TmpfsOptionsSchema = z.object({
    size: z.string().optional(),
    mode: z.string().optional(),
    options: z.array(z.array(z.string())).nullish(),
});

const ClusterOptionsSchema = z.object({
    subpath: z.string().optional(),
    noCopy: z.boolean().optional(),
    labels: z
        .record(z.string())
        .nullish()
        .transform(rec => rec ?? ({} as Record<string, string>)),
    driverConfig: VolumeDriverSchema.nullish(),
});

const MountSourceAppSchema = z.object({
    appId: z.string().optional(),
    write: z.boolean().optional(),
    name: z.string().optional(),
    dangling: z.boolean().optional(),
});

const MountSchema = z.object({
    key: z.string().optional(),
    type: z.nativeEnum(EMountType).optional(),
    source: z.string().optional(),
    target: z.string().optional(),
    readOnly: z.boolean().optional(),
    consistency: z.nativeEnum(EMountConsistency).optional(),
    bindOptions: BindOptionsSchema.optional(),
    volumeOptions: VolumeOptionsSchema.optional(),
    tmpfsOptions: TmpfsOptionsSchema.optional(),
    clusterOptions: ClusterOptionsSchema.optional(),
    sourceApp: MountSourceAppSchema.nullish(),
});

const MountBorrowerSchema = z.object({
    appId: z.string(),
    name: z.string(),
    target: z.string(),
    subpath: z.string().optional(),
    write: z.boolean().optional(),
});

const AppStorageSettingsSchema = z.object({
    mounts: z.array(MountSchema).nullish(),
    borrowedBy: z.array(MountBorrowerSchema).nullish(),
    updateVer: z.number(),
});

const StorageFindingSchema = z.object({
    target: z.string().catch(""),
    volume: z.object({ id: z.string().catch(""), name: z.string().catch("") }),
    path: z.string().catch(""),
});

const PreflightSchema = z.object({
    data: z.object({
        storage: z.array(StorageFindingSchema).nullish(),
        storageUnchecked: z.array(StorageFindingSchema).nullish(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const FindOneSchema = z.object({
    data: AppStorageSettingsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

export class AppStorageSettingsApiValidator {
    preflight = (response: AxiosResponse): AppStorageSettings_Preflight_Res => {
        const { data, meta } = parseApiResponse({ response, schema: PreflightSchema });
        return { data: { storage: data.storage ?? [], unchecked: data.storageUnchecked ?? [] }, meta };
    };

    findOne = (response: AxiosResponse): AppStorageSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return {
            data: {
                mounts:
                    data.mounts?.map(item => ({
                        key: item.key,
                        type: item.type,
                        source: item.source,
                        target: item.target,
                        readOnly: item.readOnly,
                        consistency: item.consistency,
                        bindOptions: item.bindOptions,
                        volumeOptions: item.volumeOptions,
                        tmpfsOptions: item.tmpfsOptions,
                        clusterOptions: item.clusterOptions,
                        sourceApp: item.sourceApp ?? undefined,
                    })) ?? [],
                borrowedBy: data.borrowedBy ?? [],
                updateVer: data.updateVer,
            },
            meta,
        };
    };
}
