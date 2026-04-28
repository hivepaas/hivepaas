import { type AxiosResponse } from "axios";
import { z } from "zod";
import { EMountConsistency, EMountPropagation, EMountType } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import {
    type AppStorageSettings_FindOne_Res,
    type AppStorageSettings_UpdateOne_Res,
} from "./app-storage-settings.api.contracts";

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
    labels: z.record(z.string()).optional(),
    driverConfig: VolumeDriverSchema.optional(),
});

const TmpfsOptionsSchema = z.object({
    size: z.string().optional(),
    mode: z.number().optional(),
    options: z.array(z.array(z.string())).optional(),
});

const ClusterOptionsSchema = z.object({
    volume: z.string().optional(),
    subpath: z.string().optional(),
    subpathRequired: z.string().optional(),
    noCopy: z.boolean().optional(),
    labels: z.record(z.string()).optional(),
    driverConfig: VolumeDriverSchema.optional(),
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
    updateVer: z.number(),
});

const FindOneSchema = z.object({
    data: AppStorageSettingsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

const UpdateOneSchema = z.object({
    data: z.object({ type: z.literal("success") }),
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
                updateVer: data.updateVer,
            },
            meta,
        };
    };

    updateOne = (response: AxiosResponse): AppStorageSettings_UpdateOne_Res => {
        return parseApiResponse({ response, schema: UpdateOneSchema });
    };
}
