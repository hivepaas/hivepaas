import { type AxiosResponse } from "axios";
import { z } from "zod";

import { SettingsBaseEntitySchema } from "~/settings/module-shared/schemas";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { GitCredentials_FindManyPaginated_Res } from "./git-credentials.api.contracts";

const GitCredentialSettingEntitySchema = SettingsBaseEntitySchema.omit({ description: true }).extend({
    description: z.string().optional(),
    kind: z.string().optional(),
    inherited: z.boolean().optional(),
});

const FindManyPaginatedSchema = z.object({
    data: z.array(GitCredentialSettingEntitySchema),
    meta: PagingMetaApiSchema,
});

export class GitCredentialsApiValidator {
    findManyPaginated = (response: AxiosResponse): GitCredentials_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return {
            data,
            meta,
        };
    };
}
