import { type AxiosResponse } from "axios";
import { z } from "zod";
import { EMountConsistency, EMountPropagation, EMountType } from "~/projects/module-shared/enums";

import { StorageSettingsEntitySchema } from "@application/modules/settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import { type AppStorageSettings_FindOne_Res } from "./app-storage-settings.api.contracts";

const VolumeDriverSchema = z.object({
    name: z.string().optional(),
    options: z.record(z.string()).optional(),
});

const BindOptionsSchema = z.object({
    baseDir: z.string().optional(),
    subpath: z.string().optional(),
    subpathRequired: z.string().optional(),
    propagation: z.nativeEnum(EMountPropagation).optional(),
    nonRecursive: z.boolean().optional(),
    createMountpoint: z.boolean().optional(),
    readOnlyNonRecursive: z.boolean().optional(),
    readOnlyForceRecursive: z.boolean().optional(),
});

const VolumeOptionsSchema = z.object({
    volume: z.string().optional(),
    subpath: z.string().optional(),
    subpathRequired: z.string().optional(),
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
    volume: z.string().optional(),
    subpath: z.string().optional(),
    subpathRequired: z.string().optional(),
    noCopy: z.boolean().optional(),
    labels: z
        .record(z.string())
        .nullish()
        .transform(rec => rec ?? ({} as Record<string, string>)),
    driverConfig: VolumeDriverSchema.nullish(),
});

const MountSchema = z.object({
    type: z.nativeEnum(EMountType).optional(),
    target: z.string().optional(),
    readOnly: z.boolean().optional(),
    consistency: z.nativeEnum(EMountConsistency).optional(),
    bindOptions: BindOptionsSchema.optional(),
    volumeOptions: VolumeOptionsSchema.optional(),
    tmpfsOptions: TmpfsOptionsSchema.optional(),
    clusterOptions: ClusterOptionsSchema.optional(),
});

const AppStorageSettingsSchema = z.object({
    mounts: z.array(MountSchema).nullish(),
    settings: StorageSettingsEntitySchema,
    updateVer: z.number(),
});

const FindOneSchema = z.object({
    data: AppStorageSettingsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

export class AppStorageSettingsApiValidator {
    findOne = (response: AxiosResponse): AppStorageSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return {
            data: {
                mounts:
                    data.mounts?.map(item => ({
                        type: item.type,
                        target: item.target,
                        readOnly: item.readOnly,
                        consistency: item.consistency,
                        bindOptions: item.bindOptions,
                        volumeOptions: item.volumeOptions,
                        tmpfsOptions: item.tmpfsOptions,
                        clusterOptions: item.clusterOptions,
                    })) ?? [],
                settings: data.settings,
                updateVer: data.updateVer,
            },
            meta,
        };
    };
}
