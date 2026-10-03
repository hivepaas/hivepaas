import type { LoggingPerformanceCapacity, LoggingPerformanceCapacityInfo } from "~/system-settings/domain";

export const CAPACITY_LABEL: Record<LoggingPerformanceCapacity, string> = {
    small: "Small",
    medium: "Medium",
    large: "Large",
};

/** A capacity as an agent wrote it, as a person reads it. */
export function capacityLabel(capacity: string): string {
    return capacity in CAPACITY_LABEL ? CAPACITY_LABEL[capacity as LoggingPerformanceCapacity] : capacity;
}

/** A capacity as a choice reads: its memory and how many it tracks at once. */
export function capacityText(info: LoggingPerformanceCapacityInfo | undefined): string {
    if (!info) {
        return "";
    }
    return `about ${info.memoryMiB} MiB, ${info.tracked.toLocaleString()} at once`;
}

/** A node's memory in GiB, as its owner bought it, give or take what the kernel keeps. */
export function formatNodeMemory(bytes: number): string {
    if (bytes <= 0) {
        return "-";
    }
    return `${(bytes / 1024 ** 3).toFixed(1)} GiB`;
}
