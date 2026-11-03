import { z } from "zod";
import type { AppScheduledJobs_Trigger_Payload } from "~/projects/api/services";
import type { AppScheduledJobTrigger } from "~/projects/domain";
import { ESchedJobTriggerEvent } from "~/projects/module-shared/enums";

/** The most triggers a job has, as the server allows. */
export const JOB_MAX_TRIGGERS = 10;

export const JobTriggerFormSchema = z.object({
    event: z.nativeEnum(ESchedJobTriggerEvent),
    /** The apps whose events run the job: an env's job only. */
    appIds: z.array(z.string()),
    wait: z.boolean(),
});

export const JobTriggersFormSchema = z
    .array(JobTriggerFormSchema)
    .max(JOB_MAX_TRIGGERS, `A job has at most ${JOB_MAX_TRIGGERS} triggers`);

export type JobTriggerFormValue = z.input<typeof JobTriggerFormSchema>;

export function createDefaultJobTrigger(): JobTriggerFormValue {
    return { event: ESchedJobTriggerEvent.PostDeploy, appIds: [], wait: false };
}

export function mapJobTriggersToFormValues(triggers: AppScheduledJobTrigger[]): JobTriggerFormValue[] {
    return triggers.map(trigger => ({
        event: trigger.event,
        appIds: trigger.apps.map(app => app.id),
        wait: trigger.wait,
    }));
}

/** What the server takes: apps only when named, wait only for pre-deploy. */
export function mapJobTriggersFormValuesToPayload(values: JobTriggerFormValue[]): AppScheduledJobs_Trigger_Payload[] {
    return values.map(value => ({
        event: value.event,
        ...(value.appIds.length > 0 ? { apps: value.appIds.map(id => ({ id })) } : {}),
        ...(value.event === ESchedJobTriggerEvent.PreDeploy && value.wait ? { wait: true } : {}),
    }));
}

/** How a list names a trigger: its event, and that the deploy waits for it. */
export function formatJobTrigger(trigger: AppScheduledJobTrigger): string {
    return trigger.wait ? `${trigger.event} · waits` : trigger.event;
}
