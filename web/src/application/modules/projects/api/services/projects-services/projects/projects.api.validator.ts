import { type AxiosResponse } from "axios";
import { z } from "zod";
import { type Projects_FindManyPaginated_Res } from "~/projects/api/services/projects-services/projects/projects.api.contracts";
import { EProjectStatus } from "~/projects/module-shared/enums";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * Project schema
 */
const ProjectSchema = z.object({
    id: z.string(),
    name: z.string(),
    key: z.string(),
    status: z.nativeEnum(EProjectStatus),
    photo: z.string(),
    note: z.string(),
    tags: z.array(z.string()),
    updateVer: z.number(),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullable(),
});

/**
 * Find many projects paginated API response schema
 */
const FindManyPaginatedSchema = z.object({
    data: z.array(ProjectSchema),
    meta: PagingMetaApiSchema,
});

export class ProjectsApiValidator {
    /**
     * Validate and transform find many projects paginated API response
     */
    findManyPaginated = (response: AxiosResponse): Projects_FindManyPaginated_Res => {
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
