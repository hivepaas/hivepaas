export const EAppDeploymentStatus = {
    NotStarted: "not-started",
    InProgress: "in-progress",
    Canceled: "canceled",
    Done: "done",
    Failed: "failed",
} as const;

export type EAppDeploymentStatus = (typeof EAppDeploymentStatus)[keyof typeof EAppDeploymentStatus];
