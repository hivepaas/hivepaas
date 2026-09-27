import type { AppScheduledJobs_Upsert_Payload } from "~/projects/api/services";
import type { AppScheduledJob } from "~/projects/domain";
import {
    EAppScheduledJobScheduleMode,
    EAppScheduledJobTaskPriority,
    EAppScheduledJobType,
    ESchedJobSeqMode,
    ESchedJobSeqOnFailure,
} from "~/projects/module-shared/enums";

import { ESettingStatus } from "@application/shared/enums";

import {
    createDefaultJobScheduleFormValues,
    mapJobScheduleFormValuesToPayload,
    mapJobScheduleToFormValues,
} from "../job-schedule-fields";

import type { JobSequenceFormInput, JobSequenceFormOutput } from "./job-sequence.form.schema";

/** A new sequence runs by hand until it is given a schedule, and can be canceled. */
export function createEmptyJobSequenceFormDefaults(): JobSequenceFormInput {
    return {
        name: "",
        ...createDefaultJobScheduleFormValues(EAppScheduledJobScheduleMode.None),
        steps: [],
        mode: ESchedJobSeqMode.Sequential,
        onFailure: ESchedJobSeqOnFailure.Stop,
        timeout: "",
        priority: EAppScheduledJobTaskPriority.Default,
        controlEnabled: true,
        notification: {
            successUseDefault: true,
            success: undefined,
            failureUseDefault: true,
            failure: undefined,
        },
    };
}

export function mapJobSequenceToFormInput(job: AppScheduledJob): JobSequenceFormInput {
    const { sequence } = job;

    return {
        name: job.name,
        ...mapJobScheduleToFormValues(job.schedule),
        steps: (sequence?.steps ?? []).map(step => ({
            jobId: step.job.id,
            // A job deleted since is named so; the server skips its step.
            jobName: step.job.status === ESettingStatus.Missing ? "Deleted job" : step.job.name,
            appName: step.app?.name ?? "",
            name: step.name,
        })),
        mode: sequence?.mode ?? ESchedJobSeqMode.Sequential,
        onFailure: sequence?.onFailure ?? ESchedJobSeqOnFailure.Stop,
        timeout: job.timeout,
        priority: job.priority,
        controlEnabled: !job.controlDisabled,
        notification: {
            successUseDefault: job.notification?.successUseDefault ?? true,
            success: job.notification?.success,
            failureUseDefault: job.notification?.failureUseDefault ?? true,
            failure: job.notification?.failure,
        },
    };
}

/** A sequence has no app and no command: each step's job has its own. */
export function mapJobSequenceFormToPayload(values: JobSequenceFormOutput): AppScheduledJobs_Upsert_Payload {
    return {
        inheritable: false,
        default: false,
        name: values.name,
        jobType: EAppScheduledJobType.JobSequence,
        schedule: mapJobScheduleFormValuesToPayload(values),
        priority: values.priority,
        maxRetry: 0,
        retryBackoff: false,
        ...(values.timeout ? { timeout: values.timeout } : {}),
        controlDisabled: !values.controlEnabled,
        sequence: {
            mode: values.mode,
            onFailure: values.onFailure,
            steps: values.steps.map(step => ({ job: { id: step.jobId }, name: step.name })),
        },
        notification: {
            successUseDefault: values.notification.successUseDefault,
            ...(!values.notification.successUseDefault && values.notification.success
                ? { success: values.notification.success }
                : {}),
            failureUseDefault: values.notification.failureUseDefault,
            ...(!values.notification.failureUseDefault && values.notification.failure
                ? { failure: values.notification.failure }
                : {}),
        },
    };
}
