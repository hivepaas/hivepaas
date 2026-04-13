import { type AxiosResponse } from "axios";
import { z } from "zod";
import { EMountConsistency, EMountType } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import {
    type AppStorageSettings_FindOne_Res,
    type AppStorageSettings_UpdateOne_Res,
} from "./app-storage-settings.api.contracts";

const MountSchema = z.object({
    type: z.nativeEnum(EMountType).optional(),
    source: z.string().optional(),
    target: z.string().optional(),
    readOnly: z.boolean().optional(),
    consistency: z.nativeEnum(EMountConsistency).optional(),
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
                        source: item.source,
                        target: item.target,
                        readOnly: item.readOnly,
                        consistency: item.consistency,
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
