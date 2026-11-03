import type { ImageBuildSettings } from "~/settings/domain";

import type { SettingsImageBuildFormSchemaInput } from "../schemas";

export function mapSettingsImageBuildToFormInput(data: ImageBuildSettings): SettingsImageBuildFormSchemaInput {
    return {
        workers: {
            nodes: data.workers.nodes,
            nodeLabels: data.workers.nodeLabels,
            maxParallelism: data.workers.maxParallelism,
        },
        resources: {
            cpus: data.resources.cpus,
            mem: data.resources.mem,
            memSwap: data.resources.memSwap,
            shmSize: data.resources.shmSize,
        },
        sources: {
            repoCache: data.sources.repoCache,
        },
        noCache: data.noCache,
        noVerbose: data.noVerbose,
    };
}
