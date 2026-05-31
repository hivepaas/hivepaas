export const EAppScheduledJobScheduleMode = {
    Interval: "interval",
    Cron: "cron",
} as const;

export type EAppScheduledJobScheduleMode =
    (typeof EAppScheduledJobScheduleMode)[keyof typeof EAppScheduledJobScheduleMode];
