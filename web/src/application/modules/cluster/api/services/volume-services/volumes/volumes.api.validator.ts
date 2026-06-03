import type { AxiosResponse } from "axios";
import { z } from "zod";

import { PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { ClusterVolumes_List_Res } from "./volumes.api.contracts";

const ClusterVolumeSchema = z.object({
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
    data: z.array(ClusterVolumeSchema),
    meta: PagingMetaApiSchema,
});

export class ClusterVolumesApiValidator {
    list = (response: AxiosResponse): ClusterVolumes_List_Res => {
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
