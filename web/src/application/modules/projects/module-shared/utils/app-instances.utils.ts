import type { AppServiceTask } from "~/projects/domain";

/** Polling interval shared by every view that watches app service tasks. */
export const APP_SERVICE_TASKS_REFETCH_INTERVAL_MS = 5_000;

const RUNNING_STATE = "running";

export type AppInstancesHealth = {
    /** Instances reported as running among the desired ones. */
    current: number;
    /** Instances the swarm is expected to keep running. */
    total: number;
};

function isRunning(state: string | null | undefined): boolean {
    return (state ?? "").trim().toLowerCase() === RUNNING_STATE;
}

/**
 * Desired instances are the tasks swarm wants running; current ones are those actually running.
 * Tasks with any other desired state (shutdown, remove, ...) are leftovers and must be ignored.
 */
export function computeAppInstancesHealth(tasks: AppServiceTask[]): AppInstancesHealth {
    const desiredTasks = tasks.filter(task => isRunning(task.desiredState));

    return {
        current: desiredTasks.filter(task => isRunning(task.status?.state)).length,
        total: desiredTasks.length,
    };
}

export type AppReplicasStatus = "healthy" | "degraded" | "down" | "over";

/**
 * Classifies a running/desired replica pair.
 *
 * `over` matters: more tasks running than desired means the service is converging down or holds
 * stale tasks. Folding it into `healthy` would hide a real anomaly behind a green light.
 */
export function resolveAppReplicasStatus(running: number, desired: number): AppReplicasStatus {
    if (running > desired) {
        return "over";
    }

    // Equal counts, including 0/0: nothing is expected, so nothing is wrong.
    if (running === desired) {
        return "healthy";
    }

    return running > 0 ? "degraded" : "down";
}

export const APP_REPLICAS_STATUS_DOT_CLASS: Record<AppReplicasStatus, string> = {
    healthy: "bg-green-500",
    degraded: "bg-orange-500",
    down: "bg-red-500",
    over: "bg-blue-500",
};

export const APP_REPLICAS_STATUS_BADGE_CLASS: Record<AppReplicasStatus, string> = {
    healthy: "bg-green-500 text-white",
    degraded: "bg-orange-500 text-white",
    down: "bg-red-500 text-white",
    over: "bg-blue-500 text-white",
};

export function getAppReplicasStatusLabel(running: number, desired: number): string {
    switch (resolveAppReplicasStatus(running, desired)) {
        case "over":
            return `${running} of ${desired} desired replicas running - more tasks than desired, the service may be scaling down or holding stale tasks`;
        case "degraded":
            return `${running} of ${desired} desired replicas running`;
        case "down":
            return `No replica running out of ${desired} desired`;
        default:
            return `${running} of ${desired} desired replicas running`;
    }
}
