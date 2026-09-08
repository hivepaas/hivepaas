import type { AxiosResponse } from "axios";
import { z } from "zod";
import {
    HivePaaSServiceSettingsEntitySchema,
    SettingsPendingChangeSchema,
} from "~/system-settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    HivePaaSServiceSettings_ConfirmChange_Res,
    HivePaaSServiceSettings_FindOne_Res,
    HivePaaSServiceSettings_RevertChange_Res,
    HivePaaSServiceSettings_UpdateOne_Res,
} from "./hivepaas-service-settings.api.contracts";

const FindOneSchema = z.object({
    data: HivePaaSServiceSettingsEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

const UpdateOneSchema = z.object({
    data: SettingsPendingChangeSchema.nullish(),
    meta: BaseMetaApiSchema.nullish(),
});

const RevertChangeSchema = z.object({
    data: z
        .object({
            reverted: z.boolean(),
            reason: z.string().nullish(),
        })
        .nullish(),
    meta: BaseMetaApiSchema.nullish(),
});

export class HivePaaSServiceSettingsApiValidator {
    findOne = (response: AxiosResponse): HivePaaSServiceSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data: { ...data, pendingChange: data.pendingChange ?? null }, meta };
    };

    updateOne = (response: AxiosResponse): HivePaaSServiceSettings_UpdateOne_Res => {
        const { data } = parseApiResponse({ response, schema: UpdateOneSchema });
        return { data: { pendingChange: data ?? null } };
    };

    confirmChange = (response: AxiosResponse): HivePaaSServiceSettings_ConfirmChange_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    revertChange = (response: AxiosResponse): HivePaaSServiceSettings_RevertChange_Res => {
        const { data } = parseApiResponse({ response, schema: RevertChangeSchema });
        return { data: { reverted: data?.reverted ?? false, reason: data?.reason ?? null } };
    };
}
