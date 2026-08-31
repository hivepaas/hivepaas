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
