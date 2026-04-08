import { type AxiosResponse } from "axios";
import { z } from "zod";
import type { ProjectNetworks_FindManyPaginated_Res } from "~/projects/api/services/projects-services";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const ProjectNetworkSchema = z.object({
    id: z.string(),
    name: z.string(),
    availableInProjects: z.boolean(),
    driver: z.string(),
    internal: z.boolean(),
    attachable: z.boolean(),
    ingress: z.boolean(),
    enableIPv4: z.boolean(),
    enableIPv6: z.boolean(),
    options: z
        .record(z.string())
        .nullish()
        .default({})
        .transform(value => value ?? {}),
    labels: z
        .record(z.string())
        .nullish()
        .default({})
        .transform(value => value ?? {}),
    createdAt: z.coerce.date(),
});

const FindManyPaginatedSchema = z.object({
    data: z.array(ProjectNetworkSchema),
    meta: PagingMetaApiSchema,
});

export class ProjectNetworksApiValidator {
    findManyPaginated = (response: AxiosResponse): ProjectNetworks_FindManyPaginated_Res => {
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
