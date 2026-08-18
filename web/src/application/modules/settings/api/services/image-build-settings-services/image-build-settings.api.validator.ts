import { type AxiosResponse } from "axios";
import { z } from "zod";
import {
    ImageBuildRepoCacheClearResultSchema,
    ImageBuildRepoCacheInfoSchema,
    ImageBuildSettingsEntitySchema,
} from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    ImageBuildSettings_ClearRepoCache_Res,
    ImageBuildSettings_FindOne_Res,
    ImageBuildSettings_FindRepoCache_Res,
} from "./image-build-settings.api.contracts";

const NODE_DISPLAY_PREFIX = " (node: ";

const FindOneSchema = z.object({
    data: ImageBuildSettingsEntitySchema,
    meta: BaseMetaApiSchema.nullable(),
});

const FindRepoCacheSchema = z.object({
    data: ImageBuildRepoCacheInfoSchema,
    meta: BaseMetaApiSchema.nullable(),
});

const ClearRepoCacheSchema = z.object({
    data: ImageBuildRepoCacheClearResultSchema,
    meta: BaseMetaApiSchema.nullable(),
});

function parseWorkerNodeName(id: string, name: string): string {
    const prefix = `${id}${NODE_DISPLAY_PREFIX}`;
    if (name.startsWith(prefix) && name.endsWith(")")) {
        return name.slice(prefix.length, -1);
    }

    return name;
}

export class ImageBuildSettingsApiValidator {
    findOne = (response: AxiosResponse): ImageBuildSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });

        return {
            data: {
                ...data,
                expireAt: data.expireAt ?? null,
                workers: {
                    nodes: (data.workers?.nodes ?? []).map(node => ({
                        id: node.id,
                        name: parseWorkerNodeName(node.id, node.name),
                    })),
                    nodeLabels: data.workers?.nodeLabels ?? [],
                    maxParallelism: data.workers?.maxParallelism ?? 0,
                },
                resources: {
                    cpus: data.resources?.cpus,
                    mem: data.resources?.mem,
                    memSwap: data.resources?.memSwap,
                    shmSize: data.resources?.shmSize,
                },
                sources: {
                    repoCache: data.sources?.repoCache ?? false,
                },
                noCache: data.noCache ?? false,
                noVerbose: data.noVerbose ?? false,
            },
            meta,
        };
    };

    findRepoCache = (response: AxiosResponse): ImageBuildSettings_FindRepoCache_Res => {
        return parseApiResponse({ response, schema: FindRepoCacheSchema });
    };

    clearRepoCache = (response: AxiosResponse): ImageBuildSettings_ClearRepoCache_Res => {
        return parseApiResponse({ response, schema: ClearRepoCacheSchema });
    };
}
