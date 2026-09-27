export const EAppScheduledJobScheduleMode = {
    /** Run by hand, or as a step of a job sequence. */
    None: "none",
    Interval: "interval",
    Cron: "cron",
} as const;

export type EAppScheduledJobScheduleMode =
    (typeof EAppScheduledJobScheduleMode)[keyof typeof EAppScheduledJobScheduleMode];
