import type { HivePaaSLoggingPerformance_UpdateOnePayload } from "~/system-settings/api/services";
import type { LoggingPerformance } from "~/system-settings/domain";

import type { LoggingPerformanceFormInput, LoggingPerformanceFormOutput } from "../schemas";

export function toLoggingPerformanceFormInput(data?: LoggingPerformance): LoggingPerformanceFormInput {
    return {
        enabled: data?.enabled ?? false,
        nodes: (data?.nodes ?? []).map(node => ({ id: node.id, enabled: node.enabled, capacity: node.capacity })),
    };
}

/** The nodes that run OBI are those enabled; the others are left out. */
export function toLoggingPerformancePayload(
    values: LoggingPerformanceFormOutput,
    updateVer: number,
): HivePaaSLoggingPerformance_UpdateOnePayload {
    return {
        updateVer,
        enabled: values.enabled,
        nodes: values.nodes.filter(node => node.enabled).map(node => ({ id: node.id, capacity: node.capacity })),
    };
}
