import type { AppScheduledJobs_Schedule_Payload } from "~/projects/api/services";
import type { AppScheduledJobSchedule } from "~/projects/domain";
import { EAppScheduledJobScheduleMode } from "~/projects/module-shared/enums";

/** The schedule fields a job form holds, under these names. */
export interface JobScheduleFormValues {
    scheduleMode: EAppScheduledJobScheduleMode;
    scheduleInterval: string;
    scheduleCronExpr: string;
    scheduleFrom: Date | null;
    scheduleTo: Date | null;
}

export function createDefaultJobScheduleFormValues(
    mode: EAppScheduledJobScheduleMode = EAppScheduledJobScheduleMode.Interval,
): JobScheduleFormValues {
    return {
        scheduleMode: mode,
        scheduleInterval: "",
        scheduleCronExpr: "",
        scheduleFrom: null,
        scheduleTo: null,
    };
}

export function mapJobScheduleToFormValues(schedule: AppScheduledJobSchedule | null): JobScheduleFormValues {
    if (!schedule) {
        return createDefaultJobScheduleFormValues(EAppScheduledJobScheduleMode.None);
    }

    const hasInterval = schedule.interval.trim().length > 0;

    return {
        scheduleMode: hasInterval ? EAppScheduledJobScheduleMode.Interval : EAppScheduledJobScheduleMode.Cron,
        scheduleInterval: schedule.interval,
        scheduleCronExpr: schedule.cronExpr,
        scheduleFrom: schedule.initialTime,
        scheduleTo: schedule.endTime,
    };
}

/** The schedule to send: null for No schedule. */
export function mapJobScheduleFormValuesToPayload(
    values: JobScheduleFormValues,
): AppScheduledJobs_Schedule_Payload | null {
    if (values.scheduleMode === EAppScheduledJobScheduleMode.None) {
        return null;
    }

    const interval = values.scheduleInterval.trim();
    const cronExpr = values.scheduleCronExpr.trim();

    return {
        ...(values.scheduleMode === EAppScheduledJobScheduleMode.Interval && interval ? { interval } : {}),
        ...(values.scheduleMode === EAppScheduledJobScheduleMode.Cron && cronExpr ? { cronExpr } : {}),
        ...(values.scheduleFrom ? { initialTime: values.scheduleFrom } : {}),
        ...(values.scheduleTo ? { endTime: values.scheduleTo } : {}),
    };
}

/** How a list shows a job's schedule. */
export function formatJobSchedule(schedule: AppScheduledJobSchedule | null): string {
    if (schedule?.interval) {
        return `every ${schedule.interval}`;
    }

    if (schedule?.cronExpr) {
        return `cron: ${schedule.cronExpr}`;
    }

    return "No schedule";
}
