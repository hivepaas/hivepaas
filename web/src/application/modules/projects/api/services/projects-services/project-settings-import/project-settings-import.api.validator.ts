import { z } from "zod";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { ProjectSettingsImport_Import_Res } from "./project-settings-import.api.contracts";

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class ProjectSettingsImportApiValidator {
    importSettings = (response: ApiHttpResponse): ProjectSettingsImport_Import_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
