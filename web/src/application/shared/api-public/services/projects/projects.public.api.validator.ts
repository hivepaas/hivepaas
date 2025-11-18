import { type AxiosResponse } from "axios";
import { z } from "zod";

import { type Public_Projects_FindManyPaginated_Res } from "@application/shared/api-public/services";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * Find many paginated API response schema (base-list).
 */
const FindManyPaginatedSchema = z.object({
    data: z.array(
        z.object({
            id: z.string(),
            name: z.string(),
        }),
    ),
    meta: PagingMetaApiSchema,
});

export class ProjectsPublicApiValidator {
    /**
     * Validate and transform find many public projects API response.
     */
    findManyPaginated = (response: AxiosResponse): Public_Projects_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return {
            data: data.map(project => ({
                id: project.id,
                name: project.name,
            })),
            meta,
        };
    };
}
