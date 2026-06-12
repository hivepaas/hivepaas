type ScheduleModeLike = {
    readonly Interval: string;
    readonly Cron: string;
};

export function getScheduleModeFromCronExpr<TScheduleMode extends ScheduleModeLike>(
    cronExpr: string | null | undefined,
    scheduleMode: TScheduleMode,
): TScheduleMode["Interval"] | TScheduleMode["Cron"] {
    return (cronExpr ?? "").trim().length > 0 ? scheduleMode.Cron : scheduleMode.Interval;
}
