import { type AxiosResponse } from "axios";
import { z } from "zod";

import { type Public_Apps_FindMany_Res } from "@application/shared/api-public/services";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * The fields a picker needs from GET /projects/{projectID}/apps; the rest of
 * each app is dropped.
 */
const FindManySchema = z.object({
    data: z.array(
        z.object({
            id: z.string(),
            name: z.string(),
        }),
    ),
    meta: PagingMetaApiSchema,
});

export class AppsPublicApiValidator {
    /**
     * Validate and transform find many public apps API response.
     */
    findMany = (response: AxiosResponse): Public_Apps_FindMany_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManySchema,
        });

        return {
            data: data.map(app => ({
                id: app.id,
                name: app.name,
            })),
            meta,
        };
    };
}
