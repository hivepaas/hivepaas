import { type AppResourceSettings } from "~/projects/domain";

import { type AppConfigResourcesFormSchemaInput } from "../schemas";

export function mapAppResourceSettingsToFormInput(data: AppResourceSettings): AppConfigResourcesFormSchemaInput {
    return {
        reservations: {
            cpus: data.reservations?.cpus,
            memoryMB: data.reservations?.memoryMB,
            genericResources: (data.reservations?.genericResources ?? []).map(item => ({
                kind: item.kind,
                value: item.value,
            })),
        },
        limits: {
            cpus: data.limits?.cpus,
            memoryMB: data.limits?.memoryMB,
            pids: data.limits?.pids,
        },
        ulimits: data.ulimits.map(item => ({
            name: item.name,
            hard: item.hard,
            soft: item.soft,
        })),
        capabilities: {
            capabilityAdd: data.capabilities?.capabilityAdd ?? [],
            capabilityDrop: data.capabilities?.capabilityDrop ?? [],
            enableGPU: data.capabilities?.enableGPU ?? false,
            oomScoreAdj: data.capabilities?.oomScoreAdj,
            sysctls: Object.entries(data.capabilities?.sysctls ?? {}).map(([name, value]) => ({ name, value })),
        },
    };
}
