import { type AxiosResponse } from "axios";
import { z } from "zod";
import { AppPlacementSettingsEntitySchema } from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { AppPlacementSettings_FindOne_Res } from "./app-placement-settings.api.contracts";

const FindOneSchema = z.object({
    data: AppPlacementSettingsEntitySchema,
    meta: BaseMetaApiSchema.nullable(),
});

export class AppPlacementSettingsApiValidator {
    findOne = (response: AxiosResponse): AppPlacementSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });

        return {
            data: {
                ...data,
                expireAt: data.expireAt ?? null,
                excludeManagerNodes: data.excludeManagerNodes ?? false,
                excludeBuildNodes: data.excludeBuildNodes ?? false,
            },
            meta,
        };
    };
}
