export const EAppScheduledJobTaskStatus = {
    NotStarted: "not-started",
    InProgress: "in-progress",
    Canceled: "canceled",
    Done: "done",
    Failed: "failed",
} as const;

export type EAppScheduledJobTaskStatus = (typeof EAppScheduledJobTaskStatus)[keyof typeof EAppScheduledJobTaskStatus];
