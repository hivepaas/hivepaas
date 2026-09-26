import { type AppDockerApiSettings_UpdatePayload } from "~/projects/api/services";
import { type AppDockerApiSettings } from "~/projects/domain";

import { type AppConfigDockerApiFormSchemaInput, type AppConfigDockerApiFormSchemaOutput } from "../schemas";

export function mapAppDockerApiSettingsToFormInput(data: AppDockerApiSettings): AppConfigDockerApiFormSchemaInput {
    return {
        enabled: data.enabled,
        mode: data.mode,
        images: data.images.map(value => ({ value })),
        sharedDirs: data.sharedDirs.map(value => ({ value })),
        sharedVolumes: Object.entries(data.sharedVolumes).map(([key, value]) => ({ key, value })),
        envNetwork: data.networks.includes("env"),
        allow: data.allow,
        limits: {
            // Zero is the default, which the fields show as their placeholder.
            containers: data.limits.containers > 0 ? data.limits.containers : undefined,
            memory: data.limits.memory !== "" ? data.limits.memory : undefined,
            cpus: data.limits.cpus > 0 ? data.limits.cpus : undefined,
        },
    };
}

/** The whole new state, as the server takes it: a limit left empty keeps the default. */
export function mapDockerApiFormValuesToPayload(
    values: AppConfigDockerApiFormSchemaOutput,
    updateVer: number,
): AppDockerApiSettings_UpdatePayload {
    const memory = values.limits.memory?.trim();
    return {
        enabled: values.enabled,
        mode: values.mode,
        images: values.images.map(item => item.value),
        sharedDirs: values.sharedDirs.map(item => item.value),
        sharedVolumes: Object.fromEntries(values.sharedVolumes.map(item => [item.key, item.value])),
        networks: values.envNetwork ? ["env"] : [],
        allow: values.allow,
        limits: {
            containers: values.limits.containers,
            memory: memory === "" ? undefined : memory,
            cpus: values.limits.cpus,
        },
        updateVer,
    };
}
