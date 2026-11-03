import { z } from "zod";
import { DomainSettingsEntitySchema } from "~/settings/module-shared/schemas";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    DomainSettings_DeleteOne_Res,
    DomainSettings_FindOne_Res,
    DomainSettings_UpdateOne_Res,
    DomainSettings_UpdateStatus_Res,
} from "./domain-settings.api.contracts";

const FindOneSchema = z.object({
    data: DomainSettingsEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class DomainSettingsApiValidator {
    findOne = (response: ApiHttpResponse): DomainSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneSchema,
        });

        return { data, meta };
    };

    updateOne = (response: ApiHttpResponse): DomainSettings_UpdateOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    updateStatus = (response: ApiHttpResponse): DomainSettings_UpdateStatus_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    deleteOne = (response: ApiHttpResponse): DomainSettings_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
