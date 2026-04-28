import { type AxiosResponse } from "axios";
import { z } from "zod";
import type { ProjectDockerVolumes_List_Res } from "~/projects/api/services/projects-services";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const DockerVolumeSchema = z.object({
    id: z.string(),
    name: z.string(),
    driver: z.string(),
    createdAt: z.coerce.date(),
    mountpoint: z.string(),
    labels: z
        .record(z.string())
        .nullish()
        .default({})
        .transform(value => value ?? {}),
    options: z
        .record(z.string())
        .nullish()
        .default({})
        .transform(value => value ?? {}),
    scope: z.string(),
    availableInProjects: z.boolean(),
    refCount: z.number(),
    size: z.number(),
    updateVer: z.number(),
    status: z
        .record(z.unknown())
        .nullish()
        .default({})
        .transform(value => value ?? {}),
    clusterVolumeSpec: z.unknown().nullish().default(null),
});

const ListSchema = z.object({
    data: z.array(DockerVolumeSchema),
    meta: PagingMetaApiSchema,
});

export class ProjectDockerVolumesApiValidator {
    list = (response: AxiosResponse): ProjectDockerVolumes_List_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: ListSchema,
        });

        return {
            data,
            meta,
        };
    };
}
