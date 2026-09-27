import { z } from "zod";
import { JobTriggersFormSchema } from "~/projects/module-shared/components/job-triggers-field/job-triggers.helpers";
import {
    EAppScheduledJobScheduleMode,
    EAppScheduledJobTaskPriority,
    ESchedJobSeqMode,
    ESchedJobSeqOnFailure,
} from "~/projects/module-shared/enums";

/** The most steps a sequence has, as the server allows. */
export const JOB_SEQUENCE_MAX_STEPS = 50;

const NotificationRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

const JobSequenceStepSchema = z.object({
    jobId: z.string().min(1, "Pick a job"),
    /** How the picker named the job, shown in the step's row. */
    jobName: z.string(),
    appName: z.string(),
    /** An optional label for the step, shown in the run. */
    name: z.string().trim().max(100, "At most 100 characters"),
});

export const JobSequenceFormSchema = z.object({
    name: z.string().trim().min(1, "Name is required"),
    scheduleMode: z.nativeEnum(EAppScheduledJobScheduleMode),
    scheduleInterval: z.string().trim(),
    scheduleCronExpr: z.string().trim(),
    scheduleFrom: z.date().nullable(),
    scheduleTo: z.date().nullable(),
    steps: z
        .array(JobSequenceStepSchema)
        .min(1, "Add at least one step")
        .max(JOB_SEQUENCE_MAX_STEPS, `A sequence has at most ${JOB_SEQUENCE_MAX_STEPS} steps`),
    mode: z.nativeEnum(ESchedJobSeqMode),
    onFailure: z.nativeEnum(ESchedJobSeqOnFailure),
    timeout: z.string().trim(),
    priority: z.nativeEnum(EAppScheduledJobTaskPriority),
    controlEnabled: z.boolean(),
    triggers: JobTriggersFormSchema,
    notification: z.object({
        successUseDefault: z.boolean(),
        success: NotificationRefSchema.optional(),
        failureUseDefault: z.boolean(),
        failure: NotificationRefSchema.optional(),
    }),
});

export type JobSequenceFormInput = z.input<typeof JobSequenceFormSchema>;
export type JobSequenceFormOutput = z.output<typeof JobSequenceFormSchema>;
export type JobSequenceStepFormValue = JobSequenceFormInput["steps"][number];
