import type { ProjectImageBuildSettings } from "~/projects/domain";

import type { ProjectImageBuildSettingsFormSchemaInput } from "../schemas";

export function mapProjectImageBuildSettingsToFormInput(
    data: ProjectImageBuildSettings,
): ProjectImageBuildSettingsFormSchemaInput {
    return {
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
