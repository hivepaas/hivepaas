import { type AxiosResponse } from "axios";
import { z } from "zod";

import { type Public_Apps_FindManyBase_Res } from "@application/shared/api-public/services";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * Find many base API response schema (base-list).
 */
const FindManyBaseSchema = z.object({
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
     * Validate and transform find many public apps base API response.
     */
    findManyBase = (response: AxiosResponse): Public_Apps_FindManyBase_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyBaseSchema,
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
