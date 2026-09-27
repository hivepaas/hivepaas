import { type AxiosResponse } from "axios";
import { z } from "zod";

import { type Public_Projects_FindManyPaginated_Res } from "@application/shared/api-public/services";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * The fields a picker needs from GET /projects; the rest of each project is
 * dropped.
 */
const FindManyPaginatedSchema = z.object({
    data: z.array(
        z.object({
            id: z.string(),
            name: z.string(),
            envs: z
                .array(
                    z.object({
                        id: z.string(),
                        name: z.string(),
                        color: z.string(),
                    }),
                )
                .nullish(),
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
                envs: project.envs ?? [],
            })),
            meta,
        };
    };
}
