import { z } from "zod";
import {
    HivePaaSServiceSettingsEntitySchema,
    SettingsPendingChangeSchema,
} from "~/system-settings/module-shared/schemas";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

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
    findOne = (response: ApiHttpResponse): HivePaaSServiceSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data: { ...data, pendingChange: data.pendingChange ?? null }, meta };
    };

    updateOne = (response: ApiHttpResponse): HivePaaSServiceSettings_UpdateOne_Res => {
        const { data } = parseApiResponse({ response, schema: UpdateOneSchema });
        return { data: { pendingChange: data ?? null } };
    };

    confirmChange = (response: ApiHttpResponse): HivePaaSServiceSettings_ConfirmChange_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    revertChange = (response: ApiHttpResponse): HivePaaSServiceSettings_RevertChange_Res => {
        const { data } = parseApiResponse({ response, schema: RevertChangeSchema });
        return { data: { reverted: data?.reverted ?? false, reason: data?.reason ?? null } };
    };
}
