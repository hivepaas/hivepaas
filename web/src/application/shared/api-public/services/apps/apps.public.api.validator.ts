import { z } from "zod";

import { type Public_Apps_FindMany_Res } from "@application/shared/api-public/services";

import { type ApiHttpResponse, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * The fields a picker needs from GET /projects/{projectID}[/{env}]/apps; the rest of
 * each app is dropped.
 */
const FindManySchema = z.object({
    data: z.array(
        z.object({
            id: z.string(),
            name: z.string(),
            env: z.string().optional(),
        }),
    ),
    meta: PagingMetaApiSchema,
});

export class AppsPublicApiValidator {
    /**
     * Validate and transform find many public apps API response.
     */
    findMany = (response: ApiHttpResponse): Public_Apps_FindMany_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManySchema,
        });

        return {
            data: data.map(app => ({
                id: app.id,
                name: app.name,
                env: app.env,
            })),
            meta,
        };
    };
}
