export const EAppScheduledJobTaskPriority = {
    Low: "low",
    Default: "default",
    Critical: "critical",
} as const;

export type EAppScheduledJobTaskPriority =
    (typeof EAppScheduledJobTaskPriority)[keyof typeof EAppScheduledJobTaskPriority];
